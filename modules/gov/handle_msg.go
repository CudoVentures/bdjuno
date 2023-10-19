package gov

import (
	"fmt"
	"strconv"
	"time"

	"github.com/forbole/bdjuno/v4/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	legacyTypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	juno "github.com/forbole/juno/v5/types"
)

// HandleMsg implements modules.MessageModule
func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	if len(tx.Logs) == 0 {
		return nil
	}

	switch cosmosMsg := msg.(type) {
	case *govtypesv1.MsgSubmitProposal, *legacyTypes.MsgSubmitProposal:
		return m.handleMsgSubmitProposalWithLegacySupport(tx, index, cosmosMsg)

	case *govtypesv1.MsgDeposit, *legacyTypes.MsgDeposit:
		return m.handleMsgDepositWithLegacySupport(tx, cosmosMsg)

	case *govtypesv1.MsgVote, *legacyTypes.MsgVote:
		return m.handleMsgVoteWithLegacySupport(tx, cosmosMsg)

	case *govtypesv1.MsgVoteWeighted, *legacyTypes.MsgVoteWeighted:
		return m.handleMsgVoteWeightedWithLegacySupport(tx, cosmosMsg)
	}

	return nil
}

// handleMsgSubmitProposal allows to properly handle a handleMsgSubmitProposal
//
//lint:ignore U1000 we might need the original implementation later
func (m *Module) handleMsgSubmitProposal(tx *juno.Tx, index int, msg *govtypesv1.MsgSubmitProposal) error {
	// Get the proposal id
	event, err := tx.FindEventByType(index, gov.EventTypeSubmitProposal)
	if err != nil {
		return fmt.Errorf("error while searching for EventTypeSubmitProposal: %s", err)
	}

	id, err := tx.FindAttributeByKey(event, gov.AttributeKeyProposalID)
	if err != nil {
		return fmt.Errorf("error while searching for AttributeKeyProposalID: %s", err)
	}

	proposalID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return fmt.Errorf("error while parsing proposal id: %s", err)
	}

	// Get the proposal
	proposal, err := m.source.Proposal(tx.Height, proposalID)
	if err != nil {
		return fmt.Errorf("error while getting proposal: %s", err)
	}

	// Store the proposal
	proposalObj := types.NewProposal(
		proposal.Id,
		proposal.Title,
		proposal.Summary,
		proposal.Metadata,
		msg.Messages,
		proposal.Status.String(),
		*proposal.SubmitTime,
		*proposal.DepositEndTime,
		proposal.VotingStartTime,
		proposal.VotingEndTime,
		msg.Proposer,
	)

	err = m.db.SaveProposals([]types.Proposal{proposalObj})
	if err != nil {
		return err
	}

	txTimestamp, err := time.Parse(time.RFC3339, tx.Timestamp)
	if err != nil {
		return fmt.Errorf("error while parsing time: %s", err)
	}

	// Store the deposit
	deposit := types.NewDeposit(proposal.Id, msg.Proposer, msg.InitialDeposit, txTimestamp, tx.Height)
	return m.db.SaveDeposits([]types.Deposit{deposit})
}

// handleMsgDeposit allows to properly handle a handleMsgDeposit
//
//lint:ignore U1000 we might need the original implementation later
func (m *Module) handleMsgDeposit(tx *juno.Tx, msg *govtypesv1.MsgDeposit) error {
	deposit, err := m.source.ProposalDeposit(tx.Height, msg.ProposalId, msg.Depositor)
	if err != nil {
		return fmt.Errorf("error while getting proposal deposit: %s", err)
	}
	txTimestamp, err := time.Parse(time.RFC3339, tx.Timestamp)
	if err != nil {
		return fmt.Errorf("error while parsing time: %s", err)
	}

	return m.db.SaveDeposits([]types.Deposit{
		types.NewDeposit(msg.ProposalId, msg.Depositor, deposit.Amount, txTimestamp, tx.Height),
	})
}

// handleMsgVote allows to properly handle a handleMsgVote
//
//lint:ignore U1000 we might need the original implementation later
func (m *Module) handleMsgVote(tx *juno.Tx, msg *govtypesv1.MsgVote) error {
	txTimestamp, err := time.Parse(time.RFC3339, tx.Timestamp)
	if err != nil {
		return fmt.Errorf("error while parsing time: %s", err)
	}

	vote := types.NewVote(msg.ProposalId, msg.Voter, msg.Option, txTimestamp, tx.Height)

	return m.db.SaveVote(vote)
}

//lint:ignore U1000 we might need the original implementation later
func (m *Module) handleMsgVoteWeighted(tx *juno.Tx, msg *govtypesv1.MsgVoteWeighted) error {
	weightedVote := types.NewWeightedVote(msg.ProposalId, msg.Voter, msg.Options, tx.Height)
	return m.db.SaveWeightedVote(weightedVote)
}

