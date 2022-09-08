package database

import "fmt"

func (db *Db) SaveMarketplaceCollection(txHash string, id uint64, denomID, mintRoyalties, resaleRoyalties, creator string) error {
	_, err := db.Sql.Exec(`INSERT INTO marketplace_collection (transaction_hash, id, denom_id, mint_royalties, resale_royalties, creator) 
		VALUES($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`, txHash, id, denomID, mintRoyalties, resaleRoyalties, creator)
	return err
}

func (db *Db) SaveMarketplaceNft(txHash string, id, nftID uint64, denomID, price, creator string) error {
	_, err := db.Sql.Exec(`INSERT INTO marketplace_nft (transaction_hash, id, nft_id, denom_id, price, creator) 
		VALUES($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`, txHash, id, nftID, denomID, price, creator)
	return err
}

func (db *Db) RemoveMarketplaceNft(id uint64) error {
	_, err := db.Sql.Exec(`DELETE FROM marketplace_nft WHERE id = $1`, id)
	return err
}

func (db *Db) SaveMarketplaceNftBuy(txHash string, id uint64, buyer string, timestamp uint64) error {
	type nft struct {
		TokenID uint64 `db:"token_id"`
		DenomID string `db:"denom_id"`
		Price   string `db:"price"`
		Seller  string `db:"creator"`
	}

	var rows []nft
	if err := db.Sql.QueryRow(`SELECT * FROM marketplace_nft WHERE id = $1`, id).Scan(&rows); err != nil {
		return err
	}

	if len(rows) == 0 {
		return fmt.Errorf("nft (%d) not found for sale", id)
	}

	_, err := db.Sql.Exec(`INSERT INTO marketplace_nft_buy_history (transaction_hash, token_id, denom_id, price, seller, buyer, timestamp) 
		VALUES($1, $2, $3, $4, $5, $6, $7)`, txHash, rows[0].TokenID, rows[0].DenomID, rows[0].Price, rows[0].Seller, buyer, timestamp)
	return err
}
