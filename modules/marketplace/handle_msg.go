package marketplace

import (
	"strconv"

	marketplaceTypes "github.com/CudoVentures/cudos-node/x/marketplace/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v2/modules/utils"
	generalUtils "github.com/forbole/bdjuno/v2/utils"
	juno "github.com/forbole/juno/v3/types"
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
		return m.handleMsgMintNft(tx, cosmosMsg)
	case *marketplaceTypes.MsgBuyNft:
		return m.handleMsgBuyNft(tx, cosmosMsg)
	case *marketplaceTypes.MsgRemoveNft:
		return m.handleMsgRemoveNft(tx, cosmosMsg)
	default:
		return nil
	}
}

func (m *Module) handleMsgPublishCollection(index int, tx *juno.Tx, msg *marketplaceTypes.MsgPublishCollection) error {
	collectionID, err := utils.GetUint64FromLogs(index, tx.Logs, tx.TxHash, marketplaceTypes.EventPublishCollectionType, marketplaceTypes.AttributeKeyCollectionID)
	if err != nil {
		return err
	}
	return m.db.SaveMarketplaceCollection(tx.TxHash, collectionID, msg.DenomId, royaltiesToText(msg.MintRoyalties), royaltiesToText(msg.ResaleRoyalties), msg.Creator)
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

	return m.db.SaveMarketplaceNft(tx.TxHash, nftID, tokenID, msg.Price.String(), msg.DenomId, msg.Creator)
}

func (m *Module) handleMsgMintNft(tx *juno.Tx, msg *marketplaceTypes.MsgMintNft) error {
	// We have nothing to do here for now
	return nil
}

func (m *Module) handleMsgBuyNft(tx *juno.Tx, msg *marketplaceTypes.MsgBuyNft) error {
	if err := m.db.RemoveMarketplaceNft(msg.Id); err != nil {
		return err
	}

	timestamp, err := generalUtils.ISO8601ToTimestamp(tx.Timestamp)
	if err != nil {
		return err
	}

	return m.db.SaveMarketplaceNftBuy(tx.TxHash, msg.Id, msg.Creator, uint64(timestamp))
}

func (m *Module) handleMsgRemoveNft(tx *juno.Tx, msg *marketplaceTypes.MsgRemoveNft) error {
	return m.db.RemoveMarketplaceNft(msg.Id)
}

func royaltiesToText(royalties []marketplaceTypes.Royalty) string {
	text := ""
	for _, royalty := range royalties {
		text += royalty.String()
	}
	return text
}
