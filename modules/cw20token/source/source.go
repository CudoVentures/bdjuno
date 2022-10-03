package source

import (
	"github.com/forbole/bdjuno/v2/types"
)

type Source interface {
	GetTokenInfo(contractAddress string, height int64) (*types.TokenInfo, error)
	GetBalance(contractAddress string, address string, height int64) (uint64, error)
	GetTotalSupply(contractAddress string, height int64) (uint64, error)
}
