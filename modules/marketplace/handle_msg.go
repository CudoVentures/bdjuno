package marketplace

import (
	"fmt"
	"strconv"

	marketplaceTypes "github.com/CudoVentures/cudos-node/x/marketplace/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v2/client/coingecko"
	"github.com/forbole/bdjuno/v2/database"
	utils "github.com/forbole/bdjuno/v2/modules/utils"
	generalUtils "github.com/forbole/bdjuno/v2/utils"
	juno "github.com/forbole/juno/v2/types"
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
		return m.handleMsgBuyNft(tx, cosmosMsg)
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
	collectionID, err := utils.GetUint64FromLogs(index, tx.Logs, tx.TxHash, marketplaceTypes.EventPublishCollectionType, marketplaceTypes.AttributeKeyCollectionID)
	if err != nil {
		return err
	}
	return m.db.SaveMarketplaceCollection(tx.TxHash, collectionID, msg.DenomId, royaltiesToText(msg.MintRoyalties), royaltiesToText(msg.ResaleRoyalties), msg.Creator, false)
}

func (m *Module) handleMsgPublishNft(index int, tx *juno.Tx, msg *marketplaceTypes.MsgPublishNft) error {
	nftID, err := utils.GetUint64FromLogs(index, tx.Logs, tx.TxHash, marketplaceTypes.EventPublishNftType, marketplaceTypes.AttributeKeyNftID)
	if err != nil {
		return err
	}

	tokenID, err := strconv.ParseUint(msg.TokenId, 10, 64)
	if err != nil {
		return err
	}

	return m.db.SaveMarketplaceNft(tx.TxHash, nftID, tokenID, msg.DenomId, "", msg.Price.Amount.String(), msg.Creator)
}

func (m *Module) handleMsgMintNft(index int, tx *juno.Tx, msg *marketplaceTypes.MsgMintNft) error {
	tokenIDStr := utils.GetValueFromLogs(uint32(index), tx.Logs, marketplaceTypes.EventMintNftType, marketplaceTypes.AttributeKeyNftID)
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

	if err := m.db.SaveNFT(tx.TxHash, tokenID, msg.DenomId, msg.Name, msg.Uri, utils.SanitizeUTF8(dataJSON), dataText, msg.Recipient, msg.Creator, ""); err != nil {
		return err
	}

	usdPrice, err := coingecko.GetCUDOSPrice("usd")
	if err != nil {
		return err
	}

	btcPrice, err := coingecko.GetCUDOSPrice("btc")
	if err != nil {
		return err
	}

	if err := m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		return dbTx.SaveMarketplaceNftMint(tx.TxHash, tokenID, msg.Recipient, msg.DenomId, msg.Price.Amount.String(), uint64(timestamp), usdPrice, btcPrice)
	}); err != nil {
		return err
	}

	if err := m.db.SaveMarketplaceNft(tx.TxHash, 0, tokenID, msg.DenomId, msg.Uid, msg.Price.Amount.String(), msg.Recipient); err != nil {
		return err
	}

	return m.db.UpdateNFTHistory(tx.TxHash, tokenID, msg.DenomId, "0x0", msg.Recipient, uint64(timestamp))
}

func (m *Module) handleMsgBuyNft(tx *juno.Tx, msg *marketplaceTypes.MsgBuyNft) error {
	timestamp, err := generalUtils.ISO8601ToTimestamp(tx.Timestamp)
	if err != nil {
		return err
	}

	usdPrice, err := coingecko.GetCUDOSPrice("usd")
	if err != nil {
		return err
	}

	btcPrice, err := coingecko.GetCUDOSPrice("btc")
	if err != nil {
		return err
	}

	return m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		if err := dbTx.SaveMarketplaceNftBuy(tx.TxHash, msg.Id, msg.Creator, uint64(timestamp), usdPrice, btcPrice); err != nil {
			return err
		}

		return dbTx.SetMarketplaceNFTPrice(msg.Id, "0")
	})
}

func (m *Module) handleMsgRemoveNft(msg *marketplaceTypes.MsgRemoveNft) error {
	return m.db.SetMarketplaceNFTPrice(msg.Id, "0")
}

func (m *Module) handleMsgVerifyCollection(msg *marketplaceTypes.MsgVerifyCollection) error {
	return m.db.SetMarketplaceCollectionVerificationStatus(msg.Id, true)
}

func (m *Module) handleMsgUnverifyCollection(msg *marketplaceTypes.MsgUnverifyCollection) error {
	return m.db.SetMarketplaceCollectionVerificationStatus(msg.Id, false)
}

func (m *Module) handleMsgUpdatePrice(msg *marketplaceTypes.MsgUpdatePrice) error {
	return m.db.SetMarketplaceNFTPrice(msg.Id, msg.Price.Amount.String())
}

func (m *Module) handleMsgUpdateRoyalties(msg *marketplaceTypes.MsgUpdateRoyalties) error {
	return m.db.SetMarketplaceCollectionRoyalties(msg.Id, royaltiesToText(msg.MintRoyalties), royaltiesToText(msg.ResaleRoyalties))
}

func (m *Module) handleMsgCreateCollection(index int, tx *juno.Tx, msg *marketplaceTypes.MsgCreateCollection) error {
	collectionID, err := utils.GetUint64FromLogs(index, tx.Logs, tx.TxHash, marketplaceTypes.EventCreateCollectionType, marketplaceTypes.AttributeKeyCollectionID)
	if err != nil {
		return err
	}

	return m.db.SaveMarketplaceCollection(tx.TxHash, collectionID, msg.Name, royaltiesToText(msg.MintRoyalties), royaltiesToText(msg.ResaleRoyalties), msg.Creator, msg.Verified)
}

func royaltiesToText(royalties []marketplaceTypes.Royalty) string {
	text := ""
	for _, royalty := range royalties {
		text += royalty.String()
	}
	return text
}
