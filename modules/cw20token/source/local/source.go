package local

import (
	"fmt"

	wasm "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/forbole/bdjuno/v2/modules/cw20token/source"
	"github.com/forbole/juno/v2/node/local"
)

var (
	_ source.Source = &Source{}
)

type Source struct {
	*local.Source
	wasmClient wasm.QueryServer
}

func NewSource(source *local.Source, wasmClient wasm.QueryServer) *Source {
	return &Source{
		Source:     source,
		wasmClient: wasmClient,
	}
}

func (s Source) GetTokenInfo(contract string, height int64) (*wasm.QueryAllContractStateResponse, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return nil, err
	}

	res := &wasm.QueryAllContractStateResponse{Pagination: &query.PageResponse{}}
	for {
		req := &wasm.QueryAllContractStateRequest{
			Address:    contract,
			Pagination: &query.PageRequest{Key: res.Pagination.NextKey},
		}

		r, err := s.wasmClient.AllContractState(ctx, req)
		if err != nil {
			return nil, err
		}

		res.Models = append(res.Models, r.Models...)

		if r.Pagination.NextKey == nil {
			break
		}

		res.Pagination.NextKey = r.Pagination.NextKey
	}

	return res, nil
}

func (s Source) GetBalance(contract string, address string, height int64) (*wasm.QuerySmartContractStateResponse, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`{"balance":{"address":"%s"}}`, address)
	req := &wasm.QuerySmartContractStateRequest{
		Address:   contract,
		QueryData: []byte(query),
	}

	res, err := s.wasmClient.SmartContractState(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s Source) GetCirculatingSupply(contract string, height int64) (*wasm.QuerySmartContractStateResponse, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return nil, err
	}

	query := `{"token_info":{}}`
	req := &wasm.QuerySmartContractStateRequest{
		Address:   contract,
		QueryData: []byte(query),
	}

	res, err := s.wasmClient.SmartContractState(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
