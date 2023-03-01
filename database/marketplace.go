package database

import (
	"fmt"

	"github.com/forbole/bdjuno/v2/database/utils"
)

func (db *Db) CheckIfNftExists(tokenId uint64, denomId string) error {
	var rows []string

	err := db.Sqlx.Select(&rows, `SELECT denom_id FROM marketplace_nft WHERE token_id=$1 AND denom_id=$2`, tokenId, denomId)
	if err != nil {
		return err
	}

	if len(rows) != 1 {
		return fmt.Errorf("Not found.")
	}

	return nil
}

func (db *Db) SaveMarketplaceCollection(txHash string, id uint64, denomID, mintRoyalties, resaleRoyalties, creator string, verified bool) error {
	_, err := db.Sql.Exec(`INSERT INTO marketplace_collection (transaction_hash, id, denom_id, mint_royalties, resale_royalties, verified, creator) 
		VALUES($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING`, txHash, id, denomID, mintRoyalties, resaleRoyalties, verified, creator)
	return err
}

func (tx *DbTx) ListNft(txHash string, id, tokenId uint64, denomID, price string) error {
	_, err := tx.Exec(`UPDATE marketplace_nft SET transaction_hash=$1, id=$2, price=$3 WHERE token_id=$4 AND denom_id=$5`,
		txHash, id, price, tokenId, denomID)
	fmt.Println(err)
	return err
}

func (tx *DbTx) SaveMarketplaceNft(txHash string, tokenId uint64, denomID, uid, price, creator string) error {
	_, err := tx.Exec(`INSERT INTO marketplace_nft (transaction_hash, uid, token_id, denom_id, price, creator, uniq_id) 
		VALUES($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (token_id, denom_id) DO UPDATE SET price = EXCLUDED.price, id = EXCLUDED.id`,
		txHash, uid, tokenId, denomID, price, creator, utils.FormatUniqID(tokenId, denomID))
	return err
}

func (tx *DbTx) SaveMarketplaceNftBuy(txHash string, id uint64, buyer string, timestamp uint64, usdPrice, btcPrice string) error {
	var tokenID uint64
	var denomID, price, seller string

	if err := tx.QueryRow(`SELECT token_id, denom_id, price, creator FROM marketplace_nft WHERE id = $1`, id).Scan(&tokenID, &denomID, &price, &seller); err != nil {
		return err
	}

	if seller == "" {
		return fmt.Errorf("nft (%d) not found for sale", id)
	}

	return tx.saveMarketplaceNftBuy(txHash, buyer, timestamp, tokenID, denomID, price, seller, usdPrice, btcPrice)
}

func (tx *DbTx) saveMarketplaceNftBuy(txHash string, buyer string, timestamp, tokenID uint64, denomID, price, seller, usdPrice, btcPrice string) error {
	_, err := tx.Exec(`INSERT INTO marketplace_nft_buy_history (transaction_hash, token_id, denom_id, price, seller, buyer, usd_price, btc_price, timestamp, uniq_id) 
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`, txHash, tokenID, denomID, price, seller, buyer, usdPrice, btcPrice, timestamp, utils.FormatUniqID(tokenID, denomID))
	return err
}

func (tx *DbTx) SaveMarketplaceNftMint(txHash string, tokenID uint64, buyer, denomID, price string, timestamp uint64, usdPrice, btcPrice string) error {
	return tx.saveMarketplaceNftBuy(txHash, buyer, timestamp, tokenID, denomID, price, "0x0", usdPrice, btcPrice)
}

func (tx *DbTx) SetMarketplaceNFTPrice(id uint64, price string) error {
	_, err := tx.Exec(`UPDATE marketplace_nft SET price = $1 WHERE id = $2`, price, id)
	return err
}

func (tx *DbTx) UnlistNft(id uint64) error {
	_, err := tx.Exec(`UPDATE marketplace_nft SET price = '0', id = null WHERE id = $1`, id)
	return err
}

func (db *Db) UnlistNft(id uint64) error {
	_, err := db.Sql.Exec(`UPDATE marketplace_nft SET price = '0', id = null WHERE id = $1`, id)
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
