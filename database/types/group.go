package types

import "database/sql"

type GroupWithPolicyRow struct {
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
	ProposalMetadata string         `db:"proposal_metadata"`
	Proposer         string         `db:"proposer"`
	Status           string         `db:"status"`
	ExecutorResult   string         `db:"executor_result"`
	Messages         string         `db:"messages"`
	TxHash           sql.NullString `db:"transaction_hash"`
	BlockHeight      int64          `db:"height"`
}

type GroupMemberRow struct {
	Address        string `db:"address"`
	GroupID        uint64 `db:"group_id"`
	Weight         uint64 `db:"weight"`
	MemberMetadata string `db:"member_metadata"`
}
