package marketplace

import (
	"strconv"

	abci "github.com/tendermint/tendermint/abci/types"

	marketplaceTypes "github.com/CudoVentures/cudos-node/x/marketplace/types"
	"github.com/forbole/bdjuno/v2/client/coingecko"
	"github.com/forbole/bdjuno/v2/database"
	juno "github.com/forbole/juno/v2/types"

	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// HandleBlock implements BlockModule
func (m *Module) HandleBlock(
	block *tmctypes.ResultBlock, res *tmctypes.ResultBlockResults, tx []*juno.Tx, _ *tmctypes.ResultValidators,
) error {
	if err := m.handleSoldAuctions(res.EndBlockEvents, uint64(block.Block.Time.Unix())); err != nil {
		return err
	}

	return m.handleAuctionPriceDiscounts(res.EndBlockEvents)
}

func (m *Module) handleSoldAuctions(events []abci.Event, timestamp uint64) error {
	for _, event := range juno.FindEventsByType(events, marketplaceTypes.EventBuyNftType) {
		auctionIDAttr, err := juno.FindAttributeByKey(event, marketplaceTypes.AttributeAuctionID)
		if err != nil {
			continue
		}

		auctionID, err := strconv.ParseUint(string(auctionIDAttr.Value), 10, 64)
		if err != nil {
			continue
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
			return dbTx.SaveMarketplaceAuctionSold(auctionID, timestamp, usdPrice, btcPrice, "")
		}); err != nil {
			return err
		}
	}

	return nil
}

func (m *Module) handleAuctionPriceDiscounts(events []abci.Event) error {
	for _, event := range juno.FindEventsByType(events, marketplaceTypes.EventDutchAuctionPriceDiscountType) {
		auctionIDAttr, err := juno.FindAttributeByKey(event, marketplaceTypes.AttributeAuctionID)
		if err != nil {
			continue
		}

		auctionID, err := strconv.ParseUint(string(auctionIDAttr.Value), 10, 64)
		if err != nil {
			continue
		}

		auctionInfo, err := juno.FindAttributeByKey(event, marketplaceTypes.AttributeAuctionInfo)
		if err != nil {
			continue
		}

		if err := m.db.UpdateMarketplaceAuctionInfo(auctionID, string(auctionInfo.Value)); err != nil {
			return err
		}
	}

	return nil
}
