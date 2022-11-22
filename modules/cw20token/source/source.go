package source

import (
	"github.com/forbole/bdjuno/v2/types"
)

type Source interface {
	TokenInfo(tokenAddr string, height int64) (types.TokenInfo, error)
	AllBalances(tokenAddr string, height int64) ([]types.TokenBalance, error)
	Balance(tokenAddr string, address string, height int64) (string, error)
	TotalSupply(tokenAddr string, height int64) (string, error)
}
