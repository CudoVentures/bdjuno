package cryptoCompare

// Token contains the information of a single token
type Token struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

// Tokens represents a list of Token objects
type Tokens []Token

// MarketTicker contains the current market data for a single token
type MarketTicker struct {
	Symbol       string  `json:"TOSYMBOL"`
	CurrentPrice float64 `json:"PRICE"`
	MarketCap    float64 `json:"MKTCAP"`
	LastUpdated  int64   `json:"LASTUPDATE"`
}

type TokenRes struct {
	Prices map[string]MarketTicker
}

type PricesRes struct {
	Tokens map[string]TokenRes `json:"Raw"`
}
