package database

func (db *Db) SaveOrchestrator(address string) error {
	_, err := db.Sql.Exec(`INSERT INTO gravity_orchestrator (address) VALUES($1) ON CONFLICT DO NOTHING`, address)
	return err
}

func (db *Db) GetOrchestratorsCount() (int, error) {
	var count int
	err := db.Sql.QueryRow(`SELECT COUNT(*) FROM gravity_orchestrator`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db *Db) SaveMsgSendToCosmosClaim(transactionHash, msgType, attestationID, receiver, orchestrator string, height int64) error {
	_, err := db.Sql.Exec(`INSERT INTO gravity_transaction AS gt (type, attestation_id, orchestrator, receiver, votes, consensus, transaction_hash, height) 
		VALUES($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (attestation_id) DO UPDATE SET 
		orchestrator = excluded.orchestrator, transaction_hash = excluded.transaction_hash, votes = gt.votes + 1`,
		msgType, attestationID, orchestrator, receiver, 1, false, transactionHash, height)
	return err
}

func (db *Db) GetGravityTransactionVotes(attestationID string) (int, error) {
	var votes int
	err := db.Sql.QueryRow(`SELECT votes FROM gravity_transaction WHERE attestation_id = $1`, attestationID).Scan(&votes)
	if err != nil {
		return 0, err
	}
	return votes, nil
}

func (db *Db) SetGravityTransactionConsensus(attestationID string, consensus bool) error {
	_, err := db.Sql.Exec(`UPDATE gravity_transaction SET consensus = $1 WHERE attestation_id = $2`, consensus, attestationID)
	return err
}
