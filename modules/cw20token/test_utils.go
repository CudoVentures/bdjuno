package cw20token

import (
	"testing"

	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/forbole/bdjuno/v2/database"
	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/types"
	"github.com/stretchr/testify/require"
)

const (
	addr1      = "cudos1"
	addr2      = "cudos2"
	tokenAddr1 = "cudos1cw201"
	tokenAddr2 = "cudos1cw202"
	str1       = "str1"
	str2       = "str2"
	num1       = 1
	fund       = 20
	logo1      = `{"url": "url"}`
	logo2      = `{"newUrl": "newUrl"}`
)

var mockTokenInfo = types.TokenInfo{
	Address:           tokenAddr1,
	Name:              str1,
	Symbol:            str1,
	Decimals:          num1,
	CirculatingSupply: fund * 2,
	MintInfo:          types.MintInfo{Minter: addr1, MaxSupply: fund * 10},
	MarketingInfo:     *types.NewMarketingInfo(str1, str1, addr1),
	Logo:              logo1,
	CodeID:            num1,
	Balances: []types.TokenBalance{
		{Address: addr1, Amount: fund},
		{Address: addr2, Amount: fund},
	},
}

const (
	validExecuteSchema   = `{"$schema":"http://json-schema.org/draft-07/schema#","title":"Cw20ExecuteMsg","oneOf":[{"type":"object","required":["transfer"],"properties":{"transfer":{"type":"object","required":["amount","recipient"],"properties":{"amount":{"$ref":"#/definitions/Uint128"},"recipient":{"type":"string"}}}},"additionalProperties":false},{"type":"object","required":["send"],"properties":{"send":{"type":"object","required":["amount","contract","msg"],"properties":{"amount":{"$ref":"#/definitions/Uint128"},"contract":{"type":"string"},"msg":{"$ref":"#/definitions/Binary"}}}},"additionalProperties":false}],"definitions":{"Uint128":{"type":"string"},"Uint64":{"type":"string"},"Binary":{"type":"string"}}}`
	validQuerySchema     = `{"$schema":"http://json-schema.org/draft-07/schema#","title":"QueryMsg","oneOf":[{"type":"object","required":["balance"],"properties":{"balance":{"type":"object","required":["address"],"properties":{"address":{"type":"string"}}}},"additionalProperties":false},{"type":"object","required":["token_info"],"properties":{"token_info":{"type":"object"}},"additionalProperties":false},{"type":"object","required":["all_accounts"],"properties":{"all_accounts":{"type":"object","properties":{"limit":{"type":["integer","null"],"format":"uint32","minimum":0},"start_after":{"type":["string","null"]}}}},"additionalProperties":false}]}`
	invalidExecuteSchema = `{"$schema":"http://json-schema.org/draft-07/schema#","title":"Cw20ExecuteMsg","oneOf":[{"type":"object","required":["transfer"],"properties":{"transfer":{"type":"object","required":["amount","recipient"],"properties":{"amount":{"$ref":"#/definitions/Uint128"},"recipient":{"type":"string"}}}},"additionalProperties":false}],"definitions":{"Uint128":{"type":"string"}}}`
	invalidQuerySchema   = `{"$schema":"http://json-schema.org/draft-07/schema#","title":"QueryMsg","oneOf":[{"type":"object","required":["balance"],"properties":{"balance":{"type":"object","required":["address"],"properties":{"address":{"type":"string"}}}},"additionalProperties":false}]}`
)

func newPubMsg(codeID uint64, executeSchema string, querySchema string) *types.MsgVerifiedContract {
	return &types.MsgVerifiedContract{
		CodeID:        codeID,
		ExecuteSchema: executeSchema,
		QuerySchema:   querySchema,
	}
}

func parseTokenInfoFromDbRow(t dbtypes.TokenInfoRow) types.TokenInfo {
	return types.TokenInfo{
		Address:           t.Address,
		Name:              t.Name,
		Symbol:            t.Symbol,
		Decimals:          t.Decimals,
		CirculatingSupply: t.CirculatingSupply,
		MintInfo:          types.MintInfo{Minter: t.Minter, MaxSupply: t.MaxSupply},
		MarketingInfo:     types.MarketingInfo{Project: t.ProjectUrl, Description: t.Description, Admin: t.MarketingAdmin},
		Logo:              t.Logo,
		CodeID:            t.CodeID,
		Balances:          []types.TokenBalance{}}
}

func newExecuteMsg(msgJson string) *wasm.MsgExecuteContract {
	return &wasm.MsgExecuteContract{
		Contract: tokenAddr1,
		Sender:   addr1,
		Msg:      []byte(msgJson),
	}
}

func assertTokenInfo(t *testing.T, db *database.Db, want types.TokenInfo) {
	var res []dbtypes.TokenInfoRow
	err := db.Sqlx.Select(&res, `SELECT * FROM cw20token_info WHERE address = $1`, want.Address)
	require.NoError(t, err)

	have := types.TokenInfo{Balances: []types.TokenBalance{}}
	if len(res) > 0 {
		have = parseTokenInfoFromDbRow(res[0])
	}

	balances := want.Balances
	want.Balances = []types.TokenBalance{}

	require.Equal(t, want, have)

	for _, b := range balances {
		var have uint64
		err = db.Sqlx.QueryRow(`SELECT balance FROM cw20token_balance WHERE address = $1 AND token = $2`, b.Address, want.Address).Scan(&have)
		require.NoError(t, err)
		require.Equal(t, b.Amount, have)
	}
}
