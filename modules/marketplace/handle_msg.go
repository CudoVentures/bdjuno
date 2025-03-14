package marketplace

import (
	"encoding/json"
	"fmt"
	"strconv"

	marketplaceTypes "github.com/CudoVentures/cudos-node/x/marketplace/types"
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
	case *marketplaceTypes.MsgPublishCollection:
		return m.handleMsgPublishCollection(index, tx, cosmosMsg)
	case *marketplaceTypes.MsgPublishNft:
		return m.handleMsgPublishNft(index, tx, cosmosMsg)
	case *marketplaceTypes.MsgMintNft:
		return m.handleMsgMintNft(index, tx, cosmosMsg)
	case *marketplaceTypes.MsgBuyNft:
		return m.handleMsgBuyNft(index, tx, cosmosMsg)
	case *marketplaceTypes.MsgRemoveNft:
		return m.handleMsgRemoveNft(cosmosMsg)
	case *marketplaceTypes.MsgVerifyCollection:
		return m.handleMsgVerifyCollection(cosmosMsg)
	case *marketplaceTypes.MsgUnverifyCollection:
		return m.handleMsgUnverifyCollection(cosmosMsg)
	case *marketplaceTypes.MsgUpdatePrice:
		return m.handleMsgUpdatePrice(cosmosMsg)
	case *marketplaceTypes.MsgUpdateRoyalties:
		return m.handleMsgUpdateRoyalties(cosmosMsg)
	case *marketplaceTypes.MsgCreateCollection:
		return m.handleMsgCreateCollection(index, tx, cosmosMsg)
	default:
		return nil
	}
}

func (m *Module) handleMsgPublishCollection(index int, tx *juno.Tx, msg *marketplaceTypes.MsgPublishCollection) error {
	log.Debug().Str("module", "marketplace").Msg("handling message publish collection")

	collectionID, err := utils.GetUint64FromLogs(index, tx.Logs, tx.TxHash, marketplaceTypes.EventPublishCollectionType, marketplaceTypes.AttributeKeyCollectionID)
	if err != nil {
		return err
	}
	log.Debug().Str("module", "marketplace").Uint64("collectionId", collectionID)

	mintRoyaltiesJSON, err := json.Marshal(msg.MintRoyalties)
	if err != nil {
		return err
	}

	resaleRoyaltiesJSON, err := json.Marshal(msg.ResaleRoyalties)
	if err != nil {
		return err
	}

	return m.db.SaveMarketplaceCollection(tx.TxHash, collectionID, msg.DenomId, string(mintRoyaltiesJSON), string(resaleRoyaltiesJSON), msg.Creator, false)
}

func (m *Module) handleMsgPublishNft(index int, tx *juno.Tx, msg *marketplaceTypes.MsgPublishNft) error {
	log.Debug().Str("module", "marketplace").Msg("handling message publish nft")

	nftID, err := utils.GetUint64FromLogs(index, tx.Logs, tx.TxHash, marketplaceTypes.EventPublishNftType, marketplaceTypes.AttributeKeyNftID)
	if err != nil {
		return err
	}
	log.Debug().Str("module", "marketplace").Uint64("nftId", nftID)

	tokenID, err := strconv.ParseUint(msg.TokenId, 10, 64)
	if err != nil {
		return err
	}

	return m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		err := m.db.CheckIfNftExists(tokenID, msg.DenomId)

		if err != nil {
			return err
		}

		return dbTx.ListNft(tx.TxHash, nftID, tokenID, msg.DenomId, msg.Price.Amount.String())
	})
}

func (m *Module) handleMsgMintNft(index int, tx *juno.Tx, msg *marketplaceTypes.MsgMintNft) error {
	log.Debug().Str("module", "marketplace").Msg("handling message mint nft")

	tokenIDStr := utils.GetValueFromLogs(uint32(index), tx.Logs, marketplaceTypes.EventMintNftType, marketplaceTypes.AttributeKeyTokenID)
	if tokenIDStr == "" {
		return fmt.Errorf("token id not found in tx %s", tx.TxHash)
	}
	log.Debug().Str("module", "marketplace").Str("tokenId", tokenIDStr)

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
		if err := dbTx.SaveNFT(tx.TxHash, tokenID, msg.DenomId, msg.Name, msg.Uri, utils.SanitizeUTF8(dataJSON), dataText, msg.Recipient, msg.Creator, ""); err != nil {
			return err
		}

		if err := dbTx.SaveMarketplaceNftMint(tx.TxHash, tokenID, msg.Recipient, msg.DenomId, msg.Price.Amount.String(), uint64(timestamp), m.cudosPrice.USD, m.cudosPrice.BTC); err != nil {
			return err
		}

		if err := dbTx.SaveMarketplaceNft(tx.TxHash, tokenID, msg.DenomId, msg.Uid, "0", msg.Recipient); err != nil {
			return err
		}

		return dbTx.UpdateNFTHistory(tx.TxHash, tokenID, msg.DenomId, "0x0", msg.Recipient, uint64(timestamp))
	})
}

