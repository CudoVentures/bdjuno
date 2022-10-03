package database

import (
	"fmt"

	"github.com/forbole/bdjuno/v2/types"
)

func (dbTx *DbTx) SaveTokenCodeID(codeID uint64) error {
	_, err := dbTx.Exec(`INSERT INTO cw20token_code_id VALUES ($1) ON CONFLICT DO NOTHING`, codeID)
	return err
}

func (dbTx *DbTx) SaveTokenInfo(token *types.TokenInfo) error {
	_, err := dbTx.Exec(
		`INSERT INTO cw20token_info VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (address) DO UPDATE SET
		code_id = excluded.code_id, name = excluded.name, symbol = excluded.symbol, decimals = excluded.decimals,
		circulating_supply = excluded.circulating_supply, max_supply = excluded.max_supply, minter = excluded.minter,
		marketing_admin = excluded.marketing_admin, project_url = excluded.project_url, description = excluded.description, logo = excluded.logo`,
		token.Address, token.CodeID, token.Name, token.Symbol, token.Decimals, token.CirculatingSupply, token.MintInfo.MaxSupply,
		token.MintInfo.Minter, token.MarketingInfo.Admin, token.MarketingInfo.Project, token.MarketingInfo.Description, token.Logo,
	)
	return err
}

func (dbTx *DbTx) SaveTokenBalances(contract string, balances []*types.TokenBalance) error {
	stmt := "INSERT INTO cw20token_balance VALUES "
	var params []interface{}
	for i, b := range balances {
		n := i * 3
		stmt += fmt.Sprintf("($%d, $%d, $%d),", n+1, n+2, n+3)
		params = append(params, b.Address, contract, b.Amount)
	}

	stmt = stmt[:len(stmt)-1]
	stmt += `ON CONFLICT (address, token) DO UPDATE SET balance = excluded.balance`

	_, err := dbTx.Exec(stmt, params...)
	return err
}

func (dbTx *DbTx) UpdateTokenCirculatingSupply(contract string, supply uint64) error {
	_, err := dbTx.Exec(`UPDATE cw20token_info SET circulating_supply = $1 WHERE address = $2`, supply, contract)
	return err
}

func (dbTx *DbTx) UpdateTokenMinter(contract string, minter string) error {
	_, err := dbTx.Exec(`UPDATE cw20token_info SET minter = $1 WHERE address = $2`, minter, contract)
	return err
}

func (dbTx *DbTx) UpdateTokenMarketing(contract string, projectUrl string, description string, admin string) error {
	_, err := dbTx.Exec(`UPDATE cw20token_info SET project_url = $1, description = $2, marketing_admin = $3 WHERE address = $4`, projectUrl, description, admin, contract)
	return err
}

func (dbTx *DbTx) UpdateTokenLogo(contract string, logo string) error {
	_, err := dbTx.Exec(`UPDATE cw20token_info SET logo = $1 WHERE address = $2`, logo, contract)
	return err
}

func (dbTx *DbTx) UpdateTokenCodeID(contract string, codeID uint64) error {
	_, err := dbTx.Exec(`UPDATE cw20token_info SET code_id = $1 WHERE address = $2`, codeID, contract)
	return err
}

func (dbTx *DbTx) DeleteAllTokenBalances(contract string) error {
	// todo check cascading errors
	_, err := dbTx.Exec(`DELETE FROM cw20token_balance WHERE token = $1`, contract)
	return err
}

func (dbTx *DbTx) IsExistingToken(contract string) (bool, error) {
	var exists bool
	// todo test if the join excludes non-existing cw20token_code_id
	err := dbTx.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM cw20token_info i JOIN cw20token_code_id c ON c.id = i.code_id WHERE i.address = $1)`, contract,
	).Scan(&exists)
	return exists, err
}

func (dbTx *DbTx) IsExistingTokenCode(codeID uint64) (bool, error) {
	var exists bool
	err := dbTx.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM cw20token_code_id WHERE code_id = $1)`, codeID,
	).Scan(&exists)
	return exists, err
}

func (dbTx *DbTx) GetContractsByCodeID(codeID uint64) ([]string, error) {
	rows, err := dbTx.Query(`SELECT result_contract_address FROM cosmwasm_instantiate WHERE code_id = $1`, codeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contracts []string
	for rows.Next() {
		var contract string
		if err := rows.Scan(&contract); err != nil {
			return nil, err
		}
		contracts = append(contracts, contract)
	}

	return contracts, rows.Err()
}
