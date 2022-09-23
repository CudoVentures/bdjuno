package source

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

type Source interface {
	AllContractState(address string, height int64) (*wasmtypes.QueryAllContractStateResponse, error)
}
