package cw20token

import (
	"encoding/json"
	"fmt"

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

		if exists, err := dbTx.IsExistingTokenCode(contract.CodeID); err != nil {
			msg.Nack()
			return err
		} else if exists {
			msg.Ack()
			return fmt.Errorf("contract is already tracked")
		}

		if isToken, err := isToken(&contract); err != nil {
			msg.Nack()
			return err
		} else if !isToken {
			msg.Ack()
			return fmt.Errorf("contract is not a cw20 token")
		}

		if err := dbTx.SaveTokenCodeID(&contract); err != nil {
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

func (m *Module) saveExistingTokens(dbTx *database.DbTx, codeID uint64) error {
	contractAddresses, err := dbTx.GetContractsByCodeID(codeID)
	if err != nil {
		return err
	}

	for _, contract := range contractAddresses {
		if exists, err := dbTx.IsExistingToken(contract); err != nil {
			return err
		} else if exists {
			continue
		}

		block, err := dbTx.GetLastBlock()
		if err != nil {
			return err
		}

		if err := m.saveTokenInfo(dbTx, contract, block.Height); err != nil {
			return err
		}
	}

	return nil
}

func isToken(contract *types.VerifiedContractPublishMessage) (bool, error) {
	executeMsgs := []string{
		`{"transfer":{"recipient":"test","amount":"1"}}`,
		`{"send":{"contract":"test","amount":"1","msg":"test"}}`,
	}
	if err := validateSchema(contract.ExecuteSchema, executeMsgs); err != nil {
		return false, err
	}

	queryMsgs := []string{
		`{"balance":{"address":"test"}}`,
		`{"token_info":{}}`,
		`{"all_accounts":{}}`,
	}
	if err := validateSchema(contract.QuerySchema, queryMsgs); err != nil {
		return false, err
	}

	return true, nil
}

func validateSchema(schema string, msgs []string) error {
	for _, msg := range msgs {
		if result, err := gjv.Validate(gjv.NewStringLoader(schema), gjv.NewStringLoader(msg)); err != nil {
			return err
		} else if !result.Valid() {
			err := ""
			for _, e := range result.Errors() {
				err += e.String()
				err += "\n"
			}
			return fmt.Errorf("%s", err)
		}
	}
	return nil
}
