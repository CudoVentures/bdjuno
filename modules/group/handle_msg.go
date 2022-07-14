package group

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
			return m.handleMsgWithdrawProposal(cosmosMsg.ProposalId)
		}

		return nil
	})

}

func (m *Module) handleMsgCreateGroupWithPolicy(tx *juno.Tx, index int, msg *group.MsgCreateGroupWithPolicy) error {
	groupIDAttr := utils.GetValueFromLogs(
		uint32(index), tx.Logs, "cosmos.group.v1.EventCreateGroup", "group_id",
	)
	if groupIDAttr == "" {
		return errors.New("error while getting groupID from tx.Logs")
	}

	groupID, err := strconv.ParseUint(groupIDAttr, 10, 64)
	if err != nil {
		return err
	}

	address := utils.GetValueFromLogs(
		uint32(index), tx.Logs, "cosmos.group.v1.EventCreateGroupPolicy", "address",
	)
	if address == "" {
		return errors.New("error while getting address from tx.Logs")
	}

	decisionPolicy, ok := msg.DecisionPolicy.GetCachedValue().(*group.ThresholdDecisionPolicy)
	if !ok {
		return errors.New("error while parsing decision policy")
	}

	threshold, err := strconv.ParseUint(decisionPolicy.Threshold, 10, 64)
	if err != nil {
		return err
	}

	if err := m.dbTx.SaveGroupWithPolicy(
		types.NewGroupWithPolicy(
			groupID,
			address,
			msg.GroupMetadata,
			msg.GroupPolicyMetadata,
			threshold,
			uint64(decisionPolicy.Windows.VotingPeriod.Seconds()),
			uint64(decisionPolicy.Windows.MinExecutionPeriod.Seconds()),
		),
	); err != nil {
		return err
	}

	members := make([]*types.GroupMember, 0)
	for _, m := range msg.Members {
		weight, err := strconv.ParseUint(m.Weight, 10, 64)
		if err != nil {
			return err
		}

		members = append(members, types.NewGroupMember(m.Address, weight, m.Metadata))
	}
	return m.dbTx.SaveGroupMembers(groupID, members)
}

