package cw20token

import (
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
	if exists, err := dbTx.IsExistingTokenCode(msg.CodeID); err != nil {
		return err
	} else if !exists {
		return nil
	}

	contract := utils.GetValueFromLogs(uint32(index), tx.Logs, wasmTypes.EventTypeInstantiate, wasmTypes.AttributeKeyContractAddr)
	if contract == "" {
		return fmt.Errorf("error while getting EventInstantiate")
	}

	return m.saveTokenInfo(dbTx, contract, msg.CodeID, tx.Height)
}

func (m *Module) saveTokenInfo(dbTx *database.DbTx, contract string, codeID uint64, height int64) error {
	tokenInfo, err := m.source.GetTokenInfo(contract, height)
	if err != nil {
		return err
	}

	tokenInfo.Address = contract
	tokenInfo.CodeID = codeID
	if err := dbTx.SaveTokenInfo(tokenInfo); err != nil {
		return err
	}

	if err := dbTx.SaveTokenBalances(contract, tokenInfo.Balances); err != nil {
		return err
	}

	return nil
}

func (m *Module) handleMsgExecuteContract(dbTx *database.DbTx, msg *wasmTypes.MsgExecuteContract, tx *juno.Tx, index int) error {
	if exists, err := dbTx.IsExistingToken(msg.Contract); err != nil {
		return err
	} else if !exists {
		return nil
	}

	req, err := ParseToMsgExecuteToken(msg.Msg)
	if err != nil {
		return err
	}

	switch req.Type {
	case "update_minter":
		return dbTx.UpdateTokenMinter(req.Contract, req.NewMinter)
	case "update_marketing":
		return dbTx.UpdateTokenMarketing(req.Contract, req.Project, req.Description, req.Admin)
	case "upload_logo":
		return dbTx.UpdateTokenLogo(req.Contract, utils.SanitizeUTF8(string(req.MsgRaw)))
	case "mint", "burn", "burn_from":
		if supply, err := m.source.GetCirculatingSupply(req.Contract, tx.Height); err != nil {
			return err
		} else if err := dbTx.UpdateTokenCirculatingSupply(req.Contract, supply); err != nil {
			return err
		}
	}

	req.Contract = msg.Contract
	req.Sender = msg.Sender
	balances, err := m.fetchBalances(req, tx.Height)
	if err != nil {
		return err
	}

	return dbTx.SaveTokenBalances(req.Contract, balances)
}

func (m *Module) fetchBalances(msg *types.MsgExecuteToken, height int64) ([]*types.TokenBalance, error) {
	balances := []*types.TokenBalance{{Address: msg.Sender}}

	if msg.Owner != "" {
		balances = append(balances, &types.TokenBalance{Address: msg.Owner})
	}

	if msg.Recipient != "" {
		balances = append(balances, &types.TokenBalance{Address: msg.Recipient})
	}

	if msg.RecipientContract != "" {
		balances = append(balances, &types.TokenBalance{Address: msg.RecipientContract})
	}

	for _, b := range balances {
		balance, err := m.source.GetBalance(msg.Contract, b.Address, height)
		if err != nil {
			return nil, err
		}

		b.Amount = balance
	}

	return balances, nil
}

func (m *Module) handleMsgMigrateContract(dbTx *database.DbTx, msg *wasmTypes.MsgMigrateContract, tx *juno.Tx, index int) error {
	if exists, err := dbTx.IsExistingToken(msg.Contract); err != nil {
		return err
	} else if !exists {
		return nil
	}

	if err := dbTx.DeleteAllTokenBalances(msg.Contract); err != nil {
		return err
	}

	return m.saveTokenInfo(dbTx, msg.Contract, msg.CodeID, tx.Height)
}
