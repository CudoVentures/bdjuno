package group

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/forbole/bdjuno/v2/database"
	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"
	juno "github.com/forbole/juno/v2/types"
)

func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	if len(tx.Logs) == 0 {
		return nil
	}

	return m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		m.dbTx = dbTx

		switch cosmosMsg := msg.(type) {
		case *group.MsgCreateGroupWithPolicy:
			return m.handleMsgCreateGroupWithPolicy(tx, index, cosmosMsg)
		case *group.MsgSubmitProposal:
			return m.handleMsgSubmitProposal(tx, index, cosmosMsg)
		case *group.MsgVote:
			return m.handleMsgVote(tx, index, cosmosMsg)
		case *group.MsgExec:
			return m.handleMsgExec(tx, index, cosmosMsg.ProposalId)
		case *group.MsgWithdrawProposal:
			return m.handleMsgWithdrawProposal(tx, index, cosmosMsg.ProposalId)
		}

		return nil
	})

}

func (m *Module) handleMsgCreateGroupWithPolicy(tx *juno.Tx, index int, msg *group.MsgCreateGroupWithPolicy) error {
	groupIDAttr := utils.GetValueFromLogs(uint32(index), tx.Logs, "cosmos.group.v1.EventCreateGroup", "group_id")
	if groupIDAttr == "" {
		return errors.New("error while getting EventCreateGroup")
	}

	groupID, err := strconv.ParseUint(groupIDAttr, 10, 64)
	if err != nil {
		return err
	}

	address := utils.GetValueFromLogs(uint32(index), tx.Logs, "cosmos.group.v1.EventCreateGroupPolicy", "address")
	if address == "" {
		return errors.New("error while getting EventCreateGroupPolicy")
	}

	decisionPolicy, ok := msg.DecisionPolicy.GetCachedValue().(*group.ThresholdDecisionPolicy)
	if !ok {
		return errors.New("error while parsing decision policy")
	}

	threshold, err := strconv.ParseUint(decisionPolicy.Threshold, 10, 64)
	if err != nil {
		return err
	}

	votingPeriod := uint64(decisionPolicy.Windows.VotingPeriod.Seconds())
	minExecutionPeriod := uint64(decisionPolicy.Windows.MinExecutionPeriod.Seconds())
	group := types.NewGroup(groupID, address, msg.GroupMetadata, msg.GroupPolicyMetadata, threshold, votingPeriod, minExecutionPeriod)

	if err := m.dbTx.SaveGroup(group); err != nil {
		return err
	}

	members := make([]*types.Member, len(msg.Members))
	for i, m := range msg.Members {
		weight, err := strconv.ParseUint(m.Weight, 10, 64)
		if err != nil {
			return err
		}

		members[i] = types.NewMember(m.Address, weight, m.Metadata)
	}

	return m.dbTx.SaveMembers(groupID, members)
}

func (m *Module) handleMsgSubmitProposal(tx *juno.Tx, index int, msg *group.MsgSubmitProposal) error {
	proposalIDAttr := utils.GetValueFromLogs(uint32(index), tx.Logs, "cosmos.group.v1.EventSubmitProposal", "proposal_id")
	if proposalIDAttr == "" {
		return errors.New("error while getting EventSubmitProposal")
	}

	proposalID, err := strconv.ParseUint(proposalIDAttr, 10, 64)
	if err != nil {
		return err
	}

	msgBytes, err := json.Marshal(msg.Messages)
	if err != nil {
		return err
	}

	groupID, err := m.dbTx.GetGroupIDByGroupAddress(msg.GroupPolicyAddress)
	if err != nil {
		return err
	}

	status := group.PROPOSAL_STATUS_SUBMITTED.String()
	result := group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN.String()
	msgs := utils.SanitizeUTF8(string(msgBytes))

	timestamp, err := time.Parse(time.RFC3339, tx.Timestamp)
	if err != nil {
		return err
	}

	proposal := types.NewGroupProposal(proposalID, groupID, msg.Metadata, msg.Proposers[0], status, result, msgs, tx.Height, timestamp)

	if err := m.dbTx.SaveProposal(proposal); err != nil {
		return err
	}

	if msg.Exec == group.Exec_EXEC_TRY {
		msgVote := group.MsgVote{
			ProposalId: proposalID,
			Voter:      msg.Proposers[0],
			Option:     group.VOTE_OPTION_YES,
			Metadata:   "",
			Exec:       group.Exec_EXEC_TRY,
		}

		if err := m.handleMsgVote(tx, index, &msgVote); err != nil {
			return err
		}
	}

	return nil
}

func (m *Module) handleMsgVote(tx *juno.Tx, index int, msg *group.MsgVote) error {
	voteEvent := utils.GetValueFromLogs(uint32(index), tx.Logs, "cosmos.group.v1.EventVote", "proposal_id")
	if voteEvent == "" {
		return errors.New("error while getting EventVote")
	}

	proposal, err := m.dbTx.GetProposal(msg.ProposalId)
	if err != nil {
		return err
	}

	timestamp, err := time.Parse(time.RFC3339, tx.Timestamp)
	if err != nil {
		return err
	}

	vote := types.NewProposalVote(msg.ProposalId, proposal.GroupID, msg.Voter, msg.Option.String(), msg.Metadata, timestamp)
	if err := m.dbTx.SaveProposalVote(vote); err != nil {
		return err
	}

	status, err := m.updateProposalStatus(msg.ProposalId, proposal.GroupID)
	if err != nil {
		return err
	}

	if msg.Exec == group.Exec_EXEC_TRY && status == group.PROPOSAL_STATUS_ACCEPTED {
		if err := m.handleMsgExec(tx, index, msg.ProposalId); err != nil {
			return err
		}
	}

	return nil
}

