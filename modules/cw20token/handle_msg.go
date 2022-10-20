package cw20token

import (
	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
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

func (m *Module) handleMsgInstantiateContract(dbTx *database.DbTx, msg *wasm.MsgInstantiateContract, tx *juno.Tx, index int) error {
	if found, err := dbTx.CodeIDExists(msg.CodeID); !found {
		return err
	}

	contract := utils.GetValueFromLogs(uint32(index), tx.Logs, wasm.EventTypeInstantiate, wasm.AttributeKeyContractAddr)

	token, err := m.fetchTokenInfo(dbTx, contract, tx.Height)
	if err != nil {
		return err
	}

	token.CodeID = msg.CodeID

	if err := dbTx.SaveInfo(token); err != nil {
		return err
	}

	return dbTx.SaveBalances(contract, token.Balances)
}

func (m *Module) fetchTokenInfo(dbTx *database.DbTx, contract string, height int64) (*types.TokenInfo, error) {
	res, err := m.source.GetTokenInfo(contract, height)
	if err != nil {
		return nil, err
	}

	tokenInfo, err := parseToTokenInfo(res)
	if err != nil {
		return nil, err
	}

	tokenInfo.Address = contract

	return tokenInfo, nil
}

func (m *Module) handleMsgExecuteContract(dbTx *database.DbTx, msg *wasm.MsgExecuteContract, tx *juno.Tx, index int) error {
	if found, err := dbTx.TokenExists(msg.Contract); err != nil || !found {
		return err
	}

	r, err := parseToMsgExecuteToken(msg)
	if err != nil {
		return err
	}

	switch r.Type {
	case "update_minter":
		return dbTx.UpdateMinter(r.Contract, r.NewMinter)
	case "update_marketing":
		return dbTx.UpdateMarketing(r.Contract, types.NewMarketingInfo(r.Project, r.Description, r.Admin))
	case "upload_logo":
		return dbTx.UpdateLogo(r.Contract, utils.SanitizeUTF8(string(r.MsgRaw)))
	case "mint", "burn", "burn_from":
		res, err := m.source.GetCirculatingSupply(r.Contract, tx.Height)
		if err != nil {
			return err
		}

		supply, err := parseToCirculatingSupply(res)
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
		res, err := m.source.GetBalance(msg.Contract, b.Address, height)
		if err != nil {
			return nil, err
		}

		balance, err := parseToBalance(res)
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