func (m *Module) handleMsgBuyNft(index int, tx *juno.Tx, msg *marketplaceTypes.MsgBuyNft) error {
	log.Debug().Str("module", "marketplace").Msg("handling message buy nft")

	timestamp, err := generalUtils.ISO8601ToTimestamp(tx.Timestamp)
	if err != nil {
		return err
	}

	tokenIDStr := utils.GetValueFromLogs(uint32(index), tx.Logs, marketplaceTypes.EventBuyNftType, marketplaceTypes.AttributeKeyTokenID)
	log.Debug().Str("module", "marketplace").Str("tokenId", tokenIDStr)

	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		return err
	}

	denomIDStr := utils.GetValueFromLogs(uint32(index), tx.Logs, marketplaceTypes.EventBuyNftType, marketplaceTypes.AttributeKeyDenomID)
	fromOwner := utils.GetValueFromLogs(uint32(index), tx.Logs, marketplaceTypes.EventBuyNftType, marketplaceTypes.AttributeKeyOwner)

	return m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		if err := dbTx.SaveMarketplaceNftBuy(tx.TxHash, msg.Id, msg.Creator, uint64(timestamp), m.cudosPrice.USD, m.cudosPrice.BTC); err != nil {
			return err
		}

		if err := dbTx.UpdateNFTHistory(tx.TxHash, tokenID, denomIDStr, fromOwner, msg.Creator, uint64(timestamp)); err != nil {
			return err
		}

		return dbTx.UnlistNft(msg.Id)
	})
}

func (m *Module) handleMsgRemoveNft(msg *marketplaceTypes.MsgRemoveNft) error {
	log.Debug().Str("module", "marketplace").Uint64("ID", msg.Id).Msg("handling message remove nft")

	return m.db.UnlistNft(msg.Id)
}

func (m *Module) handleMsgVerifyCollection(msg *marketplaceTypes.MsgVerifyCollection) error {
	log.Debug().Str("module", "marketplace").Uint64("ID", msg.Id).Msg("handling message verify collection")

	return m.db.SetMarketplaceCollectionVerificationStatus(msg.Id, true)
}

func (m *Module) handleMsgUnverifyCollection(msg *marketplaceTypes.MsgUnverifyCollection) error {
	log.Debug().Str("module", "marketplace").Uint64("ID", msg.Id).Msg("handling message unverify collection")

	return m.db.SetMarketplaceCollectionVerificationStatus(msg.Id, false)
}

func (m *Module) handleMsgUpdatePrice(msg *marketplaceTypes.MsgUpdatePrice) error {
	log.Debug().Str("module", "marketplace").Uint64("ID", msg.Id).Str("Price", msg.Price.Amount.String()).Msg("handling message update price")

	return m.db.SetMarketplaceNFTPrice(msg.Id, msg.Price.Amount.String())
}

func (m *Module) handleMsgUpdateRoyalties(msg *marketplaceTypes.MsgUpdateRoyalties) error {
	log.Debug().Str("module", "marketplace").Uint64("ID", msg.Id).Str("Mint Royalties", fmt.Sprintf("%s", msg.MintRoyalties)).Str("Resale Royalties", fmt.Sprintf("%s", &msg.ResaleRoyalties)).Msg("handling message update royalties")

	mintRoyaltiesJSON, err := json.Marshal(msg.MintRoyalties)
	if err != nil {
		return err
	}

	resaleRoyaltiesJSON, err := json.Marshal(msg.ResaleRoyalties)
	if err != nil {
		return err
	}

	return m.db.SetMarketplaceCollectionRoyalties(msg.Id, string(mintRoyaltiesJSON), string(resaleRoyaltiesJSON))
}

func (m *Module) handleMsgCreateCollection(index int, tx *juno.Tx, msg *marketplaceTypes.MsgCreateCollection) error {
	log.Debug().Str("module", "marketplace").Msg("handling message create collection")

	collectionID, err := utils.GetUint64FromLogs(index, tx.Logs, tx.TxHash, marketplaceTypes.EventCreateCollectionType, marketplaceTypes.AttributeKeyCollectionID)
	if err != nil {
		return err
	}
	log.Debug().Str("module", "marketplace").Uint64("ID", collectionID)

	mintRoyaltiesJSON, err := json.Marshal(msg.MintRoyalties)
	if err != nil {
		return err
	}

	resaleRoyaltiesJSON, err := json.Marshal(msg.ResaleRoyalties)
	if err != nil {
		return err
	}

	return m.db.SaveMarketplaceCollection(tx.TxHash, collectionID, msg.Id, string(mintRoyaltiesJSON), string(resaleRoyaltiesJSON), msg.Creator, msg.Verified)
}
