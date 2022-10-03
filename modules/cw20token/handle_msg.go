package cw20token

import (
	"encoding/json"

	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"
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
			return m.handleMsgExecuteContract(dbTx, cosmosMsg, tx, index)
		}
		// todo if migrate - we should update sth (research what)
		return nil
	})
}

func (m *Module) handleMsgInstantiateContract(dbTx *database.DbTx, msg *wasmTypes.MsgInstantiateContract, tx *juno.Tx, index int) error {
	if exists, err := dbTx.IsExistingTokenCode(msg.CodeID); err != nil {
		return err
	} else if !exists {
		return nil
	}
	contractAddress := utils.GetValueFromLogs(uint32(index), tx.Logs, wasmTypes.EventTypeInstantiate, wasmTypes.AttributeKeyContractAddr)

	return m.saveTokenInfo(dbTx, contractAddress, tx.Height)
}

func (m *Module) handleMsgExecuteContract(dbTx *database.DbTx, msg *wasmTypes.MsgExecuteContract, tx *juno.Tx, index int) error {
	if exists, err := dbTx.IsExistingToken(msg.Contract); err != nil {
		return err
	} else if !exists {
		return nil
	}

	req := map[string]json.RawMessage{}
	if err := json.Unmarshal(msg.Msg, &req); err != nil {
		return err
	}

	var msgType string
	var msgRaw []byte
	for key, val := range req {
		msgType = key
		msgRaw = val
	}

	msgDetails := types.MsgTokenExecute{}
	if err := json.Unmarshal(msgRaw, &msgDetails); err != nil {
		return err
	}

	switch msgType {
	case "update_minter":
		return dbTx.UpdateTokenMinter(msgDetails.NewMinter)
	case "update_marketing":
		return dbTx.UpdateTokenMarketing(msgDetails.Project, msgDetails.Description, msgDetails.Admin)
	case "upload_logo":
		return dbTx.UpdateTokenLogo()
	default:
		if err := m.updateBalances(dbTx, msg.Contract, msg.Sender, &msgDetails, tx.Height); err != nil {
			return err
		}

		if msgType == "mint" || msgType == "burn" || msgType == "burn_from" {
			return m.updateTotalSupply(dbTx, msg.Contract, tx.Height)
		}

		return nil
	}
}
