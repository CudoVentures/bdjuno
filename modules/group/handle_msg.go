package group

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/forbole/bdjuno/v4/database"
	dbtypes "github.com/forbole/bdjuno/v4/database/types"
	"github.com/forbole/bdjuno/v4/types"
	"github.com/forbole/bdjuno/v4/utils"
	juno "github.com/forbole/juno/v5/types"
)

func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	if len(tx.Logs) == 0 {
		return nil
	}

	return m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		switch cosmosMsg := msg.(type) {
		case *group.MsgCreateGroupWithPolicy:
			return m.handleMsgCreateGroupWithPolicy(dbTx, tx, index, cosmosMsg)
		case *group.MsgSubmitProposal:
			return m.handleMsgSubmitProposal(dbTx, tx, index, cosmosMsg)
		case *group.MsgVote:
			return m.handleMsgVote(dbTx, tx, index, cosmosMsg)
		case *group.MsgExec:
			return m.handleMsgExec(dbTx, tx, index, cosmosMsg)
		case *group.MsgWithdrawProposal:
			return m.handleMsgWithdrawProposal(dbTx, tx, index, cosmosMsg.ProposalId)
		}

		return nil
	})

}

func (m *Module) handleMsgCreateGroupWithPolicy(dbTx *database.DbTx, tx *juno.Tx, index int, msg *group.MsgCreateGroupWithPolicy) error {
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

	if err := dbTx.SaveGroup(group); err != nil {
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

	timestamp, err := time.Parse(time.RFC3339, tx.Timestamp)
	if err != nil {
		return err
	}

	return dbTx.SaveMembers(groupID, members, timestamp)
}

func (m *Module) handleMsgSubmitProposal(dbTx *database.DbTx, tx *juno.Tx, index int, msg *group.MsgSubmitProposal) error {
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

	groupID, err := dbTx.GetGroupIDByGroupAddress(msg.GroupPolicyAddress)
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

	memberCount, err := dbTx.GetGroupMemberCount(groupID)
	if err != nil {
		return err
	}

	proposal := types.NewGroupProposal(proposalID, groupID, msg.Metadata, msg.Proposers[0], status, result, msgs, tx.Height, timestamp, memberCount)

	if err := dbTx.SaveProposal(proposal); err != nil {
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

		if err := m.handleMsgVote(dbTx, tx, index, &msgVote); err != nil {
			return err
		}
	}

	return nil
}

func (m *Module) handleMsgVote(dbTx *database.DbTx, tx *juno.Tx, index int, msg *group.MsgVote) error {
	event := utils.GetValueFromLogs(uint32(index), tx.Logs, "cosmos.group.v1.EventVote", "proposal_id")
	if event == "" {
		return errors.New("error while getting EventVote")
	}

	proposal, err := dbTx.GetProposal(msg.ProposalId)
	if err != nil {
		return err
	}

	timestamp, err := time.Parse(time.RFC3339, tx.Timestamp)
	if err != nil {
		return err
	}

	vote := types.NewProposalVote(msg.ProposalId, proposal.GroupID, msg.Voter, msg.Option.String(), msg.Metadata, timestamp)
	if err := dbTx.SaveProposalVote(vote); err != nil {
		return err
	}

	status, err := m.updateProposalStatus(dbTx, msg.ProposalId, proposal.GroupID)
	if err != nil {
		return err
	}

	if msg.Exec == group.Exec_EXEC_TRY && status == group.PROPOSAL_STATUS_ACCEPTED {
		if err := m.handleMsgExec(dbTx, tx, index, &group.MsgExec{ProposalId: msg.ProposalId, Executor: msg.Voter}); err != nil {
			return err
		}
	}

	return nil
}

func (m *Module) updateProposalStatus(dbTx *database.DbTx, proposalID uint64, groupID uint64) (group.ProposalStatus, error) {
	votes, err := dbTx.GetProposalVotes(proposalID)
	if err != nil {
		return 0, err
	}

	votesYes := 0
	for _, v := range votes {
		if v == group.VOTE_OPTION_YES.String() {
			votesYes++
		}
	}

	threshold, err := dbTx.GetGroupThreshold(groupID)
	if err != nil {
		return 0, err
	}

	if votesYes >= threshold {
		err := dbTx.UpdateProposalStatus(proposalID, group.PROPOSAL_STATUS_ACCEPTED.String())
		return group.PROPOSAL_STATUS_ACCEPTED, err
	}

	totalPower, err := dbTx.GetGroupTotalVotingPower(groupID)
	if err != nil {
		return 0, err
	}

	// the real threshold of the policy is `min(threshold,total_weight)`. If
	// the group member weights changes (member leaving, member weight update)
	// and the threshold doesn't, we can end up with threshold > total_weight.
	// In this case, as long as everyone votes yes (in which case
	// `yesCount`==`realThreshold`), then the proposal still passes.
	votesRemaining := totalPower - len(votes)
	maxPossibleYesCount := votesYes + votesRemaining
	if maxPossibleYesCount < int(math.Min(float64(threshold), float64(totalPower))) {
		err := dbTx.UpdateProposalStatus(proposalID, group.PROPOSAL_STATUS_REJECTED.String())
		return group.PROPOSAL_STATUS_REJECTED, err
	}

	return 0, nil
}

func (m *Module) handleMsgExec(dbTx *database.DbTx, tx *juno.Tx, index int, msg *group.MsgExec) error {
	executorResult := utils.GetValueFromLogs(uint32(index), tx.Logs, "cosmos.group.v1.EventExec", "result")
	if executorResult == "" {
		return errors.New("error while getting EventExec")
	}

	executionLog := utils.GetValueFromLogs(uint32(index), tx.Logs, "cosmos.group.v1.EventExec", "logs")

	timestamp, err := time.Parse(time.RFC3339, tx.Timestamp)
	if err != nil {
		return err
	}

	executionResult := types.NewExecutionResult(msg.ProposalId, executorResult, msg.Executor, timestamp, executionLog, tx.TxHash)
	if err := dbTx.UpdateProposalExecutorResult(executionResult); err != nil {
		return err
	}

	proposal, err := dbTx.GetProposal(msg.ProposalId)
	if err != nil {
		return err
	}

	isSuccess := executorResult == group.PROPOSAL_EXECUTOR_RESULT_SUCCESS.String()
	if isSuccess && strings.Contains(proposal.Messages, "MsgUpdateGroup") {
		return m.handleMsgUpdateGroup(dbTx, tx, proposal)
	}

	return nil

}

func (m *Module) handleMsgUpdateGroup(dbTx *database.DbTx, tx *juno.Tx, proposal *dbtypes.GroupProposalRow) error {
	if proposal.ExecutorResult != group.PROPOSAL_EXECUTOR_RESULT_SUCCESS.String() {
		return errors.New("error while executing handleMsgUpdateGroup")
	}

	abort := group.PROPOSAL_STATUS_ABORTED.String()
	if err := dbTx.UpdateActiveProposalStatusesByGroup(proposal.GroupID, abort); err != nil {
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

			timestamp, err := time.Parse(time.RFC3339, tx.Timestamp)
			if err != nil {
				return err
			}

			if err := dbTx.SaveMembers(msg.GroupID, msg.MemberUpdates, timestamp); err != nil {
				return err
			}
		case "/cosmos.group.v1.MsgUpdateGroupMetadata":
			var msg types.MsgUpdateGroupMetadata
			if err := json.Unmarshal(msgs[i], &msg); err != nil {
				return err
			}

			if err := dbTx.UpdateGroupMetadata(proposal.GroupID, msg.Metadata); err != nil {
				return err
			}
		case "/cosmos.group.v1.MsgUpdateGroupPolicyMetadata":
			var msg group.MsgUpdateGroupPolicyMetadata
			if err := json.Unmarshal(msgs[i], &msg); err != nil {
				return err
			}

			if err := dbTx.UpdateGroupPolicyMetadata(proposal.GroupID, msg.Metadata); err != nil {
				return err
			}
		case "/cosmos.group.v1.MsgUpdateGroupPolicyDecisionPolicy":
			var msg types.MsgUpdateDecisionPolicy
			if err := json.Unmarshal(msgs[i], &msg); err != nil {
				return err
			}

			votingPeriod, err := time.ParseDuration(msg.DecisionPolicy.Windows.VotingPeriod)
			if err != nil {
				return err
			}
			votingPeriodUint := uint64(votingPeriod.Seconds())

			minExecutionPeriod, err := time.ParseDuration(msg.DecisionPolicy.Windows.MinExecutionPeriod)
			if err != nil {
				return err
			}
			minExecutionPeriodUint := uint64(minExecutionPeriod.Seconds())

			if err := dbTx.UpdateDecisionPolicy(proposal.GroupID, msg.DecisionPolicy.Threshold, votingPeriodUint, minExecutionPeriodUint); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Module) handleMsgWithdrawProposal(dbTx *database.DbTx, tx *juno.Tx, index int, proposalID uint64) error {
	event := utils.GetValueFromLogs(uint32(index), tx.Logs, "cosmos.group.v1.EventWithdrawProposal", "proposal_id")
	if event == "" {
		return errors.New("error while getting EventWithdraw")
	}

	return dbTx.UpdateProposalStatus(proposalID, group.PROPOSAL_STATUS_WITHDRAWN.String())
}
