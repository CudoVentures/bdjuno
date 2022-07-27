package utils

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/stretchr/testify/require"
)

func TestGroup_TestTxBuilder(t *testing.T) {
	timestamp := time.Now()
	tx, err := NewTestTx(timestamp).WithEventCreateGroup(1, "1").WithEventSubmitProposal(1).WithEventExec(group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN).WithEventVote().WithEventWithdrawProposal().Build()
	require.NoError(t, err)

	expectedLogs := "[{\"events\":[{\"type\":\"cosmos.group.v1.EventCreateGroup\",\"attributes\":[{\"key\":\"group_id\",\"value\":\"\\\"1\\\"\"}]},{\"type\":\"cosmos.group.v1.EventCreateGroupPolicy\",\"attributes\":[{\"key\":\"address\",\"value\":\"\\\"1\\\"\"}]},{\"type\":\"cosmos.group.v1.EventExec\",\"attributes\":[{\"key\":\"proposal_id\",\"value\":\"\\\"0\\\"\"},{\"key\":\"result\",\"value\":\"\\\"PROPOSAL_EXECUTOR_RESULT_NOT_RUN\\\"\"},{\"key\":\"logs\",\"value\":\"\\\"1\\\"\"}]},{\"type\":\"cosmos.group.v1.EventSubmitProposal\",\"attributes\":[{\"key\":\"proposal_id\",\"value\":\"\\\"1\\\"\"}]},{\"type\":\"cosmos.group.v1.EventVote\",\"attributes\":[{\"key\":\"proposal_id\",\"value\":\"\\\"0\\\"\"}]},{\"type\":\"cosmos.group.v1.EventWithdrawProposal\",\"attributes\":[{\"key\":\"proposal_id\",\"value\":\"\\\"0\\\"\"}]}]}]"
	actualLogs := tx.Logs.String()
	require.Equal(t, expectedLogs, actualLogs)

	expectedTimestamp := timestamp.Format(time.RFC3339)
	actualTimestamp := tx.Timestamp
	require.Equal(t, expectedTimestamp, actualTimestamp)

	expectedTxHash := "1"
	actualTxHash := tx.TxHash
	require.Equal(t, expectedTxHash, actualTxHash)

	expectedHeight := int64(1)
	actualHeight := tx.Height
	require.Equal(t, expectedHeight, actualHeight)
}

func TestGroup_TestTxBuilder_Error(t *testing.T) {
	timestamp := time.Now()
	_, err := NewTestTx(timestamp).WithEventCreateGroup(1, "").WithEventSubmitProposal(1).WithEventExec(group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN).WithEventVote().WithEventWithdrawProposal().Build()
	expectedError := "error while building testTx: error while building testTx: empty group address"
	require.Equal(t, expectedError, err.Error())
}
