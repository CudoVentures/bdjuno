package cw20token

import (
	"encoding/json"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"
	ps "github.com/forbole/bdjuno/v2/utils"
	"github.com/rs/zerolog/log"
)

func (m *Module) RunAdditionalOperations() error {
	utils.WatchMethod(func() error {
		return m.pubsub.Subscribe(m.subscribeCallback)
	})

	return nil
}

// if a live service like db returns err we put the msg back in queue for retry with msg.Nack()
// because it's theoretically possible that the service has been down for a moment
// if a business logic service returns err we mark the msg as processed with msg.Ack()
// on some errors we return nil and on others we return the err, because on the bottom
// of the subscribeCallback we mark the msg based on the ExecuteTx result
func (m *Module) subscribeCallback(msg *ps.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		var contract types.VerifiedContractPublishMessage
		if err := json.Unmarshal(msg.Data, &contract); err != nil {
			// todo if logging works remove this description
			//
			// here we have an exception. err is not depending on a service status
			// that's why we acknowledge the msg, since it doesn't meet the requirements
			// we also return the error, because it's not predictable and may be informative
			log.Info().Str("module", "cw20token").Msg("error while unmarshaling publish message from json")
			return nil
		}

		if exists, err := dbTx.IsExistingTokenCode(contract.CodeID); err != nil {
			return err
		} else if exists {
			log.Info().Str("module", "cw20token").Msg("token is already tracked")
			return nil
		}

		if isToken, err := isToken(&contract); err != nil {
			log.Info().Str("module", "cw20token").Msg("invalid json schema")
			return nil
		} else if !isToken {
			log.Info().Str("module", "cw20token").Msg("contract doesn't match the cw20 standard")
			return nil
		}

		if err := dbTx.SaveTokenCodeID(contract.CodeID); err != nil {
			return err
		}

		block, err := dbTx.GetLastBlock()
		if err != nil {
			return err
		}

		contracts, err := getUntrackedTokens(dbTx, contract.CodeID)
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
		// todo it's possible that this log actually breaks the subscription
		log.Error().Str("module", "cw20token").Err(err)
	} else {
		msg.Ack()
	}
}

func getUntrackedTokens(dbTx *database.DbTx, codeID uint64) ([]string, error) {
	contracts, err := dbTx.GetContractsByCodeID(codeID)
	if err != nil {
		return nil, err
	}

	res := []string{}
	for _, addr := range contracts {
		if exists, err := dbTx.IsExistingToken(addr); err != nil {
			return nil, err
		} else if exists {
			continue
		}

		res = append(res, addr)
	}

	return res, nil
}
