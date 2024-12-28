package binance

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// BinancePrice represents the response from Binance API
type BinancePrice struct {
    Symbol string `json:"symbol"`
    Price  string `json:"price"`
}

// Service handles all Binance API interactions
type Service struct {
    client *http.Client
}

// NewService creates a new Binance service with configured HTTP client
func NewService() *Service {
    return &Service{
        client: &http.Client{
            Timeout: 10 * time.Second,  // Adding timeout for safety
        },
    }
}
// GetPrice fetches the current price for a given symbol pair
func (s *Service) GetPrice(symbol string) (*BinancePrice, error) {
    mylog := log.New(os.Stdout, "binance:", log.LstdFlags)
    mylog.Printf("Fetching price for symbol: %s", symbol)
    // Construct the URL with the symbol
    url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", symbol)

    // Make the HTTP request
    resp, err := s.client.Get(url)
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

    return &price, nil
}
