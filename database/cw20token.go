package database

import (
	"github.com/forbole/bdjuno/v2/types"
)

var (
	// todo real db
	verifiedContracts = []*types.VerifiedContractPublishMessage{}
	tokens            = []*types.TokenInfo{}
	tokenBalances     = []*types.TokenBalance{}
)

func (dbTx *DbTx) SaveTokenCodeID(contract *types.VerifiedContractPublishMessage) error {
	verifiedContracts = append(verifiedContracts, contract)
	return nil
}

func (dbTx *DbTx) SaveTokenInfo(token *types.TokenInfo) error {
	tokens = append(tokens, token)
	return nil
}

func (dbTx *DbTx) SaveTokenBalances(balances []*types.TokenBalance) error {
	tokenBalances = append(tokenBalances, balances...)
	return nil
}

func (dbTx *DbTx) UpdateTokenTotalSupply(contract string, totalSupply uint64) error {
	// todo
	return nil
}

func (dbTx *DbTx) UpdateTokenMinter(newMinter string) error {
	// todo
	return nil
}

func (dbTx *DbTx) UpdateTokenMarketing(contract string, project string, description string, admin string) error {
	// todo
	return nil
}

func (dbTx *DbTx) UpdateTokenLogo(contract string, logo string) error {
	// todo
	return nil
}

func (dbTx *DbTx) UpdateTokenCodeID(contract string, codeID uint64) error {
	// todo
	return nil
}

func (dbTx *DbTx) DeleteAllTokenBalances(contract string) error {
	// todo
	return nil
}

func (dbTx *DbTx) IsExistingToken(contract string) (bool, error) {
	// todo only return true if codeID is valid tokenCodeID (becauase of possible migrations) (maybe with join, so no codeID = no result)
	for _, t := range tokens {
		if t.Address == contract {
			return true, nil
		}
	}

	return false, nil
}

func (dbTx *DbTx) IsExistingTokenCode(codeID uint64) (bool, error) {
	for _, c := range verifiedContracts {
		if c.CodeID == codeID {
			return true, nil
		}
	}

	return false, nil
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
