package cw20token

import (
	"encoding/json"
	"testing"
	"time"

	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/CudoVentures/cudos-node/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	source "github.com/forbole/bdjuno/v2/modules/cw20token/source/mock"
	"github.com/forbole/bdjuno/v2/types"
	"github.com/forbole/bdjuno/v2/utils"
	"github.com/stretchr/testify/require"
)

func TestCW20Token_HandleMsg(t *testing.T) {
	for testName, tc := range map[string]struct {
		arrange func(s *source.MockSource, txb *utils.MockTxBuilder) sdk.Msg
	}{
		"instantiate": {
			arrange: func(s *source.MockSource, txb *utils.MockTxBuilder) sdk.Msg {
				s.T.Address = tokenAddr2
				txb.WithEventInstantiateContract(tokenAddr2)
				return &wasm.MsgInstantiateContract{CodeID: s.T.CodeID}
			},
		},
		"execute transfer": {
			arrange: func(s *source.MockSource, txb *utils.MockTxBuilder) sdk.Msg {
				s.Transfer(addr1, addr2, fund)
				txb.WithEventWasmAction(string(types.TypeTransfer))
				return mockMsgExecute(t, types.MsgExecute{Transfer: types.MsgTransfer{addr2, fund}})
			},
		},
		"execute transfer_from": {
			arrange: func(s *source.MockSource, txb *utils.MockTxBuilder) sdk.Msg {
				s.Transfer(addr2, addr1, num1)
				txb.WithEventWasmAction(string(types.TypeTransferFrom))
				return mockMsgExecute(t, types.MsgExecute{TransferFrom: types.MsgTransferFrom{addr2, addr1, num1}})
			},
		},
		"execute send": {
			arrange: func(s *source.MockSource, txb *utils.MockTxBuilder) sdk.Msg {
				s.Transfer(addr1, addr2, num1)
				txb.WithEventWasmAction(string(types.TypeSend))
				return mockMsgExecute(t, types.MsgExecute{Send: types.MsgSend{addr2, num1, nil}})
			},
		},
		"execute send_from": {
			arrange: func(s *source.MockSource, txb *utils.MockTxBuilder) sdk.Msg {
				s.Transfer(addr2, addr1, num1)
				txb.WithEventWasmAction(string(types.TypeSendFrom))
				return mockMsgExecute(t, types.MsgExecute{SendFrom: types.MsgSendFrom{addr2, addr1, num1, nil}})
			},
		},
		"execute burn": {
			arrange: func(s *source.MockSource, txb *utils.MockTxBuilder) sdk.Msg {
				s.Burn(addr1, num1)
				txb.WithEventWasmAction(string(types.TypeBurn))
				return mockMsgExecute(t, types.MsgExecute{Burn: types.MsgBurn{num1}})
			},
		},
		"execute burn_from": {
			arrange: func(s *source.MockSource, txb *utils.MockTxBuilder) sdk.Msg {
				s.Burn(addr2, num1)
				txb.WithEventWasmAction(string(types.TypeBurnFrom))
				return mockMsgExecute(t, types.MsgExecute{BurnFrom: types.MsgBurnFrom{addr2, num1}})
			},
		},
		"execute mint": {
			arrange: func(s *source.MockSource, txb *utils.MockTxBuilder) sdk.Msg {
				s.Mint(addr2, num1)
				txb.WithEventWasmAction(string(types.TypeMint))
				return mockMsgExecute(t, types.MsgExecute{Mint: types.MsgMint{addr2, num1}})
			},
		},
		"execute upload_logo": {
			arrange: func(s *source.MockSource, txb *utils.MockTxBuilder) sdk.Msg {
				s.UpdateLogo(string(logo2))
				txb.WithEventWasmAction(string(types.TypeUploadLogo))
				return mockMsgExecute(t, types.MsgExecute{UploadLogo: logo2})
			},
		},
		"execute update_marketing": {
			arrange: func(s *source.MockSource, txb *utils.MockTxBuilder) sdk.Msg {
				s.UpdateMarketing(types.Marketing{str2, str2, addr2, logo1})
				txb.WithEventWasmAction(string(types.TypeUpdateMarketing))
				return mockMsgExecute(t, types.MsgExecute{UpdateMarketing: types.MsgUpdateMarketing{str2, str2, addr2}})
			},
		},
		"execute update_minter": {
			arrange: func(s *source.MockSource, txb *utils.MockTxBuilder) sdk.Msg {
				s.UpdateMinter(addr2)
				txb.WithEventWasmAction(string(types.TypeUpdateMinter))
				return mockMsgExecute(t, types.MsgExecute{UpdateMinter: types.MsgUpdateMinter{addr2}})
			},
		},
		"migrate to invalid codeID": {
			arrange: func(s *source.MockSource, txb *utils.MockTxBuilder) sdk.Msg {
				contractAddr := s.T.Address
				codeID := s.T.CodeID + 1
				s.T = types.TokenInfo{}
				return &wasm.MsgMigrateContract{Contract: contractAddr, CodeID: codeID}
			},
		},
		"migrate": {
			arrange: func(s *source.MockSource, txb *utils.MockTxBuilder) sdk.Msg {
				return &wasm.MsgMigrateContract{Contract: s.T.Address, CodeID: s.T.CodeID}
			},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			db, err := utils.NewTestDb("cw20TokenTest_handleMsg")
			require.NoError(t, err)

			s := source.NewMockSource(mockTokenInfo)

			_, err = db.Sqlx.Exec(`INSERT INTO cw20token_code_id VALUES ($1)`, s.T.CodeID)
			require.NoError(t, err)

			_, err = db.Sqlx.Exec(
				`INSERT INTO cw20token_info VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
				s.T.Address, s.T.CodeID, s.T.Name, s.T.Symbol, s.T.Decimals, s.T.TotalSupply, s.T.TotalSupply, s.T.Mint.MaxSupply,
				s.T.Mint.Minter, s.T.Marketing.Admin, s.T.Marketing.Project, s.T.Marketing.Description, s.T.Marketing.Logo, s.T.Type, s.T.Creator,
			)
			require.NoError(t, err)

			_, err = db.Sqlx.Exec(
				`INSERT INTO cw20token_balance VALUES ($1, $2, $3), ($4, $5, $6)`,
				s.T.Balances[0].Address, s.T.Address, fund, s.T.Balances[1].Address, s.T.Address, "20",
			)
			require.NoError(t, err)

			m := NewModule(simapp.MakeTestEncodingConfig().Marshaler, db, s)
			txb := utils.NewMockTxBuilder(t, time.Time{}, "", num1)
			msg := tc.arrange(s, txb)

			err = m.HandleMsg(0, msg, txb.Build())
			require.NoError(t, err)

			var res []dbtypes.TokenInfoRow
			err = db.Sqlx.Select(&res, `SELECT * FROM cw20token_info WHERE address = $1`, s.T.Address)
			require.NoError(t, err)

			have := types.TokenInfo{Balances: []types.TokenBalance{}}
			if len(res) > 0 {
				have = parseTokenInfoFromDbRow(res[0])
			}

			balances := s.T.Balances
			s.T.Balances = []types.TokenBalance{}

			require.Equal(t, s.T, have)

			for _, b := range balances {
				var have string
				err = db.Sqlx.QueryRow(`SELECT balance FROM cw20token_balance WHERE address = $1 AND token = $2`, b.Address, s.T.Address).Scan(&have)
				require.NoError(t, err)
				require.Equal(t, b.Amount, have)
			}
		})
	}
}

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
	Address:       tokenAddr1,
	Name:          str1,
	Symbol:        str1,
	Decimals:      num1,
	InitialSupply: "20",
	TotalSupply:   "20",
	Mint:          types.Mint{addr1, "200"},
	Marketing:     types.Marketing{str1, str1, addr1, logo1},
	CodeID:        num1,
	Balances:      []types.TokenBalance{{addr1, "20"}, {addr2, "20"}},
}

func mockMsgExecute(t *testing.T, msg types.MsgExecute) *wasm.MsgExecuteContract {
	msgJSON, err := json.Marshal(msg)
	require.NoError(t, err)

	return &wasm.MsgExecuteContract{
		Contract: tokenAddr1,
		Sender:   addr1,
		Msg:      msgJSON,
	}
}

func parseTokenInfoFromDbRow(t dbtypes.TokenInfoRow) types.TokenInfo {
	return types.TokenInfo{
		Address:       t.Address,
		Name:          t.Name,
		Symbol:        t.Symbol,
		Decimals:      t.Decimals,
		InitialSupply: t.InitialSupply,
		TotalSupply:   t.TotalSupply,
		Mint:          types.Mint{t.Minter, t.MaxSupply},
		Marketing:     types.Marketing{t.ProjectURL, t.Description, t.MarketingAdmin, json.RawMessage(t.Logo)},
		CodeID:        t.CodeID,
		Balances:      []types.TokenBalance{}}
}
