package cw20token

import (
	"fmt"
	"testing"
	"time"

	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/CudoVentures/cudos-node/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/forbole/bdjuno/v2/modules/cw20token/source"
	"github.com/forbole/bdjuno/v2/types"
	"github.com/forbole/bdjuno/v2/utils"
	"github.com/stretchr/testify/require"
)

func TestCW20Token_HandleMsg(t *testing.T) {
	for testName, tc := range map[string]struct {
		name    string
		arrange func(s *source.MockSource) sdk.Msg
	}{
		"instantiate": {
			arrange: func(s *source.MockSource) sdk.Msg {
				s.T.Address = tokenAddr2
				return &wasm.MsgInstantiateContract{CodeID: s.T.CodeID}
			},
		},
		"execute transfer": {
			arrange: func(s *source.MockSource) sdk.Msg {
				require.NoError(t, s.Transfer(addr1, addr2, fund))
				return newExecuteMsg(fmt.Sprintf(`{"transfer":{"recipient":"%s","amount":"%d"}}`, addr2, fund))
			},
		},
		"execute transfer_from": {
			arrange: func(s *source.MockSource) sdk.Msg {
				require.NoError(t, s.Transfer(addr2, addr1, num1))
				return newExecuteMsg(fmt.Sprintf(`{"transfer_from":{"owner":"%s","recipient":"%s","amount":"%d"}}`, addr2, addr1, num1))
			},
		},
		"execute send": {
			arrange: func(s *source.MockSource) sdk.Msg {
				require.NoError(t, s.Transfer(addr1, addr2, num1))
				return newExecuteMsg(fmt.Sprintf(`{"send":{"contract":"%s","amount":"%d","msg":{}}}`, addr2, num1))
			},
		},
		"execute send_from": {
			arrange: func(s *source.MockSource) sdk.Msg {
				require.NoError(t, s.Transfer(addr2, addr1, num1))
				return newExecuteMsg(fmt.Sprintf(`{"send_from":{"owner":"%s","contract":"%s","amount":"%d","msg":{}}}`, addr2, addr1, num1))
			},
		},
		"execute burn": {
			arrange: func(s *source.MockSource) sdk.Msg {
				require.NoError(t, s.Burn(addr1, num1))
				return newExecuteMsg(fmt.Sprintf(`{"burn":{"amount":"%d"}}`, num1))
			},
		},
		"execute burn_from": {
			arrange: func(s *source.MockSource) sdk.Msg {
				require.NoError(t, s.Burn(addr2, num1))
				return newExecuteMsg(fmt.Sprintf(`{"burn_from":{"owner":"%s","amount":"%d"}}`, addr2, num1))
			},
		},
		"execute mint": {
			arrange: func(s *source.MockSource) sdk.Msg {
				require.NoError(t, s.Mint(addr2, num1))
				return newExecuteMsg(fmt.Sprintf(`{"mint":{"recipient":"%s","amount":"%d"}}`, addr2, num1))
			},
		},
		"execute upload_logo": {
			arrange: func(s *source.MockSource) sdk.Msg {
				s.UpdateLogo(logo2)
				return newExecuteMsg(fmt.Sprintf(`{"upload_logo":%s}`, logo2))
			},
		},
		"execute update_marketing": {
			arrange: func(s *source.MockSource) sdk.Msg {
				s.UpdateMarketingInfo(*types.NewMarketingInfo(str2, str2, addr2))
				return newExecuteMsg(fmt.Sprintf(`{"update_marketing":{"project":"%s","description":"%s","marketing":"%s"}}`, str2, str2, addr2))
			},
		},
		"execute update_minter": {
			arrange: func(s *source.MockSource) sdk.Msg {
				s.UpdateMinter(addr2)
				return newExecuteMsg(fmt.Sprintf(`{"update_minter":{"new_minter":"%s"}}`, addr2))
			},
		},
		"migrate to invalid codeID": {
			arrange: func(s *source.MockSource) sdk.Msg {
				contractAddr := s.T.Address
				codeID := s.T.CodeID + 1
				s.T = types.TokenInfo{}
				return &wasm.MsgMigrateContract{Contract: contractAddr, CodeID: codeID}
			},
		},
		"migrate": {
			arrange: func(s *source.MockSource) sdk.Msg {
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
				`INSERT INTO cw20token_info VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
				s.T.Address, s.T.CodeID, s.T.Name, s.T.Symbol, s.T.Decimals, s.T.CirculatingSupply, s.T.MintInfo.MaxSupply,
				s.T.MintInfo.Minter, s.T.MarketingInfo.Admin, s.T.MarketingInfo.Project, s.T.MarketingInfo.Description, s.T.Logo,
			)
			require.NoError(t, err)

			_, err = db.Sqlx.Exec(
				`INSERT INTO cw20token_balance VALUES ($1, $2, $3), ($4, $5, $6)`,
				s.T.Balances[0].Address, s.T.Address, fund, s.T.Balances[1].Address, s.T.Address, fund,
			)
			require.NoError(t, err)

			m := NewModule(simapp.MakeTestEncodingConfig().Marshaler, db, s, nil)

			msg := tc.arrange(s)

			tx, err := utils.NewTx(time.Time{}, "", num1).WithEventInstantiateContract(s.T.Address).Build()
			require.NoError(t, err)

			err = m.HandleMsg(0, msg, tx)
			require.NoError(t, err)

			assertTokenInfo(t, db, s.T)
		})
	}
}
