// config/config.go
package config

// SymbolPair represents a trading pair
type SymbolPair struct {
	Base  string // e.g., "ETH"
	Quote string // e.g., "BTC"
}

type Quote struct {
	SymbolPair SymbolPair
	Price      float64
}

func (s SymbolPair) Symbol() string {
	return s.Base + s.Quote
}

// GetSymbolPairs returns the list of pairs we want to track automatically
func GetSymbolPairs() []SymbolPair {
	return []SymbolPair{
		{Base: "BTC", Quote: "USDT"},
		{Base: "ETH", Quote: "USDT"},
		{Base: "ETH", Quote: "BTC"},
		// Add more pairs as needed
	}
}
