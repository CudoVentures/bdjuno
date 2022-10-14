package cw20token

import (
	"encoding/json"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"
	"github.com/forbole/bdjuno/v2/utils/pubsub"
)

func (m *Module) RunAdditionalOperations() error {
	utils.WatchMethod(func() error {
		return m.pubsub.Subscribe(m.subscribeCallback)
	})

	return nil
}

// if a live service like db returns err, it might be caused by connection issues
// we put the msg back in queue with msg.Nack() so we can try to process it again
// if err is caused by business logic we remove it from queue with msg.Ack()
// we mark the msg based on ExecuteTx result, that's why we return nil on some errors
func (m *Module) subscribeCallback(msg *pubsub.PubSubMsg) {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		var contract types.MsgVerifiedContract
		if err := json.Unmarshal(msg.Data, &contract); err != nil {
			return nil
		}

		if err := validateTokenSchema(&contract); err != nil {
			return nil
		}

		found, err := dbTx.CodeIDExists(contract.CodeID)
		if err != nil {
			return err
		}

		if found {
			return nil
		}

		if err := dbTx.SaveCodeID(contract.CodeID); err != nil {
			return err
		}

		block, err := dbTx.GetLastBlock()
		if err != nil {
			return err
		}

		contracts, err := dbTx.GetContractsByCodeID(contract.CodeID)
		if err != nil {
			return err
		}

		for _, addr := range contracts {
			if err := m.saveTokenInfo(dbTx, addr, contract.CodeID, block.Height); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		msg.Nack()
		return
	}

	msg.Ack()
}
