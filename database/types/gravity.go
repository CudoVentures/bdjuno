package types

type GravityOrchestratorRow struct {
	Address string `db:"address"`
}

type GravityTransactionRow struct {
	Type            string `db:"type"`
	AttestationID   string `db:"attestation_id"`
	Orchestrator    string `db:"orchestrator"`
	Receiver        string `db:"receiver"`
	Votes           int    `db:"votes"`
	Consensus       bool   `db:"consensus"`
	Height          int64  `db:"height"`
	TransactionHash string `db:"transaction_hash"`
}
