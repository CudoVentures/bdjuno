package test

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	groupTypes "github.com/cosmos/cosmos-sdk/x/group"
	dbtypes "github.com/forbole/bdjuno/v4/database/types"
	config "github.com/forbole/bdjuno/v4/integration_tests/set_up"
	"github.com/forbole/bdjuno/v4/integration_tests/types"
	sdkTypes "github.com/forbole/bdjuno/v4/modules/actions/types"
)

var (
	groupModule  = "group"
	stringAmount = "1000000000000000000"
	groupID      = 1
)

func TestCreateGroupWithPolicy(t *testing.T) {

	// PREPARE
	policy := types.DecisionPolicy{
		Type:      "/cosmos.group.v1.ThresholdDecisionPolicy",
		Threshold: "1",
		Windows: types.Window{
			VotingPeriod:       "120h",
			MinExecutionPeriod: "0s",
		},
	}

	policyFile, err := config.SaveToTempFile(policy)
	if err != nil {
		t.Fatalf("Error saving policy to temp file: %v", err)
	}
	defer os.Remove(policyFile)

	members := types.GroupMembers{
		Members: []types.Member{
			{
				Address:  User1,
				Weight:   String1,
				Metadata: Metadata,
			},
			{
				Address:  User2,
				Weight:   String1,
				Metadata: Metadata,
			},
		},
	}
	membersFile, err := config.SaveToTempFile(members)
	if err != nil {
		t.Fatalf("Error saving members to temp file: %v", err)
	}
	defer os.Remove(membersFile)

	args := []string{
		groupModule,
		"create-group-with-policy",
		User1,
		Metadata,
		Metadata,
		membersFile,
		policyFile,
		"--group-policy-as-admin",
	}

	// EXECUTE
	result, err := config.ExecuteTxCommand(User1, args...)
	require.NoError(t, err)

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	// assert group records
	var actualGroup dbtypes.GroupRow
	err = config.QueryDatabase(`SELECT * FROM group_with_policy WHERE id = $1`, groupID).Scan(
		&actualGroup.ID,
		&actualGroup.Address,
		&actualGroup.GroupMetadata,
		&actualGroup.PolicyMetadata,
		&actualGroup.Threshold,
		&actualGroup.VotingPeriod,
		&actualGroup.MinExecutionPeriod,
	)
	require.NoError(t, err)

	expectedThreshold, err := strconv.ParseUint(policy.Threshold, 10, 64)
	require.NoError(t, err)
	require.NotEmpty(t, actualGroup.Address)
	require.Equal(t, uint64(1), actualGroup.ID)
	require.Equal(t, Metadata, actualGroup.GroupMetadata)
	require.Equal(t, Metadata, actualGroup.PolicyMetadata)
	require.Equal(t, expectedThreshold, actualGroup.Threshold)
	require.NotNil(t, actualGroup.MinExecutionPeriod)
	require.NotEmpty(t, actualGroup.VotingPeriod)

	// assert member records
	expectedMember := members.Members[0]
	var actualMember dbtypes.GroupMemberRow
	err = config.QueryDatabase(`SELECT * FROM group_member where group_id = $1 AND address = $2`, groupID, User1).Scan(
		&actualMember.GroupID,
		&actualMember.Address,
		&actualMember.Weight,
		&actualMember.Metadata,
		&actualMember.AddTime,
	)
	require.NoError(t, err)

	expecteWeight, err := strconv.ParseUint(expectedMember.Weight, 10, 64)
	require.NoError(t, err)
	require.Equal(t, expectedMember.Address, actualMember.Address)
	require.Equal(t, uint64(groupID), actualMember.GroupID)
	require.Equal(t, expectedMember.Metadata, actualMember.Metadata)
	require.Equal(t, expecteWeight, actualMember.Weight)
}

