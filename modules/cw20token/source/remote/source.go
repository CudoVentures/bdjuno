package remote

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/types/query"
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

func (s Source) AllContractState(address string, height int64) ([]wasmtypes.Model, error) {
	ctx := remote.GetHeightRequestContext(s.Ctx, height)
	state := []wasmtypes.Model{}
	nextPage := []byte{}

	for {
		req := &wasmtypes.QueryAllContractStateRequest{
			Address:    address,
			Pagination: &query.PageRequest{Key: nextPage},
		}

		res, err := s.wasmClient.AllContractState(ctx, req)
		if err != nil {
			return nil, err
		}

		state = append(state, res.Models...)
		nextPage = res.Pagination.NextKey
		if nextPage == nil {
			break
		}
	}

	return state, nil
}