func (m *Module) handleMsgSubmitProposal(tx *juno.Tx, index int, msg *group.MsgSubmitProposal) error {
	proposalIDAttr := utils.GetValueFromLogs(
		uint32(index), tx.Logs, "cosmos.group.v1.EventSubmitProposal", "proposal_id",
	)
	if proposalIDAttr == "" {
		return errors.New("error while getting proposalID from tx.Logs")
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

	if err := m.dbTx.SaveGroupProposal(
		types.NewGroupProposal(
			proposalID,
			groupID,
			msg.Metadata,
			msg.Proposers[0],
			group.PROPOSAL_STATUS_SUBMITTED.String(),
			group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN.String(),
			utils.SanitizeUTF8(string(msgBytes)),
			tx.Height,
		),
	); err != nil {
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
	proposal, err := m.dbTx.GetGroupProposal(msg.ProposalId)
	if err != nil {
		return err
	} else if proposal.Status != group.PROPOSAL_STATUS_SUBMITTED.String() {
		return errors.New("error while voting - proposal status is not PROPOSAL_STATUS_SUBMITTED")
	}

	timestamp, err := time.Parse(time.RFC3339, tx.Timestamp)
	if err != nil {
		return err
	}

	if err := m.dbTx.SaveGroupProposalVote(
		types.NewGroupProposalVote(
			msg.ProposalId,
			proposal.GroupID,
			msg.Voter,
			msg.Option.String(),
			msg.Metadata,
			timestamp,
		),
	); err != nil {
		return err
	}

	if err := m.updateProposalTallyResult(msg.ProposalId, proposal.GroupID); err != nil {
		return err
	}

	if msg.Exec == group.Exec_EXEC_TRY {
		if err := m.handleMsgExec(tx, index, msg.ProposalId); err != nil {
			return err
		}
	}

	return nil
}

func (m *Module) updateProposalTallyResult(proposalID uint64, groupID uint64) error {
	threshold, err := m.dbTx.GetGroupThreshold(groupID)
	if err != nil {
		return err
	}

	votes, err := m.dbTx.GetGroupProposalVotes(proposalID)
	if err != nil {
		return err
	}

	votesTotal := len(votes)
	votesYes := 0
	for _, v := range votes {
		if v == group.VOTE_OPTION_YES.String() {
			votesYes++
		}
	}

	if votesYes == threshold {
		return m.dbTx.UpdateGroupProposalStatus(
			[]uint64{proposalID}, group.PROPOSAL_STATUS_ACCEPTED.String(),
		)
	}

	totalPower, err := m.dbTx.GetGroupTotalVotingPower(groupID)
	if err != nil {
		return err
	}

	votesRemaining := totalPower - votesTotal
	maxPossibleYesCount := votesYes + votesRemaining
	if maxPossibleYesCount < threshold {
		return m.dbTx.UpdateGroupProposalStatus(
			[]uint64{proposalID}, group.PROPOSAL_STATUS_REJECTED.String(),
		)
	}

	return nil
}

func (m *Module) handleMsgWithdrawProposal(proposalID uint64) error {
	return m.dbTx.UpdateGroupProposalStatus(
		[]uint64{proposalID}, group.PROPOSAL_STATUS_WITHDRAWN.String(),
	)
}

func (m *Module) handleMsgExec(tx *juno.Tx, index int, proposalID uint64) error {
	policy, err := m.dbTx.GetGroupProposalDecisionPolicy(proposalID)
	if err != nil {
		return err
	} else if policy.Status != group.PROPOSAL_STATUS_ACCEPTED.String() {
		return errors.New("error while executing proposal - proposal status is not PROPOSAL_STATUS_ACCEPTED")
	}

	block, err := m.db.GetLastBlock()
	if err != nil {
		return err
	}
	minExecutionPeriod := time.Second * time.Duration(policy.MinExecutionPeriod)
	if policy.SubmitTime.Add(minExecutionPeriod).Before(block.Timestamp) {
		return errors.New("error while executing proposal - min_execution_time has not passed")
	}

	executorResult := utils.GetValueFromLogs(
		uint32(index), tx.Logs, "cosmos.group.v1.EventExec", "result",
	)
	if executorResult == "" {
		return nil
	}

	if err := m.dbTx.UpdateGroupProposalExecutorResult(
		proposalID, executorResult, tx.TxHash,
	); err != nil {
		return err
	}

	if executorResult == "PROPOSAL_EXECUTOR_RESULT_SUCCESS" {
		proposal, err := m.dbTx.GetGroupProposal(proposalID)
		if err != nil {
			return err
		}

		if strings.Contains(proposal.Messages, "MsgUpdateGroup") {
			return m.handleMsgUpdateGroup(proposal)
		}
	}

	return nil
}

func (m *Module) handleMsgUpdateGroup(proposal *dbtypes.GroupProposalRow) error {
	if err := m.dbTx.UpdateGroupProposalStatus(
		[]uint64{proposal.ID}, group.PROPOSAL_STATUS_ABORTED.String(),
	); err != nil {
		return err
	}

	var msgs []*codectypes.Any
	if err := json.Unmarshal([]byte(proposal.Messages), &msgs); err != nil {
		return err
	}

	for _, message := range msgs {
		switch message.TypeUrl {
		case "cosmos.group.v1.MsgUpdateGroupMembers":
			return m.handleMsgUpdateGroupMembers(message)
		case "cosmos.group.v1.MsgUpdateGroupMetadata":
			return m.handleMsgUpgateGroupMetadata(message, proposal)
		case "cosmos.group.v1.MsgUpdateGroupPolicyMetadata":
			return m.handleMsgUpdateGroupPolicyMetadata(message, proposal)
		case "cosmos.group.v1.MsgUpdateGroupPolicyDecisionPolicy":
			return m.handleMsgUpdateGroupDecisionPolicy(message, proposal)
		}
	}

	return nil
}

func (m *Module) handleMsgUpdateGroupMembers(message *codectypes.Any) error {
	var msg group.MsgUpdateGroupMembers
	err := json.Unmarshal(message.Value, &msg)
	if err != nil {
		return err
	}

	updateMembers := make([]*types.GroupMember, 0)
	deleteMembers := make([]string, 0)
	for _, m := range msg.MemberUpdates {
		if m.Weight == "0" {
			deleteMembers = append(deleteMembers, m.Address)
		} else {
			weight, err := strconv.ParseUint(m.Weight, 10, 64)
			if err != nil {
				return err
			}

			member := types.NewGroupMember(m.Address, weight, m.Metadata)
			updateMembers = append(updateMembers, member)
		}
	}

	if err := m.dbTx.SaveGroupMembers(msg.GroupId, updateMembers); err != nil {
		return err
	}

	return m.dbTx.DeleteGroupMembers(deleteMembers, msg.GroupId)
}

func (m *Module) handleMsgUpgateGroupMetadata(message *codectypes.Any, proposal *dbtypes.GroupProposalRow) error {
	var msg group.MsgUpdateGroupMetadata
	if err := json.Unmarshal(message.Value, &msg); err != nil {
		return err
	}

	return m.dbTx.UpdateGroupMetadata(proposal.GroupID, msg.Metadata)
}

func (m *Module) handleMsgUpdateGroupPolicyMetadata(message *codectypes.Any, proposal *dbtypes.GroupProposalRow) error {
	var msg group.MsgUpdateGroupPolicyMetadata
	if err := json.Unmarshal(message.Value, &msg); err != nil {
		return err
	}

	return m.dbTx.UpdateGroupPolicyMetadata(proposal.GroupID, msg.Metadata)
}

func (m *Module) handleMsgUpdateGroupDecisionPolicy(message *codectypes.Any, proposal *dbtypes.GroupProposalRow) error {
	var msg group.MsgUpdateGroupPolicyDecisionPolicy
	if err := json.Unmarshal(message.Value, &msg); err != nil {
		return err
	}

	policy, ok := msg.DecisionPolicy.GetCachedValue().(*group.ThresholdDecisionPolicy)
	if !ok {
		return errors.New("error while parsing decision policy")
	}

	threshold, err := strconv.ParseUint(policy.Threshold, 10, 64)
	if err != nil {
		return err
	}

	return m.dbTx.UpdateGroupPolicy(
		proposal.GroupID,
		types.NewGroupDecisionPolicy(
			proposal.GroupID,
			threshold,
			uint64(policy.Windows.VotingPeriod.Seconds()),
			uint64(policy.Windows.MinExecutionPeriod.Seconds()),
		),
	)
}
