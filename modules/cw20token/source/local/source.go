package local

import (
	wasm "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/forbole/bdjuno/v2/modules/cw20token/source"
	q "github.com/forbole/bdjuno/v2/modules/cw20token/source/queryhandler"
	"github.com/forbole/bdjuno/v2/types"
	"github.com/forbole/juno/v2/node/local"
)

var (
	_ source.Source = &Source{}
)

type Source struct {
	*local.Source
	q *q.QueryHandler
}

func NewSource(source *local.Source, querier wasm.QueryServer) *Source {
	return &Source{
		Source: source,
		q:      q.FromLocal(querier.SmartContractState),
	}
}

func (s *Source) TokenInfo(tokenAddr string, height int64) (types.TokenInfo, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return types.TokenInfo{}, err
	}

	return s.q.TokenInfo(ctx, tokenAddr, height)
}

func (s *Source) AllBalances(tokenAddr string, height int64) ([]types.TokenBalance, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return nil, err
	}

	return s.q.AllBalances(ctx, tokenAddr, height)
}

func (s *Source) Balance(tokenAddr string, address string, height int64) (string, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return "0", err
	}

	return s.q.Balance(ctx, tokenAddr, address, height)
}

func (s *Source) TotalSupply(tokenAddr string, height int64) (string, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return "0", err
	}

	return s.q.TotalSupply(ctx, tokenAddr, height)
}

func (s *Source) Allowance(tokenAddr string, owner string, spender string, height int64) (types.Allowance, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return types.Allowance{}, err
	}

	return s.q.Allowance(ctx, tokenAddr, owner, spender, height)
}
