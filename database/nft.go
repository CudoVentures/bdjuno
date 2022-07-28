package database

func (db *Db) SaveDenom(txHash, denomID, name, schema, symbol, owner, contractAddressSigner string) error {
	_, err := db.Sql.Exec(`INSERT INTO nft_denom (transaction_hash, id, name, schema, symbol, owner, contract_address_signer) 
		VALUES($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING`, txHash, denomID, name, schema, symbol, owner, contractAddressSigner)
	return err
}

func (db *Db) UpdateDenom(denomID, owner string) error {
	_, err := db.Sql.Exec(`UPDATE nft_denom SET owner = $1 WHERE id = $2`, owner, denomID)
	return err
}

func (db *Db) SaveNFT(txHash string, tokenID uint64, denomID, name, uri, data, owner, sender, contractAddressSigner string) error {
	_, err := db.Sql.Exec(`INSERT INTO nft_nft (transaction_hash, id, denom_id, name, uri, data, owner, sender, contract_address_signer) 
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT DO NOTHING`, txHash, tokenID, denomID, name, uri, data, owner, sender, contractAddressSigner)
	return err
}

func (db *Db) UpdateNFT(id, denomID, name, uri, data string) error {
	_, err := db.Sql.Exec(`UPDATE nft_nft SET name = $1, uri = $2, data = $3 WHERE id = $4 AND denom_id = $5`, name, uri, data, id, denomID)
	return err
}

func (db *Db) UpdateNFTOwner(id, denomID, owner string) error {
	_, err := db.Sql.Exec(`UPDATE nft_nft SET owner = $1 WHERE id = $2 AND denom_id = $3`, owner, id, denomID)
	return err
}

func (db *Db) BurnNFT(id, denomID string) error {
	_, err := db.Sql.Exec(`UPDATE nft_nft SET burned = true WHERE id = $1 AND denom_id = $2`, id, denomID)
	return err
}
