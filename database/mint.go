package database

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SaveInflation allows to store the inflation for the given block height as well as timestamp
func (db *Db) SaveInflation(inflation sdk.Dec, height int64) error {
	stmt := `
INSERT INTO inflation (value, height) 
VALUES ($1, $2) 
ON CONFLICT (one_row_id) DO UPDATE 
    SET value = excluded.value, 
        height = excluded.height 
WHERE inflation.height <= excluded.height`

	_, err := db.Sql.Exec(stmt, inflation.String(), height)
	if err != nil {
		return fmt.Errorf("error while storing inflation: %s", err)
	}

	return nil
}

func (db *Db) SaveAPR(apr sdk.Dec, height int64) error {
	stmt := `
INSERT INTO apr (value, height)
VALUES ($1, $2)
ON CONFLICT (one_row_id) DO UPDATE
    SET value = excluded.value,
        height = excluded.height
WHERE apr.height <= excluded.height`

	if _, err := db.Sql.Exec(stmt, apr.String(), height); err != nil {
		return fmt.Errorf("error while storing APR: %s", err)
	}

	return nil
}

func (db *Db) SaveAPRHistory(apr sdk.Dec, height, timestamp int64) error {
	stmt := `INSERT INTO apr_history (value, height, timestamp) VALUES ($1, $2, $3)`

	if _, err := db.Sql.Exec(stmt, apr.String(), height, timestamp); err != nil {
		return fmt.Errorf("error while storing APR history: %s", err)
	}

	return nil
}

func (db *Db) SaveAdjustedSupply(supply sdk.Dec, height int64) error {
	stmt := `
INSERT INTO adjusted_supply (value, height) 
VALUES ($1, $2) 
ON CONFLICT (one_row_id) DO UPDATE 
    SET value = excluded.value, 
        height = excluded.height 
WHERE adjusted_supply.height <= excluded.height`

	_, err := db.Sql.Exec(stmt, supply.String(), height)
	if err != nil {
		return fmt.Errorf("error while storing adjusted supply: %s", err)
	}

	return nil
}
