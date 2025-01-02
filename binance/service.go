package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/ston1x/exchange/config"
	"github.com/ston1x/exchange/pkg/logger"
)

// Service handles all Binance API interactions
// // as well as caching quotes in Redis
type Service struct {
	client      *http.Client
	redis       *goredis.Client
	cacheTTL    time.Duration
	symbolPairs []config.SymbolPair
	log         *logger.Logger
	metrics     *Metrics
}

// Metrics holds basic counters for monitoring
type Metrics struct {
	cacheReads  int64
	cacheWrites int64
	cacheMisses int64
	apiCalls    int64
	apiErrors   int64
}

// BinancePrice represents the response from Binance API
type BinancePrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// NewService creates a new Binance service with configured HTTP client
func NewService(redisClient *goredis.Client, symbolPairs []config.SymbolPair) *Service {
	return &Service{
		client: &http.Client{
			Timeout: 10 * time.Second, // Adding timeout for safety
		},
		redis:       redisClient,
		cacheTTL:    time.Minute,
		symbolPairs: symbolPairs,
		log:         logger.New("binance"),
		metrics:     &Metrics{},
	}
}

// GetQuote fetches the current price for a given symbol pair
func (s *Service) GetQuote(ctx context.Context, symbolPair config.SymbolPair) (*config.Quote, error) {
	cacheKey := s.getCacheKey(symbolPair)

	// Using the passed context for Redis operations. NOTE: Why is ctx needed?
	cachedPrice, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache read
		s.metrics.cacheReads++
		// Convert string price to float64
		floatPrice, convErr := strconv.ParseFloat(cachedPrice, 64)
		if convErr != nil {
			return &config.Quote{}, fmt.Errorf("failed to parse price value: %w", err)
		}
		return &config.Quote{SymbolPair: symbolPair, Price: floatPrice}, nil
	}
	// Cache miss
	s.metrics.cacheMisses++

	quote, err := s.fetchPriceFromAPI(symbolPair)
	if err != nil {
		// API Error
		s.metrics.apiErrors++
		return &config.Quote{}, fmt.Errorf("failed to fetch price from API: %w", err)
	}

	if err := s.cachePrice(ctx, cacheKey, quote); err != nil {
		s.log.Error("Failed to cache price: %v", err)
	}

	return quote, nil
}

func (s *Service) getCacheKey(symbolPair config.SymbolPair) string {
	return fmt.Sprintf("price:%s:%s", symbolPair.Base, symbolPair.Quote)
}

func (s *Service) cachePrice(ctx context.Context, key string, quote *config.Quote) error {
	// Convert price to string
	priceStr := strconv.FormatFloat(quote.Price, 'f', -1, 64)
	// Set the key in Redis
	err := s.redis.Set(ctx, key, priceStr, s.cacheTTL).Err()
	s.metrics.cacheWrites++

	return err
}

func (s *Service) UpdateAllPrices(ctx context.Context) {
	var errors []string // Collect all errors for comprehensive logging

	// TODO: Run this in threads?
	for _, pair := range s.symbolPairs {
		quote, err := s.fetchPriceFromAPI(pair)
		if err != nil {
			// Log error but continue with other pairs
			s.log.Printf("Failed to fetch price for %s: %v", pair.Symbol(), err)
			errors = append(errors, fmt.Sprintf("fetch %s: %v", pair.Symbol(), err))
			continue
		}

		cacheKey := s.getCacheKey(pair)
		if err := s.cachePrice(ctx, cacheKey, quote); err != nil {
			s.log.Printf("Failed to cache price for %s (key: %s): %v", pair.Symbol(), cacheKey, err)
			errors = append(errors, fmt.Sprintf("cache %s: %v", pair.Symbol(), err))
		}

		if len(errors) > 0 {
			s.log.Printf("Update cycle completed with %d errors: %v", len(errors), strings.Join(errors, "; "))
		} else {
			s.log.Printf("Update cycle completed successfully for all %d pairs", len(s.symbolPairs))
		}
	}
}

func (s *Service) fetchPriceFromAPI(symbolPair config.SymbolPair) (*config.Quote, error) {
	s.log.Info("Fetching price for symbol: %s" + symbolPair.Symbol())
	// Construct the URL with the symbol
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", symbolPair.Symbol())

	// Make the HTTP request
	resp, err := s.client.Get(url)
	s.metrics.apiCalls++
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the status code is not 200
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	// Decode the response
	var price BinancePrice
	if err := json.NewDecoder(resp.Body).Decode(&price); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	// Convert string price to float64
	floatPrice, err := strconv.ParseFloat(price.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price value: %w", err)
	}
	quote := config.Quote{
		SymbolPair: symbolPair,
		Price:      floatPrice,
	}
	return &quote, nil
}

// GetMetrics returns current metrics - useful for monitoring endpoint
func (s *Service) GetMetrics() map[string]int64 {
	return map[string]int64{
		"cache_reads":  s.metrics.cacheReads,
		"cache_writes": s.metrics.cacheWrites,
		"cache_misses": s.metrics.cacheMisses,
		"api_calls":    s.metrics.apiCalls,
		"api_errors":   s.metrics.apiErrors,
	}
}
