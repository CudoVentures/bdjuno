package database

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v2/types"
)

// SaveStakingPool allows to save for the given height the given stakingtypes pool
func (db *Db) SaveStakingPool(pool *types.Pool) error {
	stmt := `
INSERT INTO staking_pool (bonded_tokens, not_bonded_tokens, height) 
VALUES ($1, $2, $3)
ON CONFLICT (one_row_id) DO UPDATE 
    SET bonded_tokens = excluded.bonded_tokens, 
        not_bonded_tokens = excluded.not_bonded_tokens, 
        height = excluded.height
WHERE staking_pool.height <= excluded.height`

	_, err := db.Sql.Exec(stmt, pool.BondedTokens.String(), pool.NotBondedTokens.String(), pool.Height)
	if err != nil {
		return fmt.Errorf("error while storing staking pool: %s", err)
	}

	return nil
}

func (db *Db) GetBondedTokens() (sdk.Int, error) {
	type row struct {
		BondedTokens string `db:"bonded_tokens"`
	}

	var rows []row
	if err := db.Sqlx.Select(&rows, `SELECT bonded_tokens FROM staking_pool`); err != nil {
		return sdk.Int{}, fmt.Errorf("error while getting bonded_tokens: %s", err)
	}

	if len(rows) == 0 {
		return sdk.Int{}, errors.New("failed to find boned_tokens")
	}

	bondedTokens, ok := sdk.NewIntFromString(rows[0].BondedTokens)
	if !ok {
		return sdk.Int{}, fmt.Errorf("invalid bonded_tokens: %s", rows[0].BondedTokens)
	}

	return bondedTokens, nil
}