func TestSubmitGroupProposal(t *testing.T) {
	var groupAddress string
	err := config.QueryDatabase(`SELECT address FROM group_with_policy WHERE id = $1`, groupID).Scan(&groupAddress)
	require.NoError(t, err)

	// PREPARE
	msgSend := types.MsgSend{
		Type:        "/cosmos.bank.v1beta1.MsgSend",
		FromAddress: groupAddress,
		ToAddress:   User2,
		Amount: []sdkTypes.Coin{
			{
				Denom:  config.Denom,
				Amount: stringAmount,
			},
		},
	}
	metadataEncoded := base64.StdEncoding.EncodeToString([]byte(Metadata))
	proposal := types.GroupProposal{
		GroupPolicyAddress: groupAddress,
		Messages:           []types.MsgSend{msgSend},
		Metadata:           metadataEncoded,
		Proposers:          []string{User1},
		Title:              Metadata,
		Summary:            Metadata,
	}
	proposalFile, err := config.SaveToTempFile(proposal)
	require.NoError(t, err)
	defer os.Remove(proposalFile)
	args := []string{
		groupModule,
		"submit-proposal",
		proposalFile,
	}

	// EXECUTE
	result, err := config.ExecuteTxCommand(User1, args...)
	require.NoError(t, err)

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	var actualProposal dbtypes.GroupProposalRow
	err = config.QueryDatabase(`SELECT proposer, Metadata, messages FROM group_proposal where id = 1`).Scan(
		&actualProposal.Proposer,
		&actualProposal.ProposalMetadata,
		&actualProposal.Messages,
	)
	require.NoError(t, err)
	var actualProposalMsgs []types.MsgSend
	err = json.Unmarshal([]byte(actualProposal.Messages), &actualProposalMsgs)

	require.NoError(t, err)
	require.Equal(t, proposal.Proposers[0], actualProposal.Proposer)
	require.Equal(t, proposal.Metadata, actualProposal.ProposalMetadata)
	require.Equal(t, proposal.Messages, actualProposalMsgs)
}

func TestSubmitVoteToGroupProposal(t *testing.T) {
	// PREPARE

	// make sure the proposal have initial status column before voting
	var actualProposalStatus string
	err := config.QueryDatabase(`SELECT status FROM group_proposal where id = 1`).Scan(&actualProposalStatus)
	require.NoError(t, err)
	require.Equal(t,
		groupTypes.ProposalStatus_name[int32(groupTypes.PROPOSAL_STATUS_SUBMITTED)],
		actualProposalStatus,
	)

	voteOption := groupTypes.VoteOption_name[int32(groupTypes.VOTE_OPTION_YES)]
	args := []string{
		groupModule,
		"vote",
		"1",
		User1,
		voteOption,
		Metadata,
	}

	// EXECUTE
	result, err := config.ExecuteTxCommand(User1, args...)
	require.NoError(t, err)

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	// make sure the proposal have status column changed after voting
	err = config.QueryDatabase(`SELECT status FROM group_proposal where id = 1`).Scan(&actualProposalStatus)
	require.NoError(t, err)
	require.Equal(t,
		groupTypes.ProposalStatus_name[int32(groupTypes.PROPOSAL_STATUS_ACCEPTED)],
		actualProposalStatus,
	)

	// make sure the voter has been reflected in the DB
	var expectedVote dbtypes.GroupProposalVoteRow
	err = config.QueryDatabase(`SELECT voter, vote_option from group_proposal_vote WHERE proposal_id = 1`).Scan(
		&expectedVote.Voter,
		&expectedVote.VoteOption,
	)
	require.NoError(t, err)
	require.Equal(t, voteOption, expectedVote.VoteOption)
	require.Equal(t, User1, expectedVote.Voter)
}

func TestExecuteAcceptedGroupProposal(t *testing.T) {
	// PREPARE

	// make sure the proposal have initial execution status
	var actualProposalExecutionStatus string
	err := config.QueryDatabase(`SELECT executor_result FROM group_proposal where id = 1`).Scan(&actualProposalExecutionStatus)
	require.NoError(t, err)
	require.Equal(t,
		groupTypes.ProposalExecutorResult_name[int32(groupTypes.PROPOSAL_EXECUTOR_RESULT_NOT_RUN)],
		actualProposalExecutionStatus,
	)

	args := []string{
		groupModule,
		"exec",
		"1",
	}

	// EXECUTE
	result, err := config.ExecuteTxCommand(User1, args...)
	require.NoError(t, err)

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	// make sure the proposal is updated in DB

	var dbStatuses types.ProposalExecStatuses
	err = config.QueryDatabase(`SELECT executor_result, executor, execution_time, execution_log FROM group_proposal where id = 1`).Scan(
		&dbStatuses.ExecutorResult,
		&dbStatuses.Executor,
		&dbStatuses.ExecutionTime,
		&dbStatuses.ExecutionLog,
	)
	require.NoError(t, err)
	require.Equal(t, User1, dbStatuses.Executor)
	require.NotEmpty(t, dbStatuses.ExecutionTime)
	require.NotEmpty(t, dbStatuses.ExecutionLog)
	// We expect to be failed since the group admin, which is the group itself, have no balance to fulfill the msg
	require.Equal(t, groupTypes.ProposalExecutorResult_name[int32(groupTypes.PROPOSAL_EXECUTOR_RESULT_FAILURE)], dbStatuses.ExecutorResult)
}
