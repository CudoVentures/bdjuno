package nft

import (
	"fmt"
	"strconv"

	nftTypes "github.com/CudoVentures/cudos-node/x/nft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v2/modules/utils"
	juno "github.com/forbole/juno/v2/types"
)

// HandleMsg implements MessageModule
func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	if len(tx.Logs) == 0 {
		return nil
	}

	switch msg.(type) {
	case *nftTypes.MsgMintNFT:
		return m.handleMsgMintNFT(index, tx)
	case *nftTypes.MsgIssueDenom:
		return m.handleMsgIssueDenom(index, tx)
	default:
		return nil
	}
}

func (m *Module) handleMsgIssueDenom(index int, tx *juno.Tx) error {
	denomID := utils.GetValueFromLogs(uint32(index), tx.Logs, nftTypes.EventTypeIssueDenom, nftTypes.AttributeKeyDenomID)
	if denomID == "" {
		return fmt.Errorf("denom id not found in tx %s", tx.TxHash)
	}

	return m.db.SaveMsgIssueDenom(tx.TxHash, denomID)
}

func (m *Module) handleMsgMintNFT(index int, tx *juno.Tx) error {
	tokenIDStr := utils.GetValueFromLogs(uint32(index), tx.Logs, nftTypes.EventTypeMintNFT, nftTypes.AttributeKeyTokenID)
	if tokenIDStr == "" {
		return fmt.Errorf("token id not found in tx %s", tx.TxHash)
	}

	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		return err
	}

	denomID := utils.GetValueFromLogs(uint32(index), tx.Logs, nftTypes.EventTypeMintNFT, nftTypes.AttributeKeyDenomID)
	if denomID == "" {
		return fmt.Errorf("denom id not found in tx %s", tx.TxHash)
	}

	return m.db.SaveMsgMintNFT(tx.TxHash, tokenID, denomID)
}
