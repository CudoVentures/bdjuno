package database

import (
	"github.com/forbole/bdjuno/v2/types"
)

var (
	// todo real db
	verifiedContracts = []*types.VerifiedContractPublishMessage{}
	tokens            = []*types.TokenInfo{}
	tokenBalances     = []types.TokenBalance{}
)

func (dbTx *DbTx) SaveTokenCode(contract *types.VerifiedContractPublishMessage) error {
	verifiedContracts = append(verifiedContracts, contract)
	return nil
}

func (dbTx *DbTx) SaveToken(token *types.TokenInfo) error {
	tokens = append(tokens, token)
	return nil
}

func (dbTx *DbTx) SaveTokenBalances(balances []types.TokenBalance) error {
	tokenBalances = append(tokenBalances, balances...)
	return nil
}

func (dbTx *DbTx) IsExistingToken(contractAddress string) (bool, error) {
	for _, t := range tokens {
		if t.Address == contractAddress {
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
