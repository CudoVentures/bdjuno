package remote

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/forbole/juno/v2/node/remote"

	"github.com/forbole/bdjuno/v2/modules/cw20token"
	"github.com/forbole/bdjuno/v2/modules/cw20token/source"
	"github.com/forbole/bdjuno/v2/types"
)

var (
	_ source.Source = &Source{}
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

func (s Source) GetTokenInfo(contract string, height int64) (*types.TokenInfo, error) {
	ctx := remote.GetHeightRequestContext(s.Ctx, height)

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

	return cw20token.ParseToTokenInfo(res)
}

func (s Source) GetBalance(contract string, address string, height int64) (uint64, error) {
	ctx := remote.GetHeightRequestContext(s.Ctx, height)

	query := fmt.Sprintf(`"balance":{"address":"%s"}`, address)
	req := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contract,
		QueryData: []byte(query),
	}

	res, err := s.wasmClient.SmartContractState(ctx, req)
	if err != nil {
		return 0, err
	}

	return cw20token.ParseToBalance(res)
}

func (s Source) GetCirculatingSupply(contract string, height int64) (uint64, error) {
	ctx := remote.GetHeightRequestContext(s.Ctx, height)

	query := `{"token_info":{}}`
	req := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contract,
		QueryData: []byte(query),
	}

	res, err := s.wasmClient.SmartContractState(ctx, req)
	if err != nil {
		return 0, err
	}

	return cw20token.ParseToTotalSupply(res)
}
