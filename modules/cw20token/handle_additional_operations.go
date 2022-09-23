package cw20token

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"
	gjv "github.com/xeipuuv/gojsonschema"
)

var (
	// todo os.Getenv cant read .env
	projectID = "multisig-firestore-1"
	subID     = "my-sub"
)

func (m *Module) RunAdditionalOperations() error {
	ctx := context.Background()
	// todo move this as dep to Module (this is untestable)
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return err
	}

	sub := client.Subscription(subID)
	sub.ReceiveSettings.MaxOutstandingMessages = 1
	sub.ReceiveSettings.NumGoroutines = 1

	utils.WatchMethod(func() error {
		m.subscribeToVerifiedContracts(ctx, sub)
		return nil
	})

	return nil
}

func (m *Module) subscribeToVerifiedContracts(ctx context.Context, sub *pubsub.Subscription) {
	var mu sync.Mutex
	sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		mu.Lock()
		defer mu.Unlock()

		m.db.ExecuteTx(func(dbTx *database.DbTx) error {
			var contract types.VerifiedContractPublishMessage

			if err := json.Unmarshal(msg.Data, &contract); err != nil {
				msg.Ack()
				return err
			}

			exists, err := dbTx.IsExistingTokenCode(contract.CodeID)
			if err != nil {
				// todo set maximum msg receives, cuz currently we considered that the db has dropped, but if it hasn't, we will keep receiving this msg forever spam
				msg.Nack()
				return err
			}

			if exists {
				msg.Ack()
				return fmt.Errorf("contract is already tracked")
			}

			if !isToken(ctx, &contract) {
				msg.Ack()
				return fmt.Errorf("contract is not a cw20 token")
			}

			if err := m.saveToken(dbTx, &contract); err != nil {
				msg.Nack()
				return err
			}

			msg.Ack()
			return nil
		})
	})

}

func (m *Module) saveToken(dbTx *database.DbTx, contract *types.VerifiedContractPublishMessage) error {
	if err := dbTx.SaveTokenCode(contract); err != nil {
		return err
	}

	contracts, err := dbTx.GetContractsByCodeID(contract.CodeID)
	if err != nil {
		return err
	}

	tokens, err := dbTx.GetAllTokens()
	if err != nil {
		return err
	}

	for _, c := range contracts {
		exists := false
		for _, t := range tokens {
			if t == c {
				exists = true
				continue
			}
		}
		if exists {
			continue
		}

		if err := dbTx.SaveToken(c); err != nil {
			return err
		}
		// todo update balances supply etc!!!! fetchTokenInfo()
		block, err := dbTx.GetLastBlock()
		if err != nil {
			return err
		}
		state, err := m.source.AllContractState(c, block.Height)
		if err != nil {
			return err
		}
		fmt.Printf("state: %v\n", state)
	}

	return nil
}

func isToken(ctx context.Context, contract *types.VerifiedContractPublishMessage) bool {
	instantiateMsgs := []string{`{"name":"test","symbol":"test","decimals":2,"initial_balances":[]}'`}
	if isValid := validateSchema(contract.InstantiateSchema, instantiateMsgs); !isValid {
		return false
	}

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
			fmt.Print(err.Error())
			for _, e := range result.Errors() {
				fmt.Print(e.String())
			}
			return false
		}
	}
	return true
}
