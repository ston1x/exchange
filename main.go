package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Quote struct {
	Symbol1 string
	Symbol2 string
	Price   float64
}

type QuotesHandler struct {
	store quoteStore
}

type quoteStore interface {
	Get(symbol1, symbol2 string) (Quote, error)
}

func main() {
	router := gin.Default()

	router.GET("/", homePage)

	router.Run()
}

func homePage(c *gin.Context) {
	quote := Quote{"btc", "usd", 100000}
	c.JSON(http.StatusOK, quote)
}
