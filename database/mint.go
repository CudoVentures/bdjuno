package database

import (
	"encoding/json"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/forbole/bdjuno/v2/types"
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

// SaveMintParams allows to store the given params inside the database
func (db *Db) SaveMintParams(params *types.MintParams) error {
	paramsBz, err := json.Marshal(&params)
	if err != nil {
		return fmt.Errorf("error while marshaling mint params: %s", err)
	}

	stmt := `
INSERT INTO mint_params (params, height) 
VALUES ($1, $2)
ON CONFLICT (one_row_id) DO UPDATE 
    SET params = excluded.params,
        height = excluded.height
WHERE mint_params.height <= excluded.height`

	_, err = db.Sql.Exec(stmt, string(paramsBz), params.Height)
	if err != nil {
		return fmt.Errorf("error while storing mint params: %s", err)
	}

	return nil
}

func (db *Db) GetMintParams() (types.MintParams, error) {
	var mintParamsRows []string
	if err := db.Sqlx.Select(&mintParamsRows, `SELECT params FROM mint_params`); err != nil {
		return types.MintParams{}, fmt.Errorf("error while getting mint params: %s", err)
	}

	if len(mintParamsRows) == 0 {
		return types.MintParams{}, errors.New("mint params not found")
	}

	var mintParams types.MintParams
	if err := json.Unmarshal([]byte(mintParamsRows[0]), &mintParams); err != nil {
		return types.MintParams{}, fmt.Errorf("invalid mint params: %+v", mintParams)
	}

	return mintParams, nil
}
