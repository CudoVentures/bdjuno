package source

import (
	"github.com/forbole/bdjuno/v2/types"
)

type Source interface {
	GetTokenInfo(contract string, height int64) (*types.TokenInfo, error)
	GetBalance(contract string, address string, height int64) (uint64, error)
	GetCirculatingSupply(contract string, height int64) (uint64, error)
}
