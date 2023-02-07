package marketplace

import (
	"strconv"
	"testing"
	"time"

	"github.com/CudoVentures/cudos-node/simapp"
	"github.com/CudoVentures/cudos-node/x/marketplace/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	dbutils "github.com/forbole/bdjuno/v2/database/utils"
	"github.com/forbole/bdjuno/v2/utils"
	juno "github.com/forbole/juno/v2/types"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	addr1              = "cudos1"
	addr2              = "cudos2"
	denomID            = "denom"
	tokenID            = uint64(1)
	auctionID          = uint64(1)
	txHash             = "txHash"
	txHash2            = "txHash2"
	auctionInfo        = `{test:test}`
	auctionInfoUpdated = `{updated: updated}`
)

var (
	startTime = time.Date(2222, 1, 1, 1, 0, 0, 0, time.FixedZone("", 0))
	endTime   = startTime.Add(time.Hour * 25)
)

func TestMarketplaceAuctions(t *testing.T) {
	// SETUP
	db, err := utils.NewTestDb("marketplaceAuctionsTest")
	require.NoError(t, err)

	_, err = db.Sql.Exec(`INSERT INTO block (height, hash, timestamp) VALUES ($1, $2, $3)`, 1, "1", time.Now())
	require.NoError(t, err)

	_, err = db.Sql.Exec(`INSERT INTO transaction (hash, height, success, signatures) VALUES ($1, $2, true, $3)`, txHash, 1, pq.Array([]string{"1"}))
	require.NoError(t, err)

	_, err = db.Sql.Exec(`INSERT INTO transaction (hash, height, success, signatures) VALUES ($1, $2, true, $3)`, txHash2, 1, pq.Array([]string{"1"}))
	require.NoError(t, err)

	_, err = db.Sql.Exec(`INSERT INTO nft_denom (transaction_hash, id, name, schema, symbol, owner, contract_address_signer, traits, minter, description, data_text, data_json) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) ON CONFLICT DO NOTHING`, txHash, denomID, "name", "schema", "symbol", addr1, "contractAddressSigner", "traits", "minter", "description", "dataText", "{}")
	require.NoError(t, err)

	_, err = db.Sql.Exec(`INSERT INTO nft_nft (transaction_hash, id, denom_id, name, uri, owner, data_json, data_text, sender, contract_address_signer, uniq_id) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) ON CONFLICT DO NOTHING`, txHash, tokenID, denomID, "name", "uri", addr1, "{}", "dataText", addr1, "contractAddressSigner", dbutils.FormatUniqID(tokenID, denomID))
	require.NoError(t, err)

	m := NewModule(simapp.MakeTestEncodingConfig().Marshaler, db)
	txb := utils.NewMockTxBuilder(t, startTime, txHash, 1)

	// TEST msgHandlePublishAuction
	txb.WithEventPublishAuction(auctionID, startTime, endTime, auctionInfo)

	err = m.handleMsgPublishAuction(0, txb.Build(), &types.MsgPublishAuction{
		TokenId: strconv.FormatUint(tokenID, 10),
		DenomId: denomID,
		Creator: addr1,
	})
	require.NoError(t, err)

	wantAuction := dbtypes.MarketplaceAuctionRow{
		ID:        auctionID,
		TokenID:   tokenID,
		DenomID:   denomID,
		Creator:   addr1,
		StartTime: startTime,
		EndTime:   endTime,
		Auction:   auctionInfo,
		Sold:      false,
	}

	var haveAuctions []dbtypes.MarketplaceAuctionRow
	err = db.Sqlx.Select(&haveAuctions, `SELECT * FROM marketplace_auction WHERE id = $1`, auctionID)
	require.NoError(t, err)
	require.Len(t, haveAuctions, 1)
	require.Equal(t, wantAuction, haveAuctions[0])

	// TEST msgHandlePlaceBid
	msgHandleBid := &types.MsgPlaceBid{
		AuctionId: auctionID,
		Amount:    sdk.NewCoin("acudos", sdk.OneInt()),
		Bidder:    addr2,
	}

	err = m.handleMsgPlaceBid(0, txb.Build(), msgHandleBid)
	require.NoError(t, err)

	var haveHistoryRowCount int
	err = db.Sqlx.QueryRow(`SELECT COUNT(*) FROM marketplace_nft_buy_history`).Scan(&haveHistoryRowCount)
	require.NoError(t, err)
	require.Zero(t, haveHistoryRowCount)

	wantBid := dbtypes.MarketplaceBidRow{
		AuctionID: auctionID,
		Bidder:    addr2,
		Price:     "1",
		Timestamp: startTime,
		TxHash:    txHash,
	}

	var haveBids []dbtypes.MarketplaceBidRow
	err = db.Sqlx.Select(&haveBids, `SELECT * FROM marketplace_bid`)
	require.NoError(t, err)
	require.Len(t, haveBids, 1)
	require.Equal(t, wantBid, haveBids[0])

	// TEST msgHandleBid with EventBuyNft
	txb.WithEventBuyNftFromAuction(auctionID, tokenID, denomID, addr2)
	err = m.handleMsgPlaceBid(0, txb.Build(), msgHandleBid)
	require.NoError(t, err)

	wantTimestamp, err := utils.ISO8601ToTimestamp(txb.Build().Timestamp)
	require.NoError(t, err)

	wantBuyHistory := dbtypes.MarketplaceNftBuyHistory{
		TxHash:    txHash,
		TokenID:   tokenID,
		DenomID:   denomID,
		Price:     "1",
		Buyer:     addr2,
		Seller:    addr1,
		Timestamp: uint64(wantTimestamp),
		UniqID:    dbutils.FormatUniqID(tokenID, denomID),
	}

	var haveBuyHistory []dbtypes.MarketplaceNftBuyHistory
	err = db.Sqlx.Select(&haveBuyHistory, `SELECT * FROM marketplace_nft_buy_history`)
	require.NoError(t, err)
	require.Len(t, haveBuyHistory, 1)
	haveBuyHistory[0].UsdPrice = ""
	haveBuyHistory[0].BtcPrice = ""
	require.Equal(t, wantBuyHistory, haveBuyHistory[0])

	wantTransferHistory := dbtypes.NftTransferHistoryRow{
		ID:        1,
		TxHash:    txHash,
		DenomID:   denomID,
		OldOwner:  addr1,
		NewOwner:  addr2,
		Timestamp: uint64(wantTimestamp),
		UniqID:    dbutils.FormatUniqID(tokenID, denomID),
	}

	var haveTransferHistory []dbtypes.NftTransferHistoryRow
	err = db.Sqlx.Select(&haveTransferHistory, `SELECT * FROM nft_transfer_history`)
	require.NoError(t, err)
	require.Len(t, haveTransferHistory, 1)
	require.Equal(t, wantTransferHistory, haveTransferHistory[0])

	var haveSold bool
	err = db.Sqlx.QueryRow(`SELECT sold FROM marketplace_auction`).Scan(&haveSold)
	require.NoError(t, err)
	require.True(t, haveSold)

	var haveBidRowCount int
	err = db.Sqlx.QueryRow(`SELECT COUNT(*) FROM marketplace_bid`).Scan(&haveBidRowCount)
	require.NoError(t, err)
	require.Equal(t, 2, haveBidRowCount)

	// TEST msgHandleAcceptBid
	txb = utils.NewMockTxBuilder(t, startTime, txHash2, 1)
	err = m.handleMsgAcceptBid(txb.Build(), &types.MsgAcceptBid{AuctionId: auctionID})
	require.NoError(t, err)

	haveBuyHistory = []dbtypes.MarketplaceNftBuyHistory{}
	err = db.Sqlx.Select(&haveBuyHistory, `SELECT * FROM marketplace_nft_buy_history`)
	require.NoError(t, err)
	require.Len(t, haveBuyHistory, 2)
	require.Equal(t, txHash2, haveBuyHistory[1].TxHash)

	haveTransferHistory = []dbtypes.NftTransferHistoryRow{}

	err = db.Sqlx.Select(&haveTransferHistory, `SELECT * FROM nft_transfer_history`)
	require.NoError(t, err)
	require.Len(t, haveTransferHistory, 2)
	require.Equal(t, txHash2, haveTransferHistory[1].TxHash)

	// TEST HandleBlock
	block := tmctypes.ResultBlock{Block: &tmtypes.Block{Header: tmtypes.Header{Time: startTime}}}
	blockEvents := tmctypes.ResultBlockResults{EndBlockEvents: []abcitypes.Event{
		abcitypes.Event(sdk.NewEvent(
			types.EventBuyNftType,
			sdk.NewAttribute(types.AttributeAuctionID, strconv.FormatUint(auctionID, 10)),
			sdk.NewAttribute(types.AttributeKeyTokenID, strconv.FormatUint(tokenID, 10)),
			sdk.NewAttribute(types.AttributeKeyDenomID, denomID),
			sdk.NewAttribute(types.AttributeKeyBuyer, addr2),
		)),
		abcitypes.Event(sdk.NewEvent(
			types.EventDutchAuctionPriceDiscountType,
			sdk.NewAttribute(types.AttributeAuctionID, strconv.FormatUint(auctionID, 10)),
			sdk.NewAttribute(types.AttributeAuctionInfo, auctionInfoUpdated),
		)),
	}}

	err = m.HandleBlock(&block, &blockEvents, []*juno.Tx{}, &tmctypes.ResultValidators{})
	require.NoError(t, err)

	var haveBuyHistoryRowCount int
	err = db.Sqlx.QueryRow(`SELECT COUNT(*) FROM marketplace_nft_buy_history`).Scan(&haveBuyHistoryRowCount)
	require.NoError(t, err)
	require.Equal(t, 3, haveBuyHistoryRowCount)

	var haveTransferHistoryRowCount int
	err = db.Sqlx.QueryRow(`SELECT COUNT(*) FROM nft_transfer_history`).Scan(&haveTransferHistoryRowCount)
	require.NoError(t, err)
	require.Equal(t, 3, haveBuyHistoryRowCount)

	var haveAuctionInfo string
	err = db.Sqlx.QueryRow(`SELECT auction FROM marketplace_auction WHERE id = $1`, auctionID).Scan(&haveAuctionInfo)
	require.NoError(t, err)
	require.Equal(t, auctionInfoUpdated, haveAuctionInfo)
}
