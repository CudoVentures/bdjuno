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
		return nil
	case *group.MsgExec:
		return nil
	case *group.MsgVote:
		return nil
	}

	return nil
}

func (m *Module) handleMsgCreateGroupWithPolicy(tx *juno.Tx, index int, msg *group.MsgCreateGroupWithPolicy) error {
	groupIdAttr, _ := strconv.Unquote(utils.GetValueFromLogs(
		uint32(index),
		tx.Logs,
		"cosmos.group.v1.EventCreateGroup",
		"group_id",
	))
	groupID, _ := strconv.ParseUint(groupIdAttr, 10, 64)

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

	return m.db.SaveGroupWithPolicy(
		*types.NewGroupWithPolicy(
			groupID,
			address,
			members,
			msg.GroupMetadata,
			msg.GroupPolicyMetadata,
			threshold,
			uint64(decisionPolicy.Windows.VotingPeriod.Seconds()),
			uint64(decisionPolicy.Windows.MinExecutionPeriod.Seconds()),
		),
	)
}

func (m *Module) handleMsgSubmitProposal(tx *juno.Tx, index int, msg *group.MsgSubmitProposal) error {
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

	proposalIdAttr, _ := strconv.Unquote(utils.GetValueFromLogs(
		uint32(index),
		tx.Logs,
		"cosmos.group.v1.EventSubmitProposal",
		"proposal_id",
	))
	proposalID, _ := strconv.ParseUint(proposalIdAttr, 10, 64)

	groupID := m.db.GetGroupId(msg.GroupPolicyAddress)

	messages, _ := json.Marshal(msg.Messages)

	timestamp, _ := time.Parse(time.RFC3339, tx.Timestamp)

	err := m.db.SaveGroupProposal(
		*types.NewGroupProposal(
			proposalID,
			groupID,
			msg.Metadata,
			msg.Proposers[0],
			timestamp,
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
