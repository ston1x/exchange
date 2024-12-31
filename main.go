package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Replace with your actual module path
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"
	"github.com/ston1x/exchange/binance"
	"github.com/ston1x/exchange/config"
	"github.com/ston1x/exchange/pkg/logger"
)

type QuoteHandler struct {
	binanceService *binance.Service
}

func main() {
	// Log
	log := logger.New("main")
	log.Info("Exchange Quotes API Project")

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})
	// Test Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Info("Redis connected")

	// Binance
	symbolPairs := config.GetSymbolPairs()
	binanceService := binance.NewService(rdb, symbolPairs)

	// Cron
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal("[cron-error] Failed to create scheduler: %v", err)
	}
	job, err := scheduler.NewJob(
		gocron.DurationJob(time.Minute),
		gocron.NewTask(func() {
			ctx := context.Background()
			binanceService.UpdateAllPrices(ctx)
		}),
	)
	if err != nil {
		log.Fatal("[cron-error] Failed to create job: %v", err)
	}
	log.Printf("[cron-info] Job ID %v", job.ID())
	scheduler.Start()

	// Set up HTTP router
	router := setupRouter(binanceService)

	// Chunk of generated handy code below
	// Configure HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Info("Server is ready to handle requests at :8080")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutdown signal received")

	// Give outstanding requests a chance to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// when you're done, shut it down
	if err := scheduler.Shutdown(); err != nil {
		log.Fatalf("Failed to shutdown scheduler: %v", err)
	}
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}

func setupRouter(binanceService *binance.Service) *gin.Engine {
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Metrics endpoint
	router.GET("/metrics", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, binanceService.GetMetrics())
	})

	router.GET("/quote", func(ctx *gin.Context) {
		symbol1 := ctx.Query("symbol1")
		symbol2 := ctx.Query("symbol2")

		if symbol1 == "" || symbol2 == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "both symbol1 and symbol2 are required",
			})
			return
		}

		pair := config.SymbolPair{
			Base:  symbol1,
			Quote: symbol2,
		}

		quote, err := binanceService.GetQuote(ctx.Request.Context(), pair)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, quote)
	})
	// router.GET("/quotes", handler.getQuotes) // TODO: To be implemented
	return router
}

// TODO: To be implemented
// func allQuotes(c *gin.Context) {
// 	quote := Quote{"BTCUSDT", 100000}
// 	quotes := []Quote{quote, quote, quote}
// 	c.JSON(http.StatusOK, quotes)
// }

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
