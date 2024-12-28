package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Quote struct {
	Symbol1 string
	Symbol2 string
	Price   float64
}

func main() {
	fmt.Println("Exchange Quotes API Project")

	router := gin.Default()
	router.GET("/quote", homePage)
	router.GET("/quotes", allQuotes)
	router.Run()
}

func homePage(c *gin.Context) {
	quote := Quote{"btc", "usd", 100000}
	c.JSON(http.StatusOK, quote)
}

func allQuotes(c *gin.Context) {
	quote := Quote{"btc", "usd", 100000}
	quotes := []Quote{quote, quote, quote}
	c.JSON(http.StatusOK, quotes)
}
