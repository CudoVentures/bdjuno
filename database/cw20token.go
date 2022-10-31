package database

import (
	"fmt"

	"github.com/forbole/bdjuno/v2/types"
)

func (dbTx *DbTx) SaveCodeID(codeID uint64) error {
	_, err := dbTx.Exec(`INSERT INTO cw20token_code_id VALUES ($1) ON CONFLICT DO NOTHING`, codeID)
	return err
}

func (dbTx *DbTx) SaveInfo(token types.TokenInfo) error {
	_, err := dbTx.Exec(
		`INSERT INTO cw20token_info VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (address) DO UPDATE SET
		code_id = excluded.code_id, name = excluded.name, symbol = excluded.symbol, decimals = excluded.decimals,
		circulating_supply = excluded.circulating_supply, max_supply = excluded.max_supply, minter = excluded.minter,
		marketing_admin = excluded.marketing_admin, project_url = excluded.project_url, description = excluded.description, logo = excluded.logo`,
		token.Address, token.CodeID, token.Name, token.Symbol, token.Decimals, token.TotalSupply, token.Mint.MaxSupply,
		token.Mint.Minter, token.Marketing.Admin, token.Marketing.Project, token.Marketing.Description, token.Marketing.Logo,
	)

	return err
}

func (dbTx *DbTx) SaveBalances(token string, balances []types.TokenBalance) error {
	stmt := "INSERT INTO cw20token_balance VALUES "
	var params []interface{}
	for i, b := range balances {
		n := i * 3
		stmt += fmt.Sprintf("($%d, $%d, $%d),", n+1, n+2, n+3)
		params = append(params, b.Address, token, b.Amount)
	}

	stmt = stmt[:len(stmt)-1]
	stmt += `ON CONFLICT (address, token) DO UPDATE SET balance = excluded.balance`
	_, err := dbTx.Exec(stmt, params...)
	if err != nil {
		return err
	}

	_, err = dbTx.Exec(`DELETE FROM cw20token_balance WHERE token = $1 AND balance <= 0`, token)
	return err
}

func (dbTx *DbTx) UpdateSupply(token string, supply uint64) error {
	_, err := dbTx.Exec(`UPDATE cw20token_info SET circulating_supply = $1 WHERE address = $2`, supply, token)
	return err
}

func (dbTx *DbTx) UpdateMinter(token string, minter string) error {
	_, err := dbTx.Exec(`UPDATE cw20token_info SET minter = $1 WHERE address = $2`, minter, token)
	return err
}

func (dbTx *DbTx) UpdateMarketing(token string, m types.Marketing) error {
	_, err := dbTx.Exec(
		`UPDATE cw20token_info SET project_url = $1, description = $2, marketing_admin = $3 WHERE address = $4`,
		m.Project, m.Description, m.Admin, token,
	)
	return err
}

func (dbTx *DbTx) UpdateLogo(token string, logo string) error {
	_, err := dbTx.Exec(`UPDATE cw20token_info SET logo = $1 WHERE address = $2`, logo, token)
	return err
}

func (dbTx *DbTx) UpdateCodeID(token string, codeID uint64) error {
	_, err := dbTx.Exec(`UPDATE cw20token_info SET code_id = $1 WHERE address = $2`, codeID, token)
	return err
}

func (dbTx *DbTx) DeleteToken(token string) error {
	_, err := dbTx.Exec(`DELETE FROM cw20token_info WHERE address = $1`, token)
	return err
}

func (dbTx *DbTx) TokenExists(token string) (bool, error) {
	var found bool
	err := dbTx.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM cw20token_info WHERE address = $1)`, token,
	).Scan(&found)
	return found, err
}

func (dbTx *DbTx) CodeIDExists(codeID uint64) (bool, error) {
	var found bool
	err := dbTx.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM cw20token_code_id WHERE id = $1)`, codeID,
	).Scan(&found)
	return found, err
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
