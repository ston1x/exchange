package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	// Replace with your actual module path
	"github.com/gin-gonic/gin"
	"github.com/ston1x/exchange/binance"
)

type Quote struct {
	Symbol  string
	Price   float64
}

type QuoteHandler struct {
    binanceService *binance.Service
}

func NewQuoteHandler() *QuoteHandler {
    return &QuoteHandler{
        binanceService: binance.NewService(),
    }
}

func main() {
    mylog := log.New(os.Stdout, "exchange:", log.LstdFlags)
	mylog.Println("Exchange Quotes API Project")

    handler := NewQuoteHandler()

	router := gin.Default()
	router.GET("/quote", handler.getQuote)
    // router.GET("/quotes", handler.getQuotes) // TODO: To be implemented
	router.Run()
}

func (h *QuoteHandler) getQuote(c *gin.Context) {
    // Get symbol from query parameters
    symbol := c.Query("symbol")
    // Get price from Binance
    price, err := h.binanceService.GetPrice(symbol)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": fmt.Sprintf("Failed to get price: %v", err),
        })
        return
    }

    // Convert price string to float64
    priceFloat, err := strconv.ParseFloat(price.Price, 64)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to parse price",
        })
        return
    }

    // Create and return quote
    quote := Quote{
        Symbol: price.Symbol,
        Price:   priceFloat,
    }

    c.JSON(http.StatusOK, quote)
}

func homePage(c *gin.Context) {
	quote := Quote{"BTCUSDT", 100000}
	c.JSON(http.StatusOK, quote)
}

func allQuotes(c *gin.Context) {
	quote := Quote{"BTCUSDT", 100000}
	quotes := []Quote{quote, quote, quote}
	c.JSON(http.StatusOK, quotes)
}
