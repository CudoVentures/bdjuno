package types

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
	GroupID uint64 `db:"group_id"`
}
