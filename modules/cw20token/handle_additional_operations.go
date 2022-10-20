package cw20token

import (
	"encoding/json"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"
	txutils "github.com/forbole/bdjuno/v2/utils"
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
		var c types.MsgVerifiedContract
		if err := json.Unmarshal(msg.Data, &c); err != nil {
			return nil
		}

		if err := validateTokenSchema(&c); err != nil {
			return nil
		}

		found, err := dbTx.CodeIDExists(c.CodeID)
		if err != nil {
			return err
		}

		if found {
			return nil
		}

		if err := dbTx.SaveCodeID(c.CodeID); err != nil {
			return err
		}

		block, err := dbTx.GetLastBlock()
		if err != nil {
			return err
		}

		contracts, err := dbTx.GetContractsByCodeID(c.CodeID)
		if err != nil {
			return err
		}

		for _, addr := range contracts {
			tx, err := txutils.NewTx(time.Now(), "", uint64(block.Height)).WithEventInstantiateContract(addr).Build()
			if err != nil {
				return err
			}

			if err := m.handleMsgInstantiateContract(dbTx, &wasm.MsgInstantiateContract{CodeID: c.CodeID}, tx, 0); err != nil {
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
