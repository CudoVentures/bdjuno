package cw20token

import (
	"encoding/json"
	"testing"

	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
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
)

var (
	logo1 = json.RawMessage(`{"url":"url"}`)
	logo2 = json.RawMessage(`{"newUrl":"newUrl"}`)
)

var mockTokenInfo = types.TokenInfo{
	Address:     tokenAddr1,
	Name:        str1,
	Symbol:      str1,
	Decimals:    num1,
	TotalSupply: fund * 2,
	Mint:        types.Mint{addr1, fund * 10},
	Marketing:   types.Marketing{str1, str1, addr1, logo1},
	CodeID:      num1,
	Balances:    []types.TokenBalance{{addr1, fund}, {addr2, fund}},
}

func mockMsgExecute(t *testing.T, msg types.MsgExecute) *wasm.MsgExecuteContract {
	msgJson, err := json.Marshal(msg)
	require.NoError(t, err)

	return &wasm.MsgExecuteContract{
		Contract: tokenAddr1,
		Sender:   addr1,
		Msg:      msgJson,
	}
}

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
