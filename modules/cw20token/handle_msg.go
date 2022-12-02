package cw20token

import (
	"encoding/json"
	"strconv"
	"strings"

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
			return m.handleMsgMigrateContract(dbTx, cosmosMsg)
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
	tokenInfo.Creator = msg.Sender

	switch {

	case strings.Contains(string(msg.Msg), "standard"):
		tokenInfo.Type = "standard"
		tokenInfo.Mint.MaxSupply = tokenInfo.TotalSupply

	case strings.Contains(string(msg.Msg), "mintable"):
		tokenInfo.Type = "mintable"
		tokenInfo.Mint.MaxSupply = tokenInfo.TotalSupply

	case strings.Contains(string(msg.Msg), "burnable"):
		tokenInfo.Type = "burnable"

		tokenInfo.Mint.MaxSupply = tokenInfo.TotalSupply
	case strings.Contains(string(msg.Msg), "unlimited"):
		tokenInfo.Type = "unlimited"

	default:
		tokenInfo.Type = ""
	}

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
	if found, err := dbTx.TokenExists(msg.Contract); !found {
		return err
	}

	msgExecute := types.MsgExecute{}
	if err := json.Unmarshal(msg.Msg, &msgExecute); err != nil {
		return err
	}

	msgType := mutils.GetValueFromLogs(uint32(index), tx.Logs, wasm.WasmModuleEventType, sdk.AttributeKeyAction)
	addresses := []string{}

	switch types.TypeMsgExecute(msgType) {
	case types.TypeTransfer:
		addresses = append(addresses, msgExecute.Transfer.Recipient, msg.Sender)
	case types.TypeTransferFrom:
		mm := msgExecute.TransferFrom
		addresses = append(addresses, mm.Owner, mm.Recipient)
	case types.TypeSend:
		addresses = append(addresses, msgExecute.Send.Contract, msg.Sender)
	case types.TypeSendFrom:
		mm := msgExecute.SendFrom
		addresses = append(addresses, mm.Owner, mm.Contract)
	case types.TypeBurn:
		addresses = append(addresses, msg.Sender)
	case types.TypeBurnFrom:
		addresses = append(addresses, msgExecute.BurnFrom.Owner)
	case types.TypeMint:
		addresses = append(addresses, msgExecute.Mint.Recipient)
	case types.TypeUpdateMinter:
		return dbTx.UpdateMinter(msg.Contract, msgExecute.UpdateMinter.NewMinter)
	case types.TypeUpdateMarketing:
		mm := msgExecute.UpdateMarketing
		return dbTx.UpdateMarketing(msg.Contract, types.Marketing{mm.Project, mm.Description, mm.Admin, nil})
	case types.TypeUploadLogo:
		return dbTx.UpdateLogo(msg.Contract, mutils.SanitizeUTF8(string(msgExecute.UploadLogo)))
	}

	supply, err := m.source.TotalSupply(msg.Contract, tx.Height)
	if err != nil {
		return err
	}

	if err := dbTx.UpdateSupply(msg.Contract, supply); err != nil {
		return err
	}

	balances := make([]types.TokenBalance, len(addresses))
	for i, a := range addresses {
		b, err := m.source.Balance(msg.Contract, a, tx.Height)
		if err != nil {
			return err
		}

		balances[i] = types.TokenBalance{a, b}
	}

	return dbTx.SaveBalances(msg.Contract, balances)
}

func (m *Module) handleMsgMigrateContract(dbTx *database.DbTx, msg *wasm.MsgMigrateContract) error {
	if found, err := dbTx.TokenExists(msg.Contract); !found {
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
