package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	dbTypes "github.com/forbole/bdjuno/v4/database/types"
	config "github.com/forbole/bdjuno/v4/integration_tests/set_up"
	"github.com/forbole/bdjuno/v4/integration_tests/types"
	sdkTypes "github.com/forbole/bdjuno/v4/modules/actions/types"
)

var (
	govModule            = "gov"
	expectedProposalID   = "1"
	voteWithNo           = "no"
	voteWithWeightedVote = "yes=0.6,no=0.3,abstain=0.05,no_with_veto=0.05"
	voteOptionNo         = govTypes.VoteOption_name[int32(govTypes.OptionNo)]
	statusDepositPeriod  = govTypes.ProposalStatus_name[int32(govTypes.StatusDepositPeriod)]
	statusVotingPeriod   = govTypes.ProposalStatus_name[int32(govTypes.StatusVotingPeriod)]
	smallDepositAmount   = fmt.Sprintf("%s%s", stringAmount, config.Denom)
	bigDepositAmount     = fmt.Sprintf("%s%s", "50000000000000000000000", config.Denom)
	govAccAddr, _        = config.GetModuleAccountAddress("gov")
	govMsg               = types.MsgSend{
		Type:        "/cosmos.bank.v1beta1.MsgSend",
		FromAddress: govAccAddr,
		ToAddress:   User2,
		Amount: []sdkTypes.Coin{
			{
				Denom:  config.Denom,
				Amount: stringAmount,
			},
		},
	}
	govProposal = types.GovProposal{
		Messages: []types.MsgSend{govMsg},
		Metadata: Metadata,
		Deposit:  smallDepositAmount,
		Title:    Title,
		Summary:  Summary,
	}
)

