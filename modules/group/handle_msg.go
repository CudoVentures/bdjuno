package group

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"
	juno "github.com/forbole/juno/v2/types"
)

func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	if len(tx.Logs) == 0 {
		return nil
	}

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
}

func (m *Module) handleMsgCreateGroupWithPolicy(
	tx *juno.Tx, index int, msg *group.MsgCreateGroupWithPolicy,
) error {
	groupIDAttr := utils.GetValueFromLogs(
		uint32(index),
		tx.Logs,
		"cosmos.group.v1.EventCreateGroup",
		"group_id",
	)
	groupID, _ := strconv.ParseUint(groupIDAttr, 10, 64)

	address := utils.GetValueFromLogs(
		uint32(index),
		tx.Logs,
		"cosmos.group.v1.EventCreateGroupPolicy",
		"address",
	)

	decisionPolicy, _ := msg.DecisionPolicy.GetCachedValue().(*group.ThresholdDecisionPolicy)
	threshold, _ := strconv.ParseUint(decisionPolicy.Threshold, 10, 64)

	return m.db.SaveGroupWithPolicy(
		types.NewGroupWithPolicy(
			groupID,
			address,
			msg.Members,
			msg.GroupMetadata,
			msg.GroupPolicyMetadata,
			threshold,
			uint64(decisionPolicy.Windows.VotingPeriod.Seconds()),
			uint64(decisionPolicy.Windows.MinExecutionPeriod.Seconds()),
		),
	)
}

func (m *Module) handleMsgSubmitProposal(
	tx *juno.Tx, index int, msg *group.MsgSubmitProposal,
) error {
	proposalIDAttr := utils.GetValueFromLogs(
		uint32(index),
		tx.Logs,
		"cosmos.group.v1.EventSubmitProposal",
		"proposal_id",
	)
	proposalID, _ := strconv.ParseUint(proposalIDAttr, 10, 64)
	msgBytes, _ := json.Marshal(msg.Messages)

	if err := m.db.SaveGroupProposal(
		types.NewGroupProposal(
			proposalID,
			m.db.GetGroupIDByGroupAddress(msg.GroupPolicyAddress),
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
	timestamp, _ := time.Parse(time.RFC3339, tx.Timestamp)

	if err := m.db.SaveGroupProposalVote(
		types.NewGroupProposalVote(
			msg.ProposalId,
			msg.Voter,
			msg.Option.String(),
			msg.Metadata,
			timestamp,
		),
	); err != nil {
		return err
	}

	if err := m.db.UpdateGroupProposalTallyResult(msg.ProposalId); err != nil {
		return err
	}

	if msg.Exec == group.Exec_EXEC_TRY {
		if err := m.handleMsgExec(tx, index, msg.ProposalId); err != nil {
			return err
		}
	}

	return nil
}

func (m *Module) handleMsgExec(tx *juno.Tx, index int, proposalID uint64) error {
	executorResult := utils.GetValueFromLogs(
		uint32(index),
		tx.Logs,
		"cosmos.group.v1.EventExec",
		"result",
	)
	if executorResult == "" {
		return nil
	}

	if err := m.db.UpdateGroupProposalExecResult(
		proposalID,
		executorResult,
		tx.TxHash,
	); err != nil {
		return err
	}

	if executorResult == "PROPOSAL_EXECUTOR_RESULT_SUCCESS" {
		proposal, err := m.db.GetGroupProposal(proposalID)
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
	if err := m.db.UpdateGroupProposalStatus(
		proposal.ID,
		group.PROPOSAL_STATUS_ABORTED.String(),
	); err != nil {
		return err
	}

	var msgs []*codectypes.Any
	json.Unmarshal([]byte(proposal.Messages), &msgs)

	for _, message := range msgs {
		switch message.TypeUrl {
		case "cosmos.group.v1.MsgUpdateGroupMembers":
			var msg group.MsgUpdateGroupMembers
			_ = json.Unmarshal(message.Value, &msg)
			return m.db.SaveGroupMembers(msg.MemberUpdates, msg.GroupId)
		case "cosmos.group.v1.MsgUpdateGroupMetadata":
			var msg group.MsgUpdateGroupMetadata
			_ = json.Unmarshal(message.Value, &msg)
			return m.db.UpdateGroupMetadata(proposal.GroupID, msg.Metadata)
		case "cosmos.group.v1.MsgUpdateGroupPolicyMetadata":
			var msg group.MsgUpdateGroupPolicyMetadata
			_ = json.Unmarshal(message.Value, &msg)
			return m.db.UpdateGroupPolicyMetadata(proposal.GroupID, msg.Metadata)
		case "cosmos.group.v1.MsgUpdateGroupPolicyDecisionPolicy":
			var msg group.MsgUpdateGroupPolicyDecisionPolicy
			_ = json.Unmarshal(message.Value, &msg)
			decisionPolicy, _ := msg.DecisionPolicy.
				GetCachedValue().(*group.ThresholdDecisionPolicy)
			return m.db.UpdateGroupPolicy(proposal.GroupID, decisionPolicy)
		}
	}

	return nil
}

func (m *Module) handleMsgWithdrawProposal(proposalID uint64) error {
	return m.db.UpdateGroupProposalStatus(proposalID, group.PROPOSAL_STATUS_WITHDRAWN.String())
}
