package pricefeed

import "github.com/forbole/bdjuno/v4/types"

type HistoryModule interface {
	UpdatePricesHistory([]types.TokenPrice) error
}
