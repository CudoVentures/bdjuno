package local

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/forbole/bdjuno/v2/modules/cw20token/source"
	cw20tokenkeeper "github.com/forbole/bdjuno/v2/modules/cw20token/source"
	"github.com/forbole/bdjuno/v2/types"
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

func (s Source) GetTokenInfo(contract string, height int64) (*types.TokenInfo, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return nil, err
	}

	res := &wasmtypes.QueryAllContractStateResponse{}
	for {
		req := &wasmtypes.QueryAllContractStateRequest{
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

	return source.ParseToTokenInfo(res)
}

func (s Source) GetBalance(contract string, address string, height int64) (uint64, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return 0, err
	}

	query := fmt.Sprintf(`"balance":{"address":"%s"}`, address)
	req := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contract,
		QueryData: []byte(query),
	}

	res, err := s.wasmClient.SmartContractState(ctx, req)
	if err != nil {
		return 0, err
	}

	return source.ParseToBalance(res)
}

func (s Source) GetCirculatingSupply(contract string, height int64) (uint64, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return 0, err
	}

	query := `{"token_info":{}}`
	req := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contract,
		QueryData: []byte(query),
	}

	res, err := s.wasmClient.SmartContractState(ctx, req)
	if err != nil {
		return 0, err
	}

	return source.ParseToTotalSupply(res)
}
