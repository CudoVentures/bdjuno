package types

import (
	"database/sql"
	"time"
)

type GroupRow struct {
	ID                 uint64 `db:"id"`
	Address            string `db:"address"`
	GroupMetadata      string `db:"group_metadata"`
	PolicyMetadata     string `db:"policy_metadata"`
	Threshold          uint64 `db:"threshold"`
	VotingPeriod       uint64 `db:"voting_period"`
	MinExecutionPeriod uint64 `db:"min_execution_period"`
}

type GroupProposalRow struct {
	ID               uint64         `db:"id"`
	GroupID          uint64         `db:"group_id"`
	ProposalMetadata string         `db:"metadata"`
	Proposer         string         `db:"proposer"`
	Status           string         `db:"status"`
	ExecutorResult   string         `db:"executor_result"`
	Executor         sql.NullString `db:"executor"`
	ExecutionTime    sql.NullTime   `db:"execution_time"`
	ExecutionLog     sql.NullString `db:"execution_log"`
	Messages         string         `db:"messages"`
	TxHash           sql.NullString `db:"transaction_hash"`
	BlockHeight      int64          `db:"height"`
	SubmitTime       time.Time      `db:"submit_time"`
}

type GroupMemberRow struct {
	Address  string `db:"address"`
	GroupID  uint64 `db:"group_id"`
	Weight   uint64 `db:"weight"`
	Metadata string `db:"metadata"`
}

type GroupProposalVoteRow struct {
	ProposalID   uint64    `db:"proposal_id"`
	GroupID      uint64    `db:"group_id"`
	Voter        string    `db:"voter"`
	VoteOption   string    `db:"vote_option"`
	VoteMetadata string    `db:"vote_metadata"`
	SubmitTime   time.Time `db:"submit_time"`
}