func (m *Module) updateProposalStatus(proposalID uint64, groupID uint64) (group.ProposalStatus, error) {
	votes, err := m.dbTx.GetProposalVotes(proposalID)
	if err != nil {
		return 0, err
	}

	votesYes := 0
	for _, v := range votes {
		if v == group.VOTE_OPTION_YES.String() {
			votesYes++
		}
	}

	threshold, err := m.dbTx.GetGroupThreshold(groupID)
	if err != nil {
		return 0, err
	}

	if votesYes >= threshold {
		err := m.dbTx.UpdateProposalStatus(proposalID, group.PROPOSAL_STATUS_ACCEPTED.String())
		return group.PROPOSAL_STATUS_ACCEPTED, err
	}
	totalPower, err := m.dbTx.GetGroupTotalVotingPower(groupID)
	if err != nil {
		return 0, err
	}

	votesRemaining := totalPower - len(votes)
	maxPossibleYesCount := votesYes + votesRemaining
	if maxPossibleYesCount < threshold {
		err := m.dbTx.UpdateProposalStatus(proposalID, group.PROPOSAL_STATUS_REJECTED.String())
		return group.PROPOSAL_STATUS_REJECTED, err
	}

	return 0, nil
}

func (m *Module) handleMsgExec(tx *juno.Tx, index int, proposalID uint64) error {
	executorResult := utils.GetValueFromLogs(uint32(index), tx.Logs, "cosmos.group.v1.EventExec", "result")
	if executorResult == "" {
		return errors.New("error while getting EventExec")
	}

	if err := m.dbTx.UpdateProposalExecutorResult(proposalID, executorResult, tx.TxHash); err != nil {
		return err
	}

	proposal, err := m.dbTx.GetProposal(proposalID)
	if err != nil {
		return err
	}

	isSuccess := executorResult == group.PROPOSAL_EXECUTOR_RESULT_SUCCESS.String()
	if isSuccess && strings.Contains(proposal.Messages, "MsgUpdateGroup") {
		return m.handleMsgUpdateGroup(proposal)
	}

	return nil

}

func (m *Module) handleMsgUpdateGroup(proposal *dbtypes.GroupProposalRow) error {
	if proposal.ExecutorResult != group.PROPOSAL_EXECUTOR_RESULT_SUCCESS.String() {
		return errors.New("error while executing handleMsgUpdateGroup")
	}

	abort := group.PROPOSAL_STATUS_ABORTED.String()
	if err := m.dbTx.UpdateActiveProposalStatusesByGroup(proposal.GroupID, abort); err != nil {
		return err
	}

	var msgTypes []types.MsgType
	if err := json.Unmarshal([]byte(proposal.Messages), &msgTypes); err != nil {
		return err
	}

	var msgs []json.RawMessage
	if err := json.Unmarshal([]byte(proposal.Messages), &msgs); err != nil {
		return err
	}

	for i := range msgs {
		switch msgTypes[i].TypeURL {
		case "/cosmos.group.v1.MsgUpdateGroupMembers":
			var msg types.MsgUpdateMembers
			if err := json.Unmarshal(msgs[i], &msg); err != nil {
				return err
			}
			err := m.handleMsgUpdateGroupMembers(&msg)
			if err != nil {
				return err
			}
		case "/cosmos.group.v1.MsgUpdateGroupMetadata":
			var msg group.MsgUpdateGroupMetadata
			if err := json.Unmarshal(msgs[i], &msg); err != nil {
				return err
			}
			err := m.dbTx.UpdateGroupMetadata(proposal.GroupID, msg.Metadata)
			if err != nil {
				return err
			}
		case "/cosmos.group.v1.MsgUpdateGroupPolicyMetadata":
			var msg group.MsgUpdateGroupPolicyMetadata
			if err := json.Unmarshal(msgs[i], &msg); err != nil {
				return err
			}
			err := m.dbTx.UpdateGroupPolicyMetadata(proposal.GroupID, msg.Metadata)
			if err != nil {
				return err
			}
		case "/cosmos.group.v1.MsgUpdateGroupPolicyDecisionPolicy":
			var msg types.MsgUpdateDecisionPolicy
			if err := json.Unmarshal(msgs[i], &msg); err != nil {
				return err
			}
			err := m.dbTx.UpdateDecisionPolicy(proposal.GroupID, msg.DecisionPolicy)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Module) handleMsgUpdateGroupMembers(msg *types.MsgUpdateMembers) error {
	updateMembers := make([]*types.Member, 0)
	deleteMembers := make([]string, 0)
	for i, m := range msg.MemberUpdates {
		if m.Weight == 0 {
			deleteMembers = append(deleteMembers, m.Address)
		} else {
			updateMembers = append(updateMembers, &msg.MemberUpdates[i])
		}
	}

	if err := m.dbTx.SaveMembers(msg.GroupID, updateMembers); err != nil {
		return err
	}

	return m.dbTx.RemoveMembers(msg.GroupID, deleteMembers)
}

func (m *Module) handleMsgWithdrawProposal(tx *juno.Tx, index int, proposalID uint64) error {
	executorResult := utils.GetValueFromLogs(uint32(index), tx.Logs, "cosmos.group.v1.EventWithdrawProposal", "proposal_id")
	if executorResult == "" {
		return errors.New("error while getting EventWithdraw")
	}

	return m.dbTx.UpdateProposalStatus(proposalID, group.PROPOSAL_STATUS_WITHDRAWN.String())
}
