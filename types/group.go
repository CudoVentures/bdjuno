package types

import "time"

type GroupMember struct {
	Address        string
	Weight         uint64
	MemberMetadata string
}

func NewGroupMember(address string, weight uint64, memberMetadata string) *GroupMember {
	return &GroupMember{
		Address:        address,
		Weight:         weight,
		MemberMetadata: memberMetadata,
	}
}

type GroupWithPolicy struct {
	ID                 uint64
	Address            string
	Members            []*GroupMember
	GroupMetadata      string
	PolicyMetadata     string
	Threshold          uint64
	VotingPeriod       uint64
	MinExecutionPeriod uint64
}

func NewGroupWithPolicy(
	id uint64,
	address string,
	members []*GroupMember,
	groupMetadata string,
	policyMetadata string,
	threshold uint64,
	votingPeriod uint64,
	minExecutionPeriod uint64,
) *GroupWithPolicy {
	return &GroupWithPolicy{
		ID:                 id,
		Address:            address,
		Members:            members,
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
}

func NewGroupProposal(
	id uint64,
	groupID uint64,
	proposalMetadata string,
	proposer string,
	submitTime time.Time,
	status string,
	executorResult string,
	messages string,
) *GroupProposal {
	return &GroupProposal{
		ID:               id,
		GroupID:          groupID,
		ProposalMetadata: proposalMetadata,
		Proposer:         proposer,
		SubmitTime:       submitTime,
		Status:           status,
		ExecutorResult:   executorResult,
		Messages:         messages,
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
