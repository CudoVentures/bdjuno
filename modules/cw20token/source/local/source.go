package local

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cw20tokenkeeper "github.com/forbole/bdjuno/v2/modules/cw20token/source"
	"github.com/forbole/juno/v2/node/local"
)

var (
	_ cw20tokenkeeper.Source = &Source{}
)

type Source struct {
	*local.Source
	wasmClient wasmtypes.QueryServer
}

func NewSource(source *local.Source, wasmClient wasmtypes.QueryServer) *Source {
	return &Source{
		Source:     source,
		wasmClient: wasmClient,
	}
}

func (s Source) AllContractState(address string, height int64) (*wasmtypes.QueryAllContractStateResponse, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return nil, err
	}

	return s.wasmClient.AllContractState(sdk.WrapSDKContext(ctx), &wasmtypes.QueryAllContractStateRequest{Address: address})
}
