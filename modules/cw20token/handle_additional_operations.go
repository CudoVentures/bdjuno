package cw20token

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"
	pubsub "github.com/forbole/bdjuno/v2/utils"
	gjv "github.com/xeipuuv/gojsonschema"
)

func (m *Module) RunAdditionalOperations() error {
	utils.WatchMethod(func() error {
		return m.pubsub.Subscribe(m.subscribeCallback)
	})
	return nil
}

func (m *Module) subscribeCallback(msg *pubsub.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		var contract types.VerifiedContractPublishMessage
		if err := json.Unmarshal(msg.Data, &contract); err != nil {
			msg.Ack()
			return err
		}

		exists, err := dbTx.IsExistingTokenCode(contract.CodeID)
		if err != nil {
			msg.Nack()
			return err
		}

		if exists {
			msg.Ack()
			return fmt.Errorf("contract is already tracked")
		}

		if !isToken(&contract) {
			msg.Ack()
			return fmt.Errorf("contract is not a cw20 token")
		}

		if err := dbTx.SaveTokenCode(&contract); err != nil {
			msg.Nack()
			return err
		}

		if err := m.saveExistingTokens(dbTx, contract.CodeID); err != nil {
			msg.Nack()
			return err
		}

		msg.Ack()
		return nil
	})
}

func (m *Module) saveExistingTokens(dbTx *database.DbTx, codeID int) error {
	contracts, err := dbTx.GetContractsByCodeID(codeID)
	if err != nil {
		return err
	}

	tokens, err := dbTx.GetAllTokenAddresses()
	if err != nil {
		return err
	}

	for _, contractAddress := range contracts {
		exists := false
		for _, t := range tokens {
			if t == contractAddress {
				exists = true
				continue
			}
		}
		if exists {
			continue
		}

		tokenInfo, balances, err := m.getTokenInfo(dbTx, contractAddress)
		if err != nil {
			return err
		}

		if err := dbTx.SaveToken(tokenInfo); err != nil {
			return err
		}

		if err := dbTx.SaveTokenBalances(balances); err != nil {
			return err
		}
	}

	return nil
}

func (m *Module) getTokenInfo(dbTx *database.DbTx, contractAddress string) (*types.TokenInfo, []types.TokenBalance, error) {
	block, err := dbTx.GetLastBlock()
	if err != nil {
		return nil, nil, err
	}
	state, err := m.source.AllContractState(contractAddress, block.Height)
	if err != nil {
		return nil, nil, err
	}

	tokenInfo := types.TokenInfo{}
	balances := []types.TokenBalance{}
	for _, s := range state {
		key := string(s.Key)

		if key == "token_info" {
			if err := json.Unmarshal(s.Value, &tokenInfo); err != nil {
				return nil, nil, err
			}
			continue
		}

		if strings.Contains(key, "balance") {
			balance, err := strconv.ParseUint(strings.ReplaceAll(string(s.Value), "\"", ""), 10, 64)
			if err != nil {
				return nil, nil, err
			}

			addressIndex := strings.Index(key, "cudos")
			balances = append(balances, types.TokenBalance{Address: key[addressIndex:], Amount: balance})
		}
	}

	tokenInfo.Address = contractAddress
	return &tokenInfo, balances, nil
}

func isToken(contract *types.VerifiedContractPublishMessage) bool {
	executeMsgs := []string{
		`{"transfer":{"recipient":"test","amount":"1"}}`,
		`{"send":{"contract":"test","amount":"1","msg":"test"}}`,
	}
	if isValid := validateSchema(contract.ExecuteSchema, executeMsgs); !isValid {
		return false
	}

	queryMsgs := []string{
		`{"balance":{"address":"test"}}`,
		`{"token_info":{}}`,
		`{"all_accounts":{}}`,
	}
	return validateSchema(contract.QuerySchema, queryMsgs)
}

func validateSchema(schema string, msgs []string) bool {
	for _, msg := range msgs {
		if result, err := gjv.Validate(gjv.NewStringLoader(schema), gjv.NewStringLoader(msg)); err != nil || !result.Valid() {
			if err != nil {
				// todo instead print, return these errors
				fmt.Print(err.Error())
			}
			if result != nil {
				for _, e := range result.Errors() {
					fmt.Print(e.String())
				}
			}
			return false
		}
	}
	return true
}
