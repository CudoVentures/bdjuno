package remote

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/forbole/juno/v2/node/remote"

	cw20tokenkeeper "github.com/forbole/bdjuno/v2/modules/cw20token/source"
)

var (
	_ cw20tokenkeeper.Source = &Source{}
)

type Source struct {
	*remote.Source
	wasmClient wasmtypes.QueryClient
}

func NewSource(source *remote.Source, wasmClient wasmtypes.QueryClient) *Source {

	return &Source{
		Source:     source,
		wasmClient: wasmClient,
	}
}

func (s Source) AllContractState(address string, height int64) (*wasmtypes.QueryAllContractStateResponse, error) {
	ctx := remote.GetHeightRequestContext(s.Ctx, height)
	return s.wasmClient.AllContractState(ctx, &wasmtypes.QueryAllContractStateRequest{Address: address})
}