func TestSubmitGovProposal(t *testing.T) {

	// PREPARE
	proposalFile, err := config.SaveToTempFile(govProposal)
	require.NoError(t, err)
	defer os.Remove(proposalFile)

	require.NoError(t, err)

	args := []string{
		govModule,
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

	// proposal
	var actualProposalFromDB dbTypes.ProposalRow
	err = config.QueryDatabase(`
		SELECT 
			title, 
			description, 
			content, 
			submit_time, 
			deposit_end_time, 
			voting_start_time, 
			voting_end_time,
			proposer_address,
			status,
			metadata FROM proposal
			WHERE id = $1`, expectedProposalID).Scan(
		&actualProposalFromDB.Title,
		&actualProposalFromDB.Description,
		&actualProposalFromDB.Content,
		&actualProposalFromDB.SubmitTime,
		&actualProposalFromDB.DepositEndTime,
		&actualProposalFromDB.VotingStartTime,
		&actualProposalFromDB.VotingEndTime,
		&actualProposalFromDB.Proposer,
		&actualProposalFromDB.Status,
		&actualProposalFromDB.Metadata,
	)
	require.NoError(t, err)

	require.Empty(t, actualProposalFromDB.VotingStartTime)
	require.Empty(t, actualProposalFromDB.VotingEndTime)

	require.NotEmpty(t, actualProposalFromDB.SubmitTime)
	require.NotEmpty(t, actualProposalFromDB.Content)
	require.NotEmpty(t, actualProposalFromDB.DepositEndTime)

	require.Equal(t, User1, actualProposalFromDB.Proposer)
	require.Equal(t, govProposal.Summary, actualProposalFromDB.Description)
	require.Equal(t, govProposal.Metadata, actualProposalFromDB.Metadata)
	require.Equal(t, govProposal.Title, actualProposalFromDB.Title)
	require.Equal(t, statusDepositPeriod, actualProposalFromDB.Status)

	// proposal_deposit
	var actualInitialDepositFromDB dbTypes.DepositRow
	err = config.QueryDatabase(`
		SELECT 
			depositor_address, 
			amount, 
			height, 
			timestamp FROM proposal_deposit
			WHERE proposal_id = $1`, expectedProposalID).Scan(
		&actualInitialDepositFromDB.Depositor,
		&actualInitialDepositFromDB.Amount,
		&actualInitialDepositFromDB.Height,
		&actualInitialDepositFromDB.Timestamp,
	)
	require.NoError(t, err)

	require.NotEmpty(t, actualInitialDepositFromDB.Timestamp)
	require.NotEmpty(t, actualInitialDepositFromDB.Height)

	require.Equal(t, actualInitialDepositFromDB.Amount.ToCoins().GetDenomByIndex(0), config.Denom)
	require.Equal(t, actualInitialDepositFromDB.Amount.ToCoins().AmountOf(config.Denom).String(), stringAmount)
	require.Equal(t, User1, actualInitialDepositFromDB.Depositor)
}

func TestMsgDepositToProposal(t *testing.T) {

	// PREPARE
	newUser := User3

	// Make sure newUser haven't deposited yet
	var existingDepositor string
	err := config.QueryDatabase(`
		SELECT 
			depositor_address FROM proposal_deposit
			WHERE proposal_id = $1 
			AND depositor_address = $2`, expectedProposalID, newUser,
	).Scan(&existingDepositor)
	require.Error(t, err)

	args := []string{
		govModule,
		"deposit",
		expectedProposalID,
		bigDepositAmount,
	}

	// EXECUTE
	result, err := config.ExecuteTxCommand(newUser, args...)
	require.NoError(t, err)

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	// Make sure newUser is now part of the depositors
	err = config.QueryDatabase(`
		SELECT 
			depositor_address FROM proposal_deposit
			WHERE proposal_id = $1 
			AND depositor_address = $2`, expectedProposalID, newUser,
	).Scan(&existingDepositor)
	require.NoError(t, err)
	require.Equal(t, newUser, existingDepositor)
}

func TestProposalIsReadyForVotingAfterDeposit(t *testing.T) {
	var actualStatus string
	err := config.QueryDatabase(`
	SELECT 
		status FROM proposal
		WHERE id = $1 
		AND voting_end_time IS NOT NULL 
		AND voting_start_time IS NOT NULL`, expectedProposalID,
	).Scan(&actualStatus)

	require.NoError(t, err)
	require.Equal(t, actualStatus, statusVotingPeriod)
}

func TestVoteOnProposal(t *testing.T) {
	// PREPARE
	voter := User1
	args := []string{
		govModule,
		"vote",
		expectedProposalID,
		voteWithNo,
	}

	// EXECUTE
	result, err := config.ExecuteTxCommand(voter, args...)
	require.NoError(t, err)

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	// proposal_vote
	var actualVoteFromDB dbTypes.VoteRow
	err = config.QueryDatabase(`
		SELECT 
			voter_address, 
			option, 
			height, 
			timestamp FROM proposal_vote
			WHERE proposal_id = $1`, expectedProposalID).Scan(
		&actualVoteFromDB.Voter,
		&actualVoteFromDB.Option,
		&actualVoteFromDB.Height,
		&actualVoteFromDB.Timestamp,
	)
	require.NoError(t, err)

	require.NotEmpty(t, actualVoteFromDB.Timestamp)
	require.NotEmpty(t, actualVoteFromDB.Height)

	require.Equal(t, voter, actualVoteFromDB.Voter)
	require.Equal(t, voteOptionNo, actualVoteFromDB.Option)
}

func TestWeightedVoteOnProposal(t *testing.T) {
	// PREPARE
	voter := User2
	args := []string{
		govModule,
		"weighted-vote",
		expectedProposalID,
		voteWithWeightedVote,
	}

	// EXECUTE
	result, err := config.ExecuteTxCommand(voter, args...)
	require.NoError(t, err)

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	// proposal_vote_weighted
	// Expecting multiple rows for 1 voter
	var actualVotesFromDB []dbTypes.WeightedVoteRow
	queryResult, err := config.QueryDatabaseMultiRows(`
		SELECT
			DISTINCT ON (option)
			option,
			weight,
			height FROM proposal_vote_weighted
			WHERE proposal_id = $1
			AND voter_address = $2`, expectedProposalID, voter)
	require.NoError(t, err)

	for queryResult.Next() {
		var vote dbTypes.WeightedVoteRow
		err := queryResult.Scan(
			&vote.Option,
			&vote.Weight,
			&vote.Height,
		)
		require.NoError(t, err)
		require.NotEmpty(t, vote.Weight)
		require.NotEmpty(t, vote.Height)

		require.Equal(
			t,
			vote.Option,
			govTypes.VoteOption_name[govTypes.VoteOption_value[vote.Option]],
		)

		actualVotesFromDB = append(actualVotesFromDB, vote)
	}

	// Our weighted votes from the same voter should be 4 distinct
	require.Len(t, actualVotesFromDB, 4)
}
