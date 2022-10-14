package source

import (
	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
)

type Source interface {
	GetTokenInfo(contract string, height int64) (*wasm.QueryAllContractStateResponse, error)
	GetBalance(contract string, address string, height int64) (*wasm.QuerySmartContractStateResponse, error)
	GetCirculatingSupply(contract string, height int64) (*wasm.QuerySmartContractStateResponse, error)
}
