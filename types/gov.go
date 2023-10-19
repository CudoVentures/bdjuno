package types

import (
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	legacyTypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalStatusInvalid = "PROPOSAL_STATUS_INVALID"
)

// GovParams contains the data of the x/gov module parameters
type GovParams struct {
	*govtypesv1.Params
	Height int64 `json:"height" ymal:"height"`
}

func NewGovParams(params *govtypesv1.Params, height int64) *GovParams {
	return &GovParams{
		Params: params,
		Height: height,
	}
}

// --------------------------------------------------------------------------------------------------------------------

// Proposal represents a single governance proposal
type Proposal struct {
	ID              uint64
	Title           string
	Summary         string
	Metadata        string
	Messages        []*codectypes.Any
	Status          string
	SubmitTime      time.Time
	DepositEndTime  time.Time
	VotingStartTime *time.Time
	VotingEndTime   *time.Time
	Proposer        string
}

// NewProposal return a new Proposal instance
func NewProposal(
	proposalID uint64,
	title string,
	summary string,
	metadata string,
	messages []*codectypes.Any,
	status string,
	submitTime time.Time,
	depositEndTime time.Time,
	votingStartTime *time.Time,
	votingEndTime *time.Time,
	proposer string,
) Proposal {
	return Proposal{
		ID:              proposalID,
		Title:           title,
		Summary:         summary,
		Metadata:        metadata,
		Messages:        messages,
		Status:          status,
		SubmitTime:      submitTime,
		DepositEndTime:  depositEndTime,
		VotingStartTime: votingStartTime,
		VotingEndTime:   votingEndTime,
		Proposer:        proposer,
	}
}

// ProposalUpdate contains the data that should be used when updating a governance proposal
type ProposalUpdate struct {
	ProposalID      uint64
	Status          string
	VotingStartTime *time.Time
	VotingEndTime   *time.Time
}

// NewProposalUpdate allows to build a new ProposalUpdate instance
func NewProposalUpdate(proposalID uint64, status string, votingStartTime, votingEndTime *time.Time) ProposalUpdate {
	return ProposalUpdate{
		ProposalID:      proposalID,
		Status:          status,
		VotingStartTime: votingStartTime,
		VotingEndTime:   votingEndTime,
	}
}

// -------------------------------------------------------------------------------------------------------------------

// Deposit contains the data of a single deposit made towards a proposal
type Deposit struct {
	ProposalID uint64
	Depositor  string
	Amount     sdk.Coins
	Timestamp  time.Time
	Height     int64
}

// NewDeposit return a new Deposit instance
func NewDeposit(
	proposalID uint64,
	depositor string,
	amount sdk.Coins,
	timestamp time.Time,
	height int64,
) Deposit {
	return Deposit{
		ProposalID: proposalID,
		Depositor:  depositor,
		Amount:     amount,
		Timestamp:  timestamp,
		Height:     height,
	}
}

// -------------------------------------------------------------------------------------------------------------------

// Vote contains the data of a single proposal vote
type Vote struct {
	ProposalID uint64
	Voter      string
	Option     govtypesv1.VoteOption
	Timestamp  time.Time
	Height     int64
}

type LegacyVote struct {
	ProposalID uint64
	Voter      string
	Option     legacyTypes.VoteOption
	Timestamp  time.Time
	Height     int64
}

// NewVote return a new Vote instance
func NewVote(
	proposalID uint64,
	voter string,
	option govtypesv1.VoteOption,
	timestamp time.Time,
	height int64,
) Vote {
	return Vote{
		ProposalID: proposalID,
		Voter:      voter,
		Option:     option,
		Timestamp:  timestamp,
		Height:     height,
	}
}

func NewLegacyVote(
	proposalID uint64,
	voter string,
	option legacyTypes.VoteOption,
	timestamp time.Time,
	height int64,
) LegacyVote {
	return LegacyVote{
		ProposalID: proposalID,
		Voter:      voter,
		Option:     option,
		Timestamp:  timestamp,
		Height:     height,
	}
}

type WeightedVoteOption struct {
	Option string
	Weight string
}

type WeightedVote struct {
	ProposalID uint64
	Voter      string
	Options    []WeightedVoteOption
	Height     int64
}

func NewWeightedVote(
	proposalID uint64,
	voter string,
	options []*govtypesv1.WeightedVoteOption,
	height int64,
) WeightedVote {
	weightedvote := WeightedVote{
		ProposalID: proposalID,
		Voter:      voter,
		Height:     height,
	}
	for _, opt := range options {
		weightedvote.Options = append(weightedvote.Options, WeightedVoteOption{
			Option: govtypesv1.VoteOption_name[int32(opt.Option)],
			Weight: opt.Weight,
		})
	}
	return weightedvote
}

func NewLegacyWeightedVote(
	proposalID uint64,
	voter string,
	options []legacyTypes.WeightedVoteOption,
	height int64,
) WeightedVote {
	weightedvote := WeightedVote{
		ProposalID: proposalID,
		Voter:      voter,
		Height:     height,
	}
	for _, opt := range options {
		weightedvote.Options = append(weightedvote.Options, WeightedVoteOption{
			Option: govtypesv1.VoteOption_name[int32(opt.Option)],
			Weight: opt.Weight.String(),
		})
	}
	return weightedvote
}

// -------------------------------------------------------------------------------------------------------------------

// TallyResult contains the data about the final results of a proposal
type TallyResult struct {
	ProposalID uint64
	Yes        string
	Abstain    string
	No         string
	NoWithVeto string
	Height     int64
}

// NewTallyResult return a new TallyResult instance
func NewTallyResult(
	proposalID uint64,
	yes string,
	abstain string,
	no string,
	noWithVeto string,
	height int64,
) TallyResult {
	return TallyResult{
		ProposalID: proposalID,
		Yes:        yes,
		Abstain:    abstain,
		No:         no,
		NoWithVeto: noWithVeto,
		Height:     height,
	}
}

// -------------------------------------------------------------------------------------------------------------------

// ProposalStakingPoolSnapshot contains the data about a single staking pool snapshot to be associated with a proposal
type ProposalStakingPoolSnapshot struct {
	ProposalID uint64
	Pool       *PoolSnapshot
}

// NewProposalStakingPoolSnapshot returns a new ProposalStakingPoolSnapshot instance
func NewProposalStakingPoolSnapshot(proposalID uint64, pool *PoolSnapshot) ProposalStakingPoolSnapshot {
	return ProposalStakingPoolSnapshot{
		ProposalID: proposalID,
		Pool:       pool,
	}
}

// -------------------------------------------------------------------------------------------------------------------

// ProposalValidatorStatusSnapshot represents a single snapshot of the status of a validator associated
// with a single proposal
type ProposalValidatorStatusSnapshot struct {
	ProposalID           uint64
	ValidatorConsAddress string
	ValidatorVotingPower int64
	ValidatorStatus      int
	ValidatorJailed      bool
	Height               int64
}

// NewProposalValidatorStatusSnapshot returns a new ProposalValidatorStatusSnapshot instance
func NewProposalValidatorStatusSnapshot(
	proposalID uint64,
	validatorConsAddr string,
	validatorVotingPower int64,
	validatorStatus int,
	validatorJailed bool,
	height int64,
) ProposalValidatorStatusSnapshot {
	return ProposalValidatorStatusSnapshot{
		ProposalID:           proposalID,
		ValidatorStatus:      validatorStatus,
		ValidatorConsAddress: validatorConsAddr,
		ValidatorVotingPower: validatorVotingPower,
		ValidatorJailed:      validatorJailed,
		Height:               height,
	}
}
