package nft

import (
	"fmt"
	"strconv"

	marketplaceTypes "github.com/CudoVentures/cudos-node/x/marketplace/types"
	nftTypes "github.com/CudoVentures/cudos-node/x/nft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v2/database"
	utils "github.com/forbole/bdjuno/v2/modules/utils"
	generalUtils "github.com/forbole/bdjuno/v2/utils"
	juno "github.com/forbole/juno/v2/types"
	"github.com/rs/zerolog/log"
)

// HandleMsg implements MessageModule
func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	if tx.Code != 0 {
		return nil
	}

	switch cosmosMsg := msg.(type) {
	case *nftTypes.MsgIssueDenom:
		return m.handleMsgIssueDenom(tx, cosmosMsg)
	case *nftTypes.MsgTransferDenom:
		return m.handleMsgTransferDenom(cosmosMsg)
	case *nftTypes.MsgMintNFT:
		return m.handleMsgMintNFT(index, tx, cosmosMsg)
	case *nftTypes.MsgEditNFT:
		return m.handleMsgEditNFT(cosmosMsg)
	case *nftTypes.MsgTransferNft:
		return m.handleMsgTransferNFT(tx, cosmosMsg)
	case *nftTypes.MsgBurnNFT:
		return m.handleMsgBurnNFT(index, tx, cosmosMsg)
	case *marketplaceTypes.MsgCreateCollection:
		return m.handleMsgCreateCollection(tx, cosmosMsg)
	case *marketplaceTypes.MsgBuyNft:
		return m.handleMsgBuyNft(index, tx, cosmosMsg)
	default:
		return nil
	}
}

func (m *Module) handleMsgIssueDenom(tx *juno.Tx, msg *nftTypes.MsgIssueDenom) error {
	log.Debug().Str("module", "nft").Str("denomId", msg.Id).Msg("handling message issue denom")

	dataJSON, dataText := utils.GetData(msg.Data)

	return m.db.SaveDenom(tx.TxHash, msg.Id, msg.Name, msg.Schema, msg.Symbol, msg.Sender, msg.ContractAddressSigner,
		msg.Traits, msg.Minter, msg.Description, dataText, utils.SanitizeUTF8(dataJSON))
}

func (m *Module) handleMsgTransferDenom(msg *nftTypes.MsgTransferDenom) error {
	log.Debug().Str("module", "nft").Str("denomId", msg.Id).Msg("handling message transfer denom")

	return m.db.UpdateDenom(msg.Id, msg.Recipient)
}

func (m *Module) handleMsgMintNFT(index int, tx *juno.Tx, msg *nftTypes.MsgMintNFT) error {
	log.Debug().Str("module", "nft").Str("denomId", msg.DenomId).Str("name", msg.Name).Msg("handling message mint nft")

	tokenIDStr := utils.GetValueFromLogs(uint32(index), tx.Logs, nftTypes.EventTypeMintNFT, nftTypes.AttributeKeyTokenID)
	if tokenIDStr == "" {
		return fmt.Errorf("token id not found in tx %s", tx.TxHash)
	}

	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		return err
	}

	timestamp, err := generalUtils.ISO8601ToTimestamp(tx.Timestamp)
	if err != nil {
		return err
	}

	dataJSON, dataText := utils.GetData(msg.Data)

	return m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		if error := dbTx.UpdateNFTHistory(tx.TxHash, tokenID, msg.DenomId, "0x0", msg.Sender, uint64(timestamp)); error != nil {
			return error
		}

		return dbTx.SaveNFT(tx.TxHash, tokenID, msg.DenomId, msg.Name, msg.URI, utils.SanitizeUTF8(dataJSON), dataText, msg.Recipient, msg.Sender, msg.ContractAddressSigner)
	})
}

func (m *Module) handleMsgEditNFT(msg *nftTypes.MsgEditNFT) error {
	log.Debug().Str("module", "nft").Str("denomId", msg.DenomId).Str("tokenId", msg.Id).Msg("handling message edit nft")

	dataJSON, dataText := utils.GetData(msg.Data)

	return m.db.UpdateNFT(msg.Id, msg.DenomId, msg.Name, msg.URI, utils.SanitizeUTF8(dataJSON), dataText)
}

func (m *Module) handleMsgTransferNFT(tx *juno.Tx, msg *nftTypes.MsgTransferNft) error {
	log.Debug().Str("module", "nft").Str("denomId", msg.DenomId).Str("tokenId", msg.TokenId).Str("from", msg.From).Str("to", msg.To).Msg("handling message transfer nft")

	timestamp, err := generalUtils.ISO8601ToTimestamp(tx.Timestamp)
	if err != nil {
		return err
	}

	tokenID, err := strconv.ParseUint(msg.TokenId, 10, 64)
	if err != nil {
		return err
	}

	return m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		if error := dbTx.UpdateNFTHistory(tx.TxHash, tokenID, msg.DenomId, msg.Sender, msg.To, uint64(timestamp)); error != nil {
			return error
		}

		return dbTx.UpdateNFTOwner(msg.TokenId, msg.DenomId, msg.To)
	})
}

func (m *Module) handleMsgBurnNFT(index int, tx *juno.Tx, msg *nftTypes.MsgBurnNFT) error {
	log.Debug().Str("module", "nft").Str("denomId", msg.DenomId).Str("tokenId", msg.Id).Str("to", msg.To).Msg("handling message burn nft")

	tokenIDStr := utils.GetValueFromLogs(uint32(index), tx.Logs, nftTypes.EventTypeBurnNFT, nftTypes.AttributeKeyTokenID)
	if tokenIDStr == "" {
		return fmt.Errorf("token id not found in tx %s", tx.TxHash)
	}

	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		return err
	}

	timestamp, err := generalUtils.ISO8601ToTimestamp(tx.Timestamp)
	if err != nil {
		return err
	}

	return m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		if error := dbTx.UpdateNFTHistory(tx.TxHash, tokenID, msg.DenomId, msg.Sender, "0x0", uint64(timestamp)); error != nil {
			return error
		}

		return dbTx.BurnNFT(msg.Id, msg.DenomId)
	})
}

func (m *Module) handleMsgCreateCollection(tx *juno.Tx, msg *marketplaceTypes.MsgCreateCollection) error {
	log.Debug().Str("module", "nft").Str("denomId", msg.Id).Msg("handling message create collection")

	dataJSON, dataText := utils.GetData(msg.Data)

	return m.db.SaveDenom(tx.TxHash, msg.Id, msg.Name, msg.Schema, msg.Symbol, msg.Creator, "",
		msg.Traits, msg.Minter, msg.Description, dataText, utils.SanitizeUTF8(dataJSON))
}

func (m *Module) handleMsgBuyNft(index int, tx *juno.Tx, msg *marketplaceTypes.MsgBuyNft) error {
	log.Debug().Str("module", "nft").Uint64("Id", msg.Id).Str("msg creator", msg.Creator).Msg("handling message buy nft")

	tokenID := utils.GetValueFromLogs(uint32(index), tx.Logs, marketplaceTypes.EventBuyNftType, marketplaceTypes.AttributeKeyTokenID)
	if tokenID == "" {
		return fmt.Errorf("token id not found in tx %s", tx.TxHash)
	}

	denomID := utils.GetValueFromLogs(uint32(index), tx.Logs, marketplaceTypes.EventBuyNftType, marketplaceTypes.AttributeKeyDenomID)
	if denomID == "" {
		return fmt.Errorf("denom id not found in tx %s", tx.TxHash)
	}

	return m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		return dbTx.UpdateNFTOwner(tokenID, denomID, msg.Creator)
	})
}
