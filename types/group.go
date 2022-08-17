package types

import (
	"time"
)

type Group struct {
	ID                 uint64
	Address            string
	GroupMetadata      string
	PolicyMetadata     string
	Threshold          uint64
	VotingPeriod       uint64
	MinExecutionPeriod uint64
}

func NewGroup(
	id uint64,
	address string,
	groupMetadata string,
	policyMetadata string,
	threshold uint64,
	votingPeriod uint64,
	minExecutionPeriod uint64,
) *Group {
	return &Group{
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
	ID             uint64
	GroupID        uint64
	Metadata       string
	Proposer       string
	Status         string
	ExecutorResult string
	Messages       string
	BlockHeight    int64
	SubmitTime     time.Time
	MemberCount    int
}

func NewGroupProposal(
	id uint64,
	groupID uint64,
	metadata string,
	proposer string,
	status string,
	executorResult string,
	messages string,
	height int64,
	submitTime time.Time,
	memberCount int,
) *GroupProposal {
	return &GroupProposal{
		ID:             id,
		GroupID:        groupID,
		Metadata:       metadata,
		Proposer:       proposer,
		Status:         status,
		ExecutorResult: executorResult,
		Messages:       messages,
		BlockHeight:    height,
		SubmitTime:     submitTime,
		MemberCount:    memberCount,
	}
}

type ProposalVote struct {
	ProposalID   uint64
	GroupID      uint64
	Voter        string
	VoteOption   string
	VoteMetadata string
	SubmitTime   time.Time
}

func NewProposalVote(
	proposalID uint64,
	groupID uint64,
	voter string,
	voteOption string,
	voteMetadata string,
	submitTime time.Time,
) *ProposalVote {
	return &ProposalVote{
		ProposalID:   proposalID,
		GroupID:      groupID,
		Voter:        voter,
		VoteOption:   voteOption,
		VoteMetadata: voteMetadata,
		SubmitTime:   submitTime,
	}
}

type ProposalDecisionPolicy struct {
	ID           uint64
	VotingPeriod int
	SubmitTime   time.Time
}

type MsgType struct {
	TypeURL string `json:"@type,omitempty"`
}

type MsgUpdateDecisionPolicy struct {
	GroupPolicyAddress string                   `json:"group_policy_address,omitempty"`
	DecisionPolicy     *ThresholdDecisionPolicy `json:"decision_policy,omitempty"`
}

type ThresholdDecisionPolicy struct {
	Threshold uint64                 `json:"threshold,omitempty,string"`
	Windows   *DecisionPolicyWindows `json:"windows,omitempty"`
}

type DecisionPolicyWindows struct {
	VotingPeriod       uint64 `json:"voting_period,omitempty,string"`
	MinExecutionPeriod uint64 `json:"min_execution_period,omitempty,string"`
}

type MsgUpdateMembers struct {
	GroupID       uint64    `json:"group_id,omitempty,string"`
	MemberUpdates []*Member `json:"member_updates"`
}

type Member struct {
	Address  string `json:"address,omitempty"`
	Weight   uint64 `json:"weight,omitempty,string"`
	Metadata string `json:"metadata,omitempty"`
}

func NewMember(
	address string,
	weight uint64,
	metadata string,
) *Member {
	return &Member{
		Address:  address,
		Weight:   weight,
		Metadata: metadata,
	}
}

type ExecutionResult struct {
	ProposalID    uint64
	Result        string
	Executor      string
	ExecutionTime time.Time
	ExecutionLog  string
	TxHash        string
}

func NewExecutionResult(
	proposalID uint64,
	result string,
	executor string,
	executionTime time.Time,
	executionLog string,
	txHash string,
) *ExecutionResult {
	return &ExecutionResult{
		ProposalID:    proposalID,
		Result:        result,
		Executor:      executor,
		ExecutionTime: executionTime,
		ExecutionLog:  executionLog,
		TxHash:        txHash,
	}
}
