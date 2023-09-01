package types

import "database/sql"

type GravityOrchestratorRow struct {
	Address string `db:"address"`
}

type GravityTransactionRow struct {
	Type            string        `db:"type"`
	AttestationID   string        `db:"attestation_id"`
	Orchestrator    string        `db:"orchestrator"`
	Receiver        string        `db:"receiver"`
	Votes           int           `db:"votes"`
	Consensus       bool          `db:"consensus"`
	Height          int64         `db:"height"`
	PartitionID     sql.NullInt64 `db:"partition_id"`
	TransactionHash string        `db:"transaction_hash"`
}
