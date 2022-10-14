package cw20token

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/CudoVentures/cudos-node/simapp"
	"github.com/forbole/bdjuno/v2/modules/cw20token/source"
	"github.com/forbole/bdjuno/v2/types"
	"github.com/forbole/bdjuno/v2/utils"
	"github.com/forbole/bdjuno/v2/utils/pubsub"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestCW20Token_HandleAdditionalOperations(t *testing.T) {
	for _, tc := range []struct {
		name              string
		arrange           func(s *source.MockSource) *types.MsgVerifiedContract
		wantAck, wantNack bool
		assertTokenInfo   bool
	}{
		{
			name: "new token codeID",
			arrange: func(s *source.MockSource) *types.MsgVerifiedContract {
				s.T.Address = tokenAddr2
				return newPubMsg(s.T.CodeID, validExecuteSchema, validQuerySchema)
			},
			wantAck:         true,
			assertTokenInfo: true,
		},
		{
			name: "error on m.saveTokenInfo()",
			arrange: func(s *source.MockSource) *types.MsgVerifiedContract {
				n := -1
				s.T.Balances[0].Amount = uint64(n)
				return newPubMsg(s.T.CodeID, validExecuteSchema, validQuerySchema)
			},
			wantNack: true,
		},
		{
			name: "existing token codeID",
			arrange: func(s *source.MockSource) *types.MsgVerifiedContract {
				return newPubMsg(s.T.CodeID+1, validExecuteSchema, validQuerySchema)
			},
			wantAck: true,
		},
		{
			name: "invalid json schema",
			arrange: func(s *source.MockSource) *types.MsgVerifiedContract {
				return newPubMsg(s.T.CodeID, "", "")
			},
			wantAck: true,
		},
		{
			name: "error on dbTx.CodeIDExists()",
			arrange: func(s *source.MockSource) *types.MsgVerifiedContract {
				n := -1
				return newPubMsg(uint64(n), validExecuteSchema, validQuerySchema)
			},
			wantNack: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := utils.NewTestDb("cw20TokenTest_handleAdditionalOperations")
			require.NoError(t, err)

			// ctx, _ := context.WithCancel(context.Background())
			ps, err := pubsub.NewFakeGooglePubSubClient(context.Background())
			require.NoError(t, err)

			s := source.NewMockSource(mockTokenInfo)

			_, err = db.Sqlx.Exec(`INSERT INTO cw20token_code_id VALUES ($1)`, s.T.CodeID+1)
			require.NoError(t, err)

			msg, err := json.Marshal(tc.arrange(s))
			require.NoError(t, err)

			ps.Publish(msg)

			_, err = db.Sql.Exec(`INSERT INTO block (height, hash, timestamp) VALUES ($1, $2, $3)`, num1, str1, time.Now())
			require.NoError(t, err)

			_, err = db.Sql.Exec(
				`INSERT INTO transaction (hash, height, success, signatures) VALUES ($1, $2, true, $3)`,
				str1, num1, pq.Array([]string{str1}),
			)
			require.NoError(t, err)

			_, err = db.Sqlx.Exec(
				`INSERT INTO cosmwasm_instantiate (transaction_hash, index, label, sender, code_id, result_contract_address, success)
				VALUES ($1, $2, $3, $4, $5, $6, true)`, str1, num1, str1, addr1, s.T.CodeID, s.T.Address,
			)
			require.NoError(t, err)

			m := NewModule(simapp.MakeTestEncodingConfig().Marshaler, db, s, ps)
			m.RunAdditionalOperations()
			time.Sleep(time.Second)

			m.mu.Lock()
			defer m.mu.Unlock()
			// cancel()

			require.Equal(t, tc.wantAck, ps.AckCount > 0)
			require.Equal(t, tc.wantNack, ps.NackCount > 0)

			if tc.assertTokenInfo {
				assertTokenInfo(t, db, s.T)
			}
		})
	}
}
