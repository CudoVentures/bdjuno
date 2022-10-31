package cw20token

import (
	"encoding/json"
	"testing"
	"time"

	wasmapp "github.com/CosmWasm/wasmd/app"
	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	source "github.com/forbole/bdjuno/v2/modules/cw20token/source/fake"
	"github.com/forbole/bdjuno/v2/utils"
	"github.com/stretchr/testify/require"
)

func TestCW20Token_HandleMsg(t *testing.T) {
	for testName, tc := range map[string]struct {
		arrange func(s *source.FakeSource) sdk.Msg
	}{
		"instantiate": {
			arrange: func(s *source.FakeSource) sdk.Msg {
				msg := wasm.MsgInstantiateContract{}
				_, err := s.Instantiate(func(m *wasm.MsgInstantiateContract) {
					m.Sender = addr1
					msgRaw, err := json.Marshal(defaultTokenInfo)
					require.NoError(t, err)
					m.Funds = sdk.Coins{}
					m.Msg = msgRaw
					m.CodeID = defaultTokenInfo.CodeID

					msg = *m
				})
				require.NoError(t, err)
				return &msg
			},
		},
		// "execute transfer": {
		// 	arrange: func(s *source.FakeSource) sdk.Msg {
		// 		require.NoError(t, s.Transfer(addr1, addr2, fund))
		// 		return newExecuteMsg(fmt.Sprintf(`{"transfer":{"recipient":"%s","amount":"%d"}}`, addr2, fund))
		// 	},
		// },
		// "execute transfer_from": {
		// 	arrange: func(s *source.FakeSource) sdk.Msg {
		// 		require.NoError(t, s.Transfer(addr2, addr1, num1))
		// 		return newExecuteMsg(fmt.Sprintf(`{"transfer_from":{"owner":"%s","recipient":"%s","amount":"%d"}}`, addr2, addr1, num1))
		// 	},
		// },
		// "execute send": {
		// 	arrange: func(s *source.FakeSource) sdk.Msg {
		// 		require.NoError(t, s.Transfer(addr1, addr2, num1))
		// 		return newExecuteMsg(fmt.Sprintf(`{"send":{"contract":"%s","amount":"%d","msg":{}}}`, addr2, num1))
		// 	},
		// },
		// "execute send_from": {
		// 	arrange: func(s *source.FakeSource) sdk.Msg {
		// 		require.NoError(t, s.Transfer(addr2, addr1, num1))
		// 		return newExecuteMsg(fmt.Sprintf(`{"send_from":{"owner":"%s","contract":"%s","amount":"%d","msg":{}}}`, addr2, addr1, num1))
		// 	},
		// },
		// "execute burn": {
		// 	arrange: func(s *source.FakeSource) sdk.Msg {
		// 		require.NoError(t, s.Burn(addr1, num1))
		// 		return newExecuteMsg(fmt.Sprintf(`{"burn":{"amount":"%d"}}`, num1))
		// 	},
		// },
		// "execute burn_from": {
		// 	arrange: func(s *source.FakeSource) sdk.Msg {
		// 		require.NoError(t, s.Burn(addr2, num1))
		// 		return newExecuteMsg(fmt.Sprintf(`{"burn_from":{"owner":"%s","amount":"%d"}}`, addr2, num1))
		// 	},
		// },
		// "execute mint": {
		// 	arrange: func(s *source.FakeSource) sdk.Msg {
		// 		require.NoError(t, s.Mint(addr2, num1))
		// 		return newExecuteMsg(fmt.Sprintf(`{"mint":{"recipient":"%s","amount":"%d"}}`, addr2, num1))
		// 	},
		// },
		// "execute upload_logo": {
		// 	arrange: func(s *source.FakeSource) sdk.Msg {
		// 		s.UpdateLogo(logo2)
		// 		return newExecuteMsg(fmt.Sprintf(`{"upload_logo":%s}`, logo2))
		// 	},
		// },
		// "execute update_marketing": {
		// 	arrange: func(s *source.FakeSource) sdk.Msg {
		// 		s.UpdateMarketing(*types.NewMarketing(str2, str2, addr2))
		// 		return newExecuteMsg(fmt.Sprintf(`{"update_marketing":{"project":"%s","description":"%s","marketing":"%s"}}`, str2, str2, addr2))
		// 	},
		// },
		// "execute update_minter": {
		// 	arrange: func(s *source.FakeSource) sdk.Msg {
		// 		s.UpdateMinter(addr2)
		// 		return newExecuteMsg(fmt.Sprintf(`{"update_minter":{"new_minter":"%s"}}`, addr2))
		// 	},
		// },
		// "migrate to invalid codeID": {
		// 	arrange: func(s *source.FakeSource) sdk.Msg {
		// 		contractAddr := tt.Address
		// 		codeID := tt.CodeID + 1
		// 		s.T = types.TokenInfo{}
		// 		return &wasm.MsgMigrateContract{Contract: contractAddr, CodeID: codeID}
		// 	},
		// },
		// "migrate": {
		// 	arrange: func(s *source.FakeSource) sdk.Msg {
		// 		return &wasm.MsgMigrateContract{Contract: tt.Address, CodeID: tt.CodeID}
		// 	},
		// },
	} {
		t.Run(testName, func(t *testing.T) {
			s, err := source.SetupFakeSource(t, defaultTokenInfo)
			require.NoError(t, err)

			db, err := utils.NewTestDb("cw20TokenTest_handleMsg")
			require.NoError(t, err)

			tt := defaultTokenInfo

			_, err = db.Sqlx.Exec(`INSERT INTO cw20token_code_id VALUES ($1)`, tt.CodeID)
			require.NoError(t, err)

			_, err = db.Sqlx.Exec(
				`INSERT INTO cw20token_info VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
				tt.Address, tt.CodeID, tt.Name, tt.Symbol, tt.Decimals, tt.TotalSupply, tt.Mint.MaxSupply,
				tt.Mint.Minter, tt.Marketing.Admin, tt.Marketing.Project, tt.Marketing.Description, tt.Marketing.Logo,
			)
			require.NoError(t, err)

			_, err = db.Sqlx.Exec(
				`INSERT INTO cw20token_balance VALUES ($1, $2, $3), ($4, $5, $6)`,
				tt.Balances[0].Address, tt.Address, fund, tt.Balances[1].Address, tt.Address, fund,
			)
			require.NoError(t, err)

			m := &Module{
				cdc:    wasmapp.MakeEncodingConfig().Marshaler,
				db:     db,
				source: s,
			}

			msg := tc.arrange(s)

			tx, err := utils.NewTx(time.Time{}, "", num1).WithEventInstantiateContract(s.TokenAddr).Build()
			require.NoError(t, err)

			err = m.HandleMsg(0, msg, tx)
			require.NoError(t, err)

			// todo maybe actually use testing.T within fakeSource?? seems like its already a convention
			assertTokenInfo(t, db, s)
		})
	}
}
