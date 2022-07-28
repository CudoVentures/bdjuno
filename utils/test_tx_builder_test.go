package utils

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/stretchr/testify/require"
)

var (
	num           = uint64(1)
	str           = "1"
	index         = uint32(0)
	resultDefault = group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN
)

func TestTxBuilder_Build(t *testing.T) {
	timestamp := time.Now()
	tx, err := NewTestTx(timestamp, str, int64(num)).WithEventCreateGroup(num, str).WithEventSubmitProposal(num).WithEventExec(resultDefault).WithEventVote().WithEventWithdrawProposal().Build()
	require.NoError(t, err)

	expectedEventCount := 6
	actualEventCount := len(tx.Logs[0].Events)
	require.Equal(t, expectedEventCount, actualEventCount)

	groupID := utils.GetValueFromLogs(index, tx.Logs, "cosmos.group.v1.EventCreateGroup", "group_id")
	require.Equal(t, str, groupID)

	address := utils.GetValueFromLogs(index, tx.Logs, "cosmos.group.v1.EventCreateGroupPolicy", "address")
	require.Equal(t, str, address)

	proposalID := utils.GetValueFromLogs(index, tx.Logs, "cosmos.group.v1.EventSubmitProposal", "proposal_id")
	require.Equal(t, str, proposalID)

	voteEvent := utils.GetValueFromLogs(index, tx.Logs, "cosmos.group.v1.EventVote", "proposal_id")
	require.Equal(t, str, voteEvent)

	executorResult := utils.GetValueFromLogs(index, tx.Logs, "cosmos.group.v1.EventExec", "result")
	require.Equal(t, resultDefault.String(), executorResult)

	withdrawEvent := utils.GetValueFromLogs(uint32(index), tx.Logs, "cosmos.group.v1.EventWithdrawProposal", "proposal_id")
	require.Equal(t, str, withdrawEvent)

	expectedTimestamp := timestamp.Format(time.RFC3339)
	actualTimestamp := tx.Timestamp
	require.Equal(t, expectedTimestamp, actualTimestamp)

	actualTxHash := tx.TxHash
	require.Equal(t, str, actualTxHash)

	actualHeight := tx.Height
	require.Equal(t, int64(num), actualHeight)
}

func TestTxBuilder_Error(t *testing.T) {
	timestamp := time.Now()
	_, err := NewTestTx(timestamp, str, int64(num)).WithEventCreateGroup(1, "").WithEventSubmitProposal(1).WithEventExec(group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN).WithEventVote().WithEventWithdrawProposal().Build()
	expectedError := "error while building testTx: error while building testTx: empty group address"
	require.Equal(t, expectedError, err.Error())
}
