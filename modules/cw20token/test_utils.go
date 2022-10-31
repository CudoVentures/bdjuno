package cw20token

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/forbole/bdjuno/v2/database"
	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/modules/cw20token/source/fake"
	"github.com/forbole/bdjuno/v2/types"

	"github.com/stretchr/testify/require"
)

var (
	_, _, accAddr1 = testdata.KeyTestPubAddr()
	_, _, accAddr2 = testdata.KeyTestPubAddr()
	addr1          = accAddr1.String()
	addr2          = accAddr2.String()
	logo1          = []byte(`{"url": "url"}`)
	logo2          = []byte(`{"newUrl": "newUrl"}`)
)

var defaultTokenInfo = types.TokenInfo{
	Name:        str1,
	Symbol:      str1,
	Decimals:    num1,
	TotalSupply: fund * 2,
	Mint:        types.Mint{addr1, fund * 10},
	Marketing:   types.NewMarketing(str1, str1, addr1, logo1),
	CodeID:      num1,
	Balances:    []types.TokenBalance{{addr1, fund}, {addr2, fund}},
}

const (
	str1 = "str"
	str2 = "str2"
	num1 = 1
	fund = 20
)

func parseTokenInfoFromDbRow(t dbtypes.TokenInfoRow) types.TokenInfo {
	return types.TokenInfo{
		Address:     t.Address,
		Name:        t.Name,
		Symbol:      t.Symbol,
		Decimals:    t.Decimals,
		TotalSupply: t.TotalSupply,
		Mint:        types.Mint{t.Minter, t.MaxSupply},
		Marketing:   types.Marketing{t.ProjectUrl, t.Description, t.MarketingAdmin, json.RawMessage(t.Logo)},
		CodeID:      t.CodeID,
		Balances:    []types.TokenBalance{}}
}

func assertTokenInfo(t *testing.T, db *database.Db, s *fake.FakeSource) {
	var res []dbtypes.TokenInfoRow
	err := db.Sqlx.Select(&res, `SELECT * FROM cw20token_info WHERE address = $1`, s.TokenAddr)
	require.NoError(t, err)

	have := types.TokenInfo{Balances: []types.TokenBalance{}}
	if len(res) > 0 {
		have = parseTokenInfoFromDbRow(res[0])
	}

	want, err := s.TokenInfo(s.TokenAddr, num1)
	require.NoError(t, err)

	balances := want.Balances
	want.Balances = []types.TokenBalance{}
	want.CodeID = defaultTokenInfo.CodeID

	require.Equal(t, want, have)

	for _, b := range balances {
		var have uint64
		err = db.Sqlx.QueryRow(`SELECT balance FROM cw20token_balance WHERE address = $1 AND token = $2`, b.Address, want.Address).Scan(&have)
		require.NoError(t, err)
		require.Equal(t, b.Amount, have)
	}
}
