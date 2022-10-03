package cw20token

import (
	"encoding/json"
	"fmt"

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
		case *wasmTypes.MsgMigrateContract:
			return m.handleMsgMigrateContract(dbTx, cosmosMsg, tx, index)
		default:
			return nil
		}
	})
}

func (m *Module) handleMsgInstantiateContract(dbTx *database.DbTx, msg *wasmTypes.MsgInstantiateContract, tx *juno.Tx, index int) error {
	exists, err := dbTx.IsExistingTokenCode(msg.CodeID)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	contract := utils.GetValueFromLogs(uint32(index), tx.Logs, wasmTypes.EventTypeInstantiate, wasmTypes.AttributeKeyContractAddr)
	if contract == "" {
		return fmt.Errorf("error while getting EventInstantiate")
	}
	return m.saveTokenInfo(dbTx, contract, msg.CodeID, tx.Height)
}

func (m *Module) handleMsgExecuteContract(dbTx *database.DbTx, msg *wasmTypes.MsgExecuteContract, tx *juno.Tx, index int) error {
	exists, err := dbTx.IsExistingToken(msg.Contract)
	if err != nil {
		return err
	}

	if !exists {
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
		return dbTx.UpdateTokenMinter(msg.Contract, msgDetails.NewMinter)
	case "update_marketing":
		return dbTx.UpdateTokenMarketing(msg.Contract, msgDetails.Project, msgDetails.Description, msgDetails.Admin)
	case "upload_logo":
		return dbTx.UpdateTokenLogo(msg.Contract, utils.SanitizeUTF8(string(msgRaw)))
	default:
		if err := m.saveBalances(dbTx, msg.Contract, msg.Sender, &msgDetails, tx.Height); err != nil {
			return err
		}

		if msgType == "mint" || msgType == "burn" || msgType == "burn_from" {
			return m.saveCirculatingSupply(dbTx, msg.Contract, tx.Height)
		}

		return nil
	}
}

func (m *Module) handleMsgMigrateContract(dbTx *database.DbTx, msg *wasmTypes.MsgMigrateContract, tx *juno.Tx, index int) error {
	exists, err := dbTx.IsExistingToken(msg.Contract)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	if err := dbTx.DeleteAllTokenBalances(msg.Contract); err != nil {
		return err
	}

	return m.saveTokenInfo(dbTx, msg.Contract, msg.CodeID, tx.Height)
}
