package queryhandler

import (
	"context"
	"encoding/json"
	"fmt"

	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
	"google.golang.org/grpc"

	"github.com/forbole/bdjuno/v2/types"
)

type queryFn func(ctx context.Context, in *wasm.QuerySmartContractStateRequest, opts ...grpc.CallOption) (*wasm.QuerySmartContractStateResponse, error)

type queryFnLocal func(context.Context, *wasm.QuerySmartContractStateRequest) (*wasm.QuerySmartContractStateResponse, error)

type QueryHandler struct {
	Query queryFn
}

func FromLocal(q queryFnLocal) *QueryHandler {
	queryFn := func(ctx context.Context, in *wasm.QuerySmartContractStateRequest, opts ...grpc.CallOption) (*wasm.QuerySmartContractStateResponse, error) {
		return q(ctx, in)
	}

	return &QueryHandler{queryFn}
}

func (q *QueryHandler) TokenInfo(ctx context.Context, tokenAddr string, height int64) (types.TokenInfo, error) {
	tokenInfo := types.TokenInfo{}

	if err := q.query(ctx, tokenAddr, `{"token_info":{}}`, &tokenInfo); err != nil {
		return types.TokenInfo{}, err
	}

	q.query(ctx, tokenAddr, `{"minter":{}}`, &tokenInfo.Mint)

	q.query(ctx, tokenAddr, `{"marketing_info":{}}`, &tokenInfo.Marketing)

	tokenInfo.Address = tokenAddr
	return tokenInfo, nil
}

func (q *QueryHandler) AllBalances(ctx context.Context, tokenAddr string, height int64) ([]types.TokenBalance, error) {
	balances := []types.TokenBalance{}

	for {
		query := `{"all_accounts":{"limit":30}}`
		if len(balances) > 0 {
			query = fmt.Sprintf(`{"all_accounts":{"limit":30,"start_after":"%s"}}`, balances[len(balances)-1].Address)
		}

		accounts := struct {
			Accounts []string `json:"accounts"`
		}{}

		if err := q.query(ctx, tokenAddr, query, &accounts); err != nil {
			return nil, err
		}

		if len(accounts.Accounts) == 0 {
			break
		}

		for _, a := range accounts.Accounts {
			balance, err := q.Balance(ctx, tokenAddr, a, height)
			if err != nil {
				return nil, err
			}

			balances = append(balances, types.TokenBalance{Address: a, Amount: balance})
		}
	}

	return balances, nil
}

func (q *QueryHandler) Balance(ctx context.Context, tokenAddr string, address string, height int64) (uint64, error) {
	balance := struct {
		Balance uint64 `json:"balance,string"`
	}{}

	query := fmt.Sprintf(`{"balance":{"address":"%s"}}`, address)
	err := q.query(ctx, tokenAddr, query, &balance)

	return balance.Balance, err
}

func (q *QueryHandler) TotalSupply(ctx context.Context, tokenAddr string, height int64) (string, error) {
	supply := struct {
		TotalSupply string `json:"total_supply"`
	}{}

	err := q.query(ctx, tokenAddr, `{"token_info":{}}`, &supply)

	return supply.TotalSupply, err
}

func (q *QueryHandler) query(ctx context.Context, tokenAddr string, query string, dest interface{}) error {
	if dest == nil {
		return nil
	}

	req := &wasm.QuerySmartContractStateRequest{
		Address:   tokenAddr,
		QueryData: []byte(query),
	}

	res, err := q.Query(ctx, req)
	if err != nil {
		return err
	}

	return json.Unmarshal(res.Data, dest)
}
