package group

import (
	"encoding/json"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
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
	case *group.MsgWithdrawProposal:
		return m.handleMsgWithdrawProposal(cosmosMsg.ProposalId)
	case *group.MsgVote:
		return m.handleMsgVote(tx, index, cosmosMsg)
	case *group.MsgExec:
		return m.handleMsgExec(tx, index, cosmosMsg)
	}

	return nil
}

func (m *Module) handleMsgCreateGroupWithPolicy(
	tx *juno.Tx,
	index int,
	msg *group.MsgCreateGroupWithPolicy,
) error {
	groupIDAttr, _ := strconv.Unquote(utils.GetValueFromLogs(
		uint32(index),
		tx.Logs,
		"cosmos.group.v1.EventCreateGroup",
		"group_id",
	))
	groupID, _ := strconv.ParseUint(groupIDAttr, 10, 64)

	address, _ := strconv.Unquote(utils.GetValueFromLogs(
		uint32(index),
		tx.Logs,
		"cosmos.group.v1.EventCreateGroupPolicy",
		"address",
	))

	members := make([]*types.GroupMember, 0)
	for _, m := range msg.Members {
		weight, _ := strconv.ParseUint(m.Weight, 10, 64)
		member := types.NewGroupMember(m.Address, weight, m.Metadata)
		members = append(members, member)
	}

	decisionPolicy, _ := msg.DecisionPolicy.GetCachedValue().(*group.ThresholdDecisionPolicy)
	threshold, _ := strconv.ParseUint(decisionPolicy.Threshold, 10, 64)
	timestamp, _ := time.Parse(time.RFC3339, tx.Timestamp)
	votingPeriod := timestamp.Add(decisionPolicy.Windows.VotingPeriod)
	minExecutionPeriod := timestamp.Add(decisionPolicy.Windows.MinExecutionPeriod)

	return m.db.SaveGroupWithPolicy(
		*types.NewGroupWithPolicy(
			groupID,
			address,
			members,
			msg.GroupMetadata,
			msg.GroupPolicyMetadata,
			threshold,
			votingPeriod,
			minExecutionPeriod,
		),
	)
}

func (m *Module) handleMsgSubmitProposal(
	tx *juno.Tx,
	index int,
	msg *group.MsgSubmitProposal,
) error {
	executorResult := group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN.String()
	status := group.PROPOSAL_STATUS_SUBMITTED.String()
	execEvent, _ := strconv.Unquote(utils.GetValueFromLogs(
		uint32(index),
		tx.Logs,
		"cosmos.group.v1.EventExec",
		"result",
	))
	if execEvent != "" {
		executorResult = execEvent
		status = group.PROPOSAL_STATUS_ACCEPTED.String()
	}

	proposalIDAttr, _ := strconv.Unquote(utils.GetValueFromLogs(
		uint32(index),
		tx.Logs,
		"cosmos.group.v1.EventSubmitProposal",
		"proposal_id",
	))
	proposalID, _ := strconv.ParseUint(proposalIDAttr, 10, 64)

	groupID, err := m.db.GetGroupID(msg.GroupPolicyAddress)
	if err != nil {
		return err
	}

	messages, _ := json.Marshal(msg.Messages)
	timestamp, _ := time.Parse(time.RFC3339, tx.Timestamp)

	err = m.db.SaveGroupProposal(
		*types.NewGroupProposal(
			proposalID,
			groupID,
			msg.Metadata,
			msg.Proposers[0],
			timestamp,
			status,
			executorResult,
			utils.SanitizeUTF8(string(messages)),
		),
	)
	if err != nil {
		return err
	}

	if msg.Exec == group.Exec_EXEC_TRY {
		err := m.db.SaveGroupProposalVote(
			*types.NewGroupProposalVote(
				proposalID,
				msg.Proposers[0],
				group.VOTE_OPTION_YES.String(),
				"",
				timestamp,
			),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Module) handleMsgWithdrawProposal(proposalID uint64) error {
	return m.db.UpdateGroupProposalStatus(
		proposalID,
		group.PROPOSAL_STATUS_WITHDRAWN.String(),
	)
}

func (m *Module) handleMsgVote(tx *juno.Tx, index int, msg *group.MsgVote) error {
	timestamp, _ := time.Parse(time.RFC3339, tx.Timestamp)

	err := m.db.SaveGroupProposalVote(
		*types.NewGroupProposalVote(
			msg.ProposalId,
			msg.Voter,
			msg.Option.String(),
			msg.Metadata,
			timestamp,
		),
	)
	if err != nil {
		return err
	}

	executorResult := group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN.String()
	execEvent, _ := strconv.Unquote(utils.GetValueFromLogs(
		uint32(index),
		tx.Logs,
		"cosmos.group.v1.EventExec",
		"result",
	))
	if execEvent != "" {
		executorResult = execEvent
	}

	return m.db.UpdateGroupProposalTallyResult(msg.ProposalId, executorResult)
}

func (m *Module) handleMsgExec(tx *juno.Tx, index int, msg *group.MsgExec) error {
	executorResult, _ := strconv.Unquote(utils.GetValueFromLogs(
		uint32(index),
		tx.Logs,
		"cosmos.group.v1.EventExec",
		"result",
	))

	return m.db.UpdateGroupProposalExecutorResult(msg.ProposalId, executorResult)
}
