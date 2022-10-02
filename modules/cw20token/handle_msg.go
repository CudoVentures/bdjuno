package cw20token

import (
	"fmt"

	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/utils"
	juno "github.com/forbole/juno/v2/types"
)

func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(tx.Logs) == 0 {
		return nil
	}

	return m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		switch cosmosMsg := msg.(type) {
		case *wasmTypes.MsgInstantiateContract:
			return m.handleMsgInstantiateContract(dbTx, cosmosMsg, tx, index)
		case *wasmTypes.MsgExecuteContract:
			return nil
		}

		return nil
	})
}

func (m *Module) handleMsgInstantiateContract(dbTx *database.DbTx, msg *wasmTypes.MsgInstantiateContract, tx *juno.Tx, index int) error {
	// todo test
	if exists, err := dbTx.IsExistingTokenCode(msg.CodeID); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("contract is not a token")
	}

	contractAddress := utils.GetValueFromLogs(uint32(index), tx.Logs, wasmTypes.EventTypeInstantiate, wasmTypes.AttributeKeyContractAddr)

	return m.saveToken(dbTx, contractAddress)
}
