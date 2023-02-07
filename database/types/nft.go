package types

type NftTransferHistoryRow struct {
	ID        uint64 `db:"id"`
	TxHash    string `db:"transaction_hash"`
	DenomID   string `db:"denom_id"`
	Price     string `db:"price"`
	OldOwner  string `db:"old_owner"`
	NewOwner  string `db:"new_owner"`
	Timestamp uint64 `db:"timestamp"`
	UniqID    string `db:"uniq_id"`
}
