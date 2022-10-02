package local

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/types/query"
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

func (s Source) AllContractState(address string, height int64) ([]wasmtypes.Model, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return nil, err
	}

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
