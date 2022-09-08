package cosmwasm

import (
	"encoding/json"
	"fmt"
	"strconv"

	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"
	juno "github.com/forbole/juno/v3/types"
)

// HandleMsg implements MessageModule
func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	if len(tx.Logs) == 0 {
		return nil
	}

	switch cosmosMsg := msg.(type) {
	case *wasmTypes.MsgStoreCode:
		return m.handleMsgStoreCode(index, cosmosMsg, tx)

	case *wasmTypes.MsgInstantiateContract:
		return m.handleMsgInstantiateContract(index, cosmosMsg, tx)

	case *wasmTypes.MsgExecuteContract:
		return m.handleMsgExecuteContract(index, cosmosMsg, tx)

	case *wasmTypes.MsgMigrateContract:
		return m.handleMsgMigrateContract(index, cosmosMsg, tx)

	case *wasmTypes.MsgUpdateAdmin:
		return m.handleMsgUpdateAdmin(index, cosmosMsg, tx)

	case *wasmTypes.MsgClearAdmin:
		return m.handleMsgClearAdmin(index, cosmosMsg, tx)
	}

	return nil
}

func (m *Module) handleMsgStoreCode(index int, msg *wasmTypes.MsgStoreCode, tx *juno.Tx) error {
	instantiatePermissionPtr := &wasmTypes.AccessConfig{}

	if msg.InstantiatePermission != nil {
		instantiatePermissionPtr = msg.InstantiatePermission
	}

	instantiatePermission, err := json.Marshal(instantiatePermissionPtr)
	if err != nil {
		return err
	}

	resultCodeID := utils.GetValueFromLogs(uint32(index), tx.Logs, wasmTypes.EventTypeStoreCode, wasmTypes.AttributeKeyCodeID)

	return m.db.SaveMsgStoreCodeData(
		types.NewMsgStoreCodeData(
			tx.TxHash,
			msg.Sender,
			index,
			isSuccess(tx.Code),
			utils.SanitizeUTF8(string(instantiatePermission)),
			resultCodeID,
		),
	)
}

func (m *Module) handleMsgInstantiateContract(index int, msg *wasmTypes.MsgInstantiateContract, tx *juno.Tx) error {
	funds, err := json.Marshal(msg.Funds)
	if err != nil {
		return err
	}

	resultContractAddress := utils.GetValueFromLogs(uint32(index), tx.Logs, wasmTypes.EventTypeInstantiate, wasmTypes.AttributeKeyContractAddr)

	return m.db.SaveMsgInstantiateContractData(
		types.NewMsgInstantiateContractData(
			tx.TxHash,
			msg.Sender,
			index,
			isSuccess(tx.Code),
			msg.Admin,
			utils.SanitizeUTF8(string(funds)),
			msg.Label,
			strconv.FormatUint(msg.CodeID, 10),
			resultContractAddress,
		),
	)
}

func (m *Module) handleMsgExecuteContract(index int, msg *wasmTypes.MsgExecuteContract, tx *juno.Tx) error {
	funds, err := json.Marshal(msg.Funds)
	if err != nil {
		return err
	}

	payload := make(map[string]interface{})
	if err := json.Unmarshal(msg.Msg, &payload); err != nil {
		return err
	}

	payloadKeys := getPayloadMapKeys(payload)
	if len(payloadKeys) != 1 {
		return fmt.Errorf("handleMsgExecuteContract message payload has more than one root property: %s", string(msg.Msg))
	}

	method := payloadKeys[0]
	arguments, err := json.Marshal(payload[payloadKeys[0]])
	if err != nil {
		return err
	}

	return m.db.SaveMsgExecuteContractData(
		types.NewMsgExecuteContractData(
			tx.TxHash,
			msg.Sender,
			index,
			isSuccess(tx.Code),
			method,
			utils.SanitizeUTF8(string(arguments)),
			utils.SanitizeUTF8(string(funds)),
			msg.Contract,
		),
	)
}

func (m *Module) handleMsgMigrateContract(index int, msg *wasmTypes.MsgMigrateContract, tx *juno.Tx) error {

	return m.db.SaveMsgMigrateContactData(
		types.NewMsgMigrateContractData(
			tx.TxHash,
			msg.Sender,
			index,
			isSuccess(tx.Code),
			msg.Contract,
			strconv.FormatUint(msg.CodeID, 10),
			utils.SanitizeUTF8(string(msg.Msg)),
		),
	)
}

func (m *Module) handleMsgUpdateAdmin(index int, msg *wasmTypes.MsgUpdateAdmin, tx *juno.Tx) error {

	return m.db.SaveMsgUpdateAdminData(
		types.NewMsgUpdateAdminData(
			tx.TxHash,
			msg.Sender,
			index,
			isSuccess(tx.Code),
			msg.Contract,
			msg.NewAdmin,
		),
	)
}

func (m *Module) handleMsgClearAdmin(index int, msg *wasmTypes.MsgClearAdmin, tx *juno.Tx) error {

	return m.db.SaveMsgClearAdminData(
		types.NewClearAdminData(
			tx.TxHash,
			msg.Sender,
			index,
			isSuccess(tx.Code),
			msg.Contract,
		),
	)
}

func isSuccess(code uint32) bool {
	return code == 0
}

func getPayloadMapKeys(payloadMap map[string]interface{}) []string {
	keys := make([]string, 0, len(payloadMap))
	for k := range payloadMap {
		keys = append(keys, k)
	}
	return keys
}
