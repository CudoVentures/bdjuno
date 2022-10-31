package cw20token

import (
	"encoding/json"
	"strconv"

	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v2/database"
	mutils "github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"
	"github.com/forbole/bdjuno/v2/utils"
	juno "github.com/forbole/juno/v2/types"
)

func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	if len(tx.Logs) == 0 {
		return nil
	}

	return m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		switch cosmosMsg := msg.(type) {
		case *wasm.MsgStoreCode:
			return m.handleMsgStoreCode(dbTx, cosmosMsg, tx, index)
		case *wasm.MsgInstantiateContract:
			return m.handleMsgInstantiateContract(dbTx, cosmosMsg, tx, index)
		case *wasm.MsgExecuteContract:
			return m.handleMsgExecuteContract(dbTx, cosmosMsg, tx, index)
		case *wasm.MsgMigrateContract:
			return m.handleMsgMigrateContract(dbTx, cosmosMsg, tx, index)
		default:
			return nil
		}
	})
}
func (m *Module) handleMsgStoreCode(dbTx *database.DbTx, msg *wasm.MsgStoreCode, tx *juno.Tx, index int) error {
	if err := utils.ValidateContract(msg.WASMByteCode, utils.CW20); err != nil {
		return nil
	}

	codeIDAttr := mutils.GetValueFromLogs(uint32(index), tx.Logs, wasm.EventTypeStoreCode, wasm.AttributeKeyCodeID)
	codeID, err := strconv.ParseUint(codeIDAttr, 10, 64)
	if err != nil {
		return err
	}

	return dbTx.SaveCodeID(codeID)
}

func (m *Module) handleMsgInstantiateContract(dbTx *database.DbTx, msg *wasm.MsgInstantiateContract, tx *juno.Tx, index int) error {
	if found, err := dbTx.CodeIDExists(msg.CodeID); !found {
		return err
	}

	contractAddr := mutils.GetValueFromLogs(uint32(index), tx.Logs, wasm.EventTypeInstantiate, wasm.AttributeKeyContractAddr)
	tokenInfo, err := m.source.TokenInfo(contractAddr, tx.Height)
	if err != nil {
		return err
	}
	tokenInfo.CodeID = msg.CodeID

	if err := dbTx.SaveInfo(tokenInfo); err != nil {
		return err
	}

	balances, err := m.source.AllBalances(contractAddr, tx.Height)
	if err != nil {
		return err
	}

	return dbTx.SaveBalances(contractAddr, balances)
}

func (m *Module) handleMsgExecuteContract(dbTx *database.DbTx, msg *wasm.MsgExecuteContract, tx *juno.Tx, index int) error {
	if found, err := dbTx.TokenExists(msg.Contract); err != nil || !found {
		return err
	}

	msgType := mutils.GetValueFromLogs(uint32(index), tx.Logs, wasm.WasmModuleEventType, "action")
	r, err := parseToMsgExecuteToken(msg)
	if err != nil {
		return err
	}

	// todo a cool idea is to make a map(similar to current approach)
	// but against each key place msg type
	// the type would be map[string]interface{}
	// then access it with map[msgType].(*SpecificType)
	// that way we only Unmarshal once and looks a good match with that switch
	// also we can have []string and append to it when msgType modifies balance
	// then after the switch we make for range []string update balance - easy
	switch msgType {
	case "update_minter":
		return dbTx.UpdateMinter(r.Contract, r.NewMinter)
	case "update_marketing":
		return dbTx.UpdateMarketing(r.Contract, types.NewMarketing(r.Project, r.Description, r.MarketingAdmin, nil))
	case "upload_logo":
		return dbTx.UpdateLogo(r.Contract, mutils.SanitizeUTF8(string(r.MsgRaw)))
	case "mint", "burn", "burn_from":
		supply, err := m.source.TotalSupply(r.Contract, tx.Height)
		if err != nil {
			return err
		}

		if err := dbTx.UpdateSupply(r.Contract, supply); err != nil {
			return err
		}
	}

	balances, err := m.fetchBalances(r, tx.Height)
	if err != nil {
		return err
	}

	return dbTx.SaveBalances(r.Contract, balances)
}

func parseToMsgExecuteToken(msg *wasm.MsgExecuteContract) (*types.MsgExecuteToken, error) {
	req := map[string]json.RawMessage{}
	if err := json.Unmarshal(msg.Msg, &req); err != nil {
		return nil, err
	}

	res := types.MsgExecuteToken{}
	for msgType, msgRaw := range req {
		if err := json.Unmarshal(msgRaw, &res); err != nil {
			return nil, err
		}

		res.Type = msgType
		res.MsgRaw = msgRaw
	}

	res.Contract = msg.Contract
	res.Sender = msg.Sender

	return &res, nil
}

func (m *Module) fetchBalances(msg *types.MsgExecuteToken, height int64) ([]types.TokenBalance, error) {
	balances := []types.TokenBalance{}

	sender := msg.Sender
	if msg.Owner != "" {
		sender = msg.Owner
	}
	balances = append(balances, types.TokenBalance{Address: sender})

	if msg.Recipient != "" {
		balances = append(balances, types.TokenBalance{Address: msg.Recipient})
	}

	if msg.RecipientContract != "" {
		balances = append(balances, types.TokenBalance{Address: msg.RecipientContract})
	}

	for i, b := range balances {
		balance, err := m.source.Balance(msg.Contract, b.Address, height)
		if err != nil {
			return nil, err
		}

		balances[i].Amount = balance
	}

	return balances, nil
}

func (m *Module) handleMsgMigrateContract(dbTx *database.DbTx, msg *wasm.MsgMigrateContract, tx *juno.Tx, index int) error {
	if found, err := dbTx.TokenExists(msg.Contract); err != nil || !found {
		return err
	}

	found, err := dbTx.CodeIDExists(msg.CodeID)
	if err != nil {
		return err
	}

	if !found {
		return dbTx.DeleteToken(msg.Contract)
	}

	return dbTx.UpdateCodeID(msg.Contract, msg.CodeID)
}
