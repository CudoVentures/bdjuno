package database

import (
	"fmt"

	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/types"
)

func (dbTx *DbTx) SaveCodeID(codeID uint64) error {
	_, err := dbTx.Exec(`INSERT INTO cw20token_code_id VALUES ($1) ON CONFLICT DO NOTHING`, codeID)
	return err
}

func (dbTx *DbTx) SaveInfo(t types.TokenInfo) error {
	_, err := dbTx.Exec(
		`INSERT INTO cw20token_info VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (address) DO UPDATE SET
		code_id = excluded.code_id, name = excluded.name, symbol = excluded.symbol, decimals = excluded.decimals,
		circulating_supply = excluded.circulating_supply, max_supply = excluded.max_supply, minter = excluded.minter,
		marketing_admin = excluded.marketing_admin, project_url = excluded.project_url, description = excluded.description, logo = excluded.logo`,
		t.Address, t.CodeID, t.Name, t.Symbol, t.Decimals, t.TotalSupply, t.TotalSupply, t.Mint.MaxSupply, t.Mint.Minter,
		t.Marketing.Admin, t.Marketing.Project, t.Marketing.Description, t.Marketing.Logo, t.Type, t.Creator,
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

	_, err = dbTx.Exec(`DELETE FROM cw20token_balance WHERE token = $1 AND balance = '0'`, token)
	return err
}

func (dbTx *DbTx) UpdateSupply(token string, supply string) error {
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

func (dbTx *DbTx) SaveAllowance(token, owner, spender, amount, expires string) error {
	if amount == "0" {
		_, err := dbTx.Exec(
			`DELETE FROM cw20token_allowance WHERE token = $1 AND owner = $2 AND spender = $3`,
			token, owner, spender,
		)
		return err
	}

	_, err := dbTx.Exec(
		`INSERT INTO cw20token_allowance VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (token, owner, spender) DO UPDATE SET
		amount = excluded.amount, expires = excluded.expires`,
		token, owner, spender, amount, expires,
	)
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

func (dbTx *DbTx) GetAllAllowances() ([]dbtypes.AllowanceRow, error) {
	rows, err := dbTx.Query(`SELECT * FROM cw20token_allowance`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	allowances := []dbtypes.AllowanceRow{}
	for rows.Next() {
		var a dbtypes.AllowanceRow
		if err := rows.Scan(&a.Token, &a.Owner, &a.Spender, &a.Amount, &a.Expires); err != nil {
			return nil, err
		}

		allowances = append(allowances, a)
	}

	return allowances, rows.Err()
}
