package types

import (
	"time"
)

type GroupWithPolicy struct {
	ID                 uint64
	Address            string
	GroupMetadata      string
	PolicyMetadata     string
	Threshold          uint64
	VotingPeriod       uint64
	MinExecutionPeriod uint64
}

func NewGroupWithPolicy(
	id uint64,
	address string,
	groupMetadata string,
	policyMetadata string,
	threshold uint64,
	votingPeriod uint64,
	minExecutionPeriod uint64,
) *GroupWithPolicy {
	return &GroupWithPolicy{
		ID:                 id,
		Address:            address,
		GroupMetadata:      groupMetadata,
		PolicyMetadata:     policyMetadata,
		Threshold:          threshold,
		VotingPeriod:       votingPeriod,
		MinExecutionPeriod: minExecutionPeriod,
	}
}

type GroupProposal struct {
	ID               uint64
	GroupID          uint64
	ProposalMetadata string
	Proposer         string
	SubmitTime       time.Time
	Status           string
	ExecutorResult   string
	Messages         string
	BlockHeight      int64
}

func NewGroupProposal(
	id uint64,
	groupID uint64,
	proposalMetadata string,
	proposer string,
	status string,
	executorResult string,
	messages string,
	height int64,
) *GroupProposal {
	return &GroupProposal{
		ID:               id,
		GroupID:          groupID,
		ProposalMetadata: proposalMetadata,
		Proposer:         proposer,
		Status:           status,
		ExecutorResult:   executorResult,
		Messages:         messages,
		BlockHeight:      height,
	}
}

type GroupProposalVote struct {
	ProposalID   uint64
	Voter        string
	VoteOption   string
	VoteMetadata string
	SubmitTime   time.Time
}

func NewGroupProposalVote(
	proposalID uint64,
	voter string,
	voteOption string,
	voteMetadata string,
	submitTime time.Time,
) *GroupProposalVote {
	return &GroupProposalVote{
		ProposalID:   proposalID,
		Voter:        voter,
		VoteOption:   voteOption,
		VoteMetadata: voteMetadata,
		SubmitTime:   submitTime,
	}
}
