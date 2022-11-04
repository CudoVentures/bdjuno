package database

import "fmt"

func (db *Db) SaveMarketplaceCollection(txHash string, id uint64, denomID, mintRoyalties, resaleRoyalties, creator string, verified bool) error {
	_, err := db.Sql.Exec(`INSERT INTO marketplace_collection (transaction_hash, id, denom_id, mint_royalties, resale_royalties, verified, creator) 
		VALUES($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING`, txHash, id, denomID, mintRoyalties, resaleRoyalties, verified, creator)
	return err
}

func (db *Db) SaveMarketplaceNft(txHash string, id, nftID uint64, denomID, price, creator string) error {
	_, err := db.Sql.Exec(`INSERT INTO marketplace_nft (transaction_hash, id, token_id, denom_id, price, creator) 
		VALUES($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`, txHash, id, nftID, denomID, price, creator)
	return err
}

func (tx *DbTx) RemoveMarketplaceNft(id uint64) error {
	_, err := tx.Exec(`DELETE FROM marketplace_nft WHERE id = $1`, id)
	return err
}

func (tx *DbTx) SaveMarketplaceNftBuy(txHash string, id uint64, buyer string, timestamp uint64) error {
	var tokenID uint64
	var denomID, price, seller string

	if err := tx.QueryRow(`SELECT token_id, denom_id, price, creator FROM marketplace_nft WHERE id = $1`, id).Scan(&tokenID, &denomID, &price, &seller); err != nil {
		return err
	}

	if seller == "" {
		return fmt.Errorf("nft (%d) not found for sale", id)
	}

	_, err := tx.Exec(`INSERT INTO marketplace_nft_buy_history (transaction_hash, token_id, denom_id, price, seller, buyer, timestamp) 
		VALUES($1, $2, $3, $4, $5, $6, $7)`, txHash, tokenID, denomID, price, seller, buyer, timestamp)
	return err
}

func (db *Db) SetMarketplaceCollectionVerificationStatus(id uint64, verified bool) error {
	_, err := db.Sql.Exec(`UPDATE marketplace_collection SET verified = $1 WHERE id = $2`, verified, id)
	return err
}

func (db *Db) SetMarketplaceNFTPrice(id uint64, price string) error {
	_, err := db.Sql.Exec(`UPDATE marketplace_nft SET price = $1 WHERE id = $2`, price, id)
	return err
}

func (db *Db) SetMarketplaceCollectionRoyalties(id uint64, mintRoyalties, resaleRoyalties string) error {
	_, err := db.Sql.Exec(`UPDATE marketplace_collection SET mint_royalties = $1, resale_royalties = $2 WHERE id = $3`, mintRoyalties, resaleRoyalties, id)
	return err
}
