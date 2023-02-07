package types

import (
	"time"
)

type MarketplaceAuctionRow struct {
	ID        uint64    `db:"id"`
	TokenID   uint64    `db:"token_id"`
	DenomID   string    `db:"denom_id"`
	Creator   string    `db:"creator"`
	StartTime time.Time `db:"start_time"`
	EndTime   time.Time `db:"end_time"`
	Auction   string    `db:"auction"`
	Sold      bool      `db:"sold"`
}

type MarketplaceBidRow struct {
	AuctionID uint64    `db:"auction_id"`
	Bidder    string    `db:"bidder"`
	Price     string    `db:"price"`
	Timestamp time.Time `db:"timestamp"`
	TxHash    string    `db:"transaction_hash"`
}

type MarketplaceNftBuyHistory struct {
	TxHash    string `db:"transaction_hash"`
	TokenID   uint64 `db:"token_id"`
	DenomID   string `db:"denom_id"`
	Price     string `db:"price"`
	Buyer     string `db:"buyer"`
	Seller    string `db:"seller"`
	UsdPrice  string `db:"usd_price"`
	BtcPrice  string `db:"btc_price"`
	Timestamp uint64 `db:"timestamp"`
	UniqID    string `db:"uniq_id"`
}
