package remote

import (
	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/forbole/juno/v2/node/remote"

	"github.com/forbole/bdjuno/v2/modules/cw20token/source"
	q "github.com/forbole/bdjuno/v2/modules/cw20token/source/queryhandler"
	"github.com/forbole/bdjuno/v2/types"
)

var (
	_ source.Source = &Source{}
)

type Source struct {
	*remote.Source
	q *q.QueryHandler
}

func NewSource(source *remote.Source, querier wasm.QueryClient) *Source {
	return &Source{
		Source: source,
		q:      &q.QueryHandler{querier.SmartContractState},
	}
}
func (s *Source) TokenInfo(tokenAddr string, height int64) (types.TokenInfo, error) {
	ctx := remote.GetHeightRequestContext(s.Ctx, height)
	return s.q.TokenInfo(ctx, tokenAddr, height)
}

func (s *Source) AllBalances(tokenAddr string, height int64) ([]types.TokenBalance, error) {
	ctx := remote.GetHeightRequestContext(s.Ctx, height)
	return s.q.AllBalances(ctx, tokenAddr, height)
}

func (s *Source) Balance(tokenAddr string, address string, height int64) (uint64, error) {
	ctx := remote.GetHeightRequestContext(s.Ctx, height)
	return s.q.Balance(ctx, tokenAddr, address, height)
}

func (s *Source) TotalSupply(tokenAddr string, height int64) (string, error) {
	ctx := remote.GetHeightRequestContext(s.Ctx, height)
	return s.q.TotalSupply(ctx, tokenAddr, height)
}
