package cw20token

import (
	"fmt"
	"testing"
	"time"

	"github.com/CudoVentures/cudos-node/simapp"
	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/modules/cw20token/source/remote"
	"github.com/forbole/bdjuno/v2/utils"
	"github.com/stretchr/testify/require"
)

func TestCW20Token_HandlePeriodicOperations_RemoveExpiredAllowances(t *testing.T) {
	db, err := utils.NewTestDb("cw20TokenTest_removeExpiredAllowances")
	require.NoError(t, err)

	_, err = db.Sqlx.Exec(`INSERT INTO cw20token_code_id VALUES (1)`)
	require.NoError(t, err)

	_, err = db.Sqlx.Exec(
		`INSERT INTO cw20token_info VALUES ('token1', 1, '1', '1', 1, '1', '1', '1', '1', '1', '1', '1', '1', '1', '1')`,
	)
	require.NoError(t, err)

	timestamp := time.Date(2022, time.January, 1, 1, 1, 1, 0, time.FixedZone("", 0))
	allowances := []dbtypes.AllowanceRow{
		{"token1", "addr1", "addr2", "1", `{"never":{}}`},
		{"token1", "addr1", "addr3", "1", `{"at_height":{"height":1}}`},
		{"token1", "addr1", "addr4", "1", `{"at_height":{"height":2}}`},
		{"token1", "addr1", "addr5", "1", fmt.Sprintf(`{"at_time":{"time":"%s"}}`, timestamp.Format(time.RFC3339))},
		{"token1", "addr1", "addr6", "1", fmt.Sprintf(`{"at_time":{"time":"%s"}}`, timestamp.Add(time.Second).Format(time.RFC3339))},
	}
	for _, a := range allowances {
		_, err = db.Sqlx.Exec(
			`INSERT INTO cw20token_allowance VALUES ($1, $2, $3, $4, $5)`,
			a.Token, a.Owner, a.Spender, a.Amount, a.Expires,
		)
		require.NoError(t, err)
	}

	_, err = db.Sql.Exec(`INSERT INTO block (height, hash, timestamp) VALUES ($1, $2, $3)`, 1, "1", timestamp)
	require.NoError(t, err)

	m := NewModule(simapp.MakeTestEncodingConfig().Marshaler, db, &remote.Source{})
	err = m.removeExpiredAllowances()
	require.NoError(t, err)
	haveAllowances := []dbtypes.AllowanceRow{}
	err = db.Sqlx.Select(&haveAllowances, `SELECT * FROM cw20token_allowance`)
	require.NoError(t, err)

	wantAllowances := []dbtypes.AllowanceRow{allowances[0], allowances[2], allowances[4]}
	require.Equal(t, wantAllowances, haveAllowances)
}
