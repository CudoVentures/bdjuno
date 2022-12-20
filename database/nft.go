package database

import "github.com/forbole/bdjuno/v2/database/utils"

func (db *Db) SaveDenom(txHash, denomID, name, schema, symbol, owner, contractAddressSigner, traits, minter, description, dataText, dataJSON string) error {
	_, err := db.Sql.Exec(`INSERT INTO nft_denom (transaction_hash, id, name, schema, symbol, owner, contract_address_signer, 
		traits, minter, description, data_text, data_json) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) ON CONFLICT DO NOTHING`,
		txHash, denomID, name, schema, symbol, owner, contractAddressSigner, traits, minter, description, dataText, dataJSON)
	return err
}

func (db *Db) UpdateDenom(denomID, owner string) error {
	_, err := db.Sql.Exec(`UPDATE nft_denom SET owner = $1 WHERE id = $2`, owner, denomID)
	return err
}

func (tx *DbTx) SaveNFT(txHash string, tokenID uint64, denomID, name, uri, dataJSON, dataText, owner, sender, contractAddressSigner string) error {
	_, err := tx.Exec(`INSERT INTO nft_nft (transaction_hash, id, denom_id, name, uri, owner, data_json, data_text, sender, contract_address_signer, uniq_id) 
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) ON CONFLICT DO NOTHING`, txHash, tokenID, denomID, name, uri, owner,
		dataJSON, dataText, sender, contractAddressSigner, utils.FormatUniqID(tokenID, denomID))
	return err
}

func (db *Db) UpdateNFT(id, denomID, name, uri, dataJSON, dataText string) error {
	_, err := db.Sql.Exec(`UPDATE nft_nft SET name = $1, uri = $2, data_json = $3, data_text = $4 WHERE id = $5 AND denom_id = $6`, name, uri, dataJSON, dataText, id, denomID)
	return err
}

func (tx *DbTx) UpdateNFTOwner(id, denomID, owner string) error {
	_, err := tx.Exec(`UPDATE nft_nft SET owner = $1 WHERE id = $2 AND denom_id = $3`, owner, id, denomID)
	return err
}

func (tx *DbTx) BurnNFT(id, denomID string) error {
	_, err := tx.Exec(`UPDATE nft_nft SET burned = true WHERE id = $1 AND denom_id = $2`, id, denomID)
	return err
}

func (tx *DbTx) UpdateNFTHistory(txHash string, tokenID uint64, denomID, from, to string, timestamp uint64) error {
	_, err := tx.Exec(`INSERT INTO nft_transfer_history (transaction_hash, id, denom_id, old_owner, new_owner, timestamp, uniq_id) 
		VALUES($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING`,
		txHash, tokenID, denomID, from, to, timestamp, utils.FormatUniqID(tokenID, denomID))
	return err
}