// handleMsgSubmitProposal allows to properly handle a handleMsgSubmitProposal
func (m *Module) handleMsgSubmitProposalWithLegacySupport(tx *juno.Tx, index int, msg interface{}) error {
	// Get the proposal id
	event, err := tx.FindEventByType(index, gov.EventTypeSubmitProposal)
	if err != nil {
		return fmt.Errorf("error while searching for EventTypeSubmitProposal: %s", err)
	}

	id, err := tx.FindAttributeByKey(event, gov.AttributeKeyProposalID)
	if err != nil {
		return fmt.Errorf("error while searching for AttributeKeyProposalID: %s", err)
	}

	proposalID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return fmt.Errorf("error while parsing proposal id: %s", err)
	}

	// Get the proposal
	proposal, err := m.source.Proposal(tx.Height, proposalID)
	if err != nil {
		return fmt.Errorf("error while getting proposal: %s", err)
	}

	// Prepare the proposalObj
	var proposalObj types.Proposal
	switch msg.(type) {
	case *govtypesv1.MsgSubmitProposal, *legacyTypes.MsgSubmitProposal:
		proposalObj = types.NewProposal(
			proposal.Id,
			proposal.Title,
			proposal.Summary,
			proposal.Metadata,
			proposal.Messages,
			proposal.Status.String(),
			*proposal.SubmitTime,
			*proposal.DepositEndTime,
			proposal.VotingStartTime,
			proposal.VotingEndTime,
			proposal.Proposer,
		)
	default:
		return fmt.Errorf("unexpected type %T", msg)
	}

	// Get the initialDeposit
	var initialDeposit []sdk.Coin
	switch msgType := msg.(type) {
	case *govtypesv1.MsgSubmitProposal:
		initialDeposit = msgType.InitialDeposit
	case *legacyTypes.MsgSubmitProposal:
		initialDeposit = msgType.InitialDeposit
	}

	err = m.db.SaveProposals([]types.Proposal{proposalObj})
	if err != nil {
		return err
	}

	txTimestamp, err := time.Parse(time.RFC3339, tx.Timestamp)
	if err != nil {
		return fmt.Errorf("error while parsing time: %s", err)
	}

	// Store the deposit
	deposit := types.NewDeposit(proposal.Id, proposal.Proposer, initialDeposit, txTimestamp, tx.Height)
	return m.db.SaveDeposits([]types.Deposit{deposit})
}

func (m *Module) handleMsgDepositWithLegacySupport(tx *juno.Tx, msg interface{}) error {

	var proposalID uint64
	var depositor string
	switch msgType := msg.(type) {
	case *govtypesv1.MsgDeposit:
		proposalID = msgType.ProposalId
		depositor = msgType.Depositor
	case *legacyTypes.MsgDeposit:
		proposalID = msgType.ProposalId
		depositor = msgType.Depositor
	}

	deposit, err := m.source.ProposalDeposit(tx.Height, proposalID, depositor)
	if err != nil {
		return fmt.Errorf("error while getting proposal deposit: %s", err)
	}
	txTimestamp, err := time.Parse(time.RFC3339, tx.Timestamp)
	if err != nil {
		return fmt.Errorf("error while parsing time: %s", err)
	}

	return m.db.SaveDeposits([]types.Deposit{
		types.NewDeposit(proposalID, depositor, deposit.Amount, txTimestamp, tx.Height),
	})
}

func (m *Module) handleMsgVoteWithLegacySupport(tx *juno.Tx, msg interface{}) error {
	txTimestamp, err := time.Parse(time.RFC3339, tx.Timestamp)
	if err != nil {
		return fmt.Errorf("error while parsing time: %s", err)
	}

	switch msgType := msg.(type) {
	case *govtypesv1.MsgVote:
		return m.db.SaveVote(
			types.NewVote(
				msgType.ProposalId,
				msgType.Voter,
				msgType.Option,
				txTimestamp,
				tx.Height,
			),
		)
	case *legacyTypes.MsgVote:
		return m.db.SaveLegacyVote(
			types.NewLegacyVote(
				msgType.ProposalId,
				msgType.Voter,
				msgType.Option,
				txTimestamp,
				tx.Height,
			),
		)
	}

	return fmt.Errorf("error while saving vote")
}

func (m *Module) handleMsgVoteWeightedWithLegacySupport(tx *juno.Tx, msg interface{}) error {
	var weightedVote types.WeightedVote
	switch msgType := msg.(type) {
	case *govtypesv1.MsgVoteWeighted:
		weightedVote = types.NewWeightedVote(
			msgType.ProposalId,
			msgType.Voter,
			msgType.Options,
			tx.Height,
		)
	case *legacyTypes.MsgVoteWeighted:
		weightedVote = types.NewLegacyWeightedVote(
			msgType.ProposalId,
			msgType.Voter,
			msgType.Options,
			tx.Height,
		)
	default:
		return fmt.Errorf("error while saving weighted vote")
	}

	return m.db.SaveWeightedVote(weightedVote)
}
