package group

import (
	"encoding/json"
	"testing"
	"time"

	abcitypes "github.com/tendermint/tendermint/abci/types"

	juno "github.com/forbole/juno/v2/types"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/utils"
)

type GroupModuleTestSuite struct {
	suite.Suite
	module *Module
}

func TestGroupModuleTestSuite(t *testing.T) {
	suite.Run(t, new(GroupModuleTestSuite))
}

func (suite *GroupModuleTestSuite) SetupTest() {
	db, cdc := utils.NewTestDb(&suite.Suite, "groupTest")
	suite.module = NewModule(cdc, db)
}

func (suite *GroupModuleTestSuite) TestGroup_MsgCreateGroupWithPolicy() {
	tx := suite.newTestTx("1", "", 1, "1", 0, 0)

	decisionPolicy, err := codectypes.NewAnyWithValue(
		group.NewThresholdDecisionPolicy("1", time.Hour, 0),
	)
	suite.Require().NoError(err)

	msg := group.MsgCreateGroupWithPolicy{
		Admin: "admin",
		Members: []group.MemberRequest{
			{Address: "1", Weight: "1", Metadata: "1"},
		},
		GroupMetadata:       "1",
		GroupPolicyMetadata: "1",
		GroupPolicyAsAdmin:  true,
		DecisionPolicy:      decisionPolicy,
	}

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	var groupRows []dbtypes.GroupWithPolicyRow
	err = suite.module.db.Sqlx.Select(
		&groupRows,
		`SELECT * FROM group_with_policy where id = 1`,
	)
	suite.Require().NoError(err)
	suite.Require().Len(groupRows, 1)
	suite.Require().Equal(dbtypes.GroupWithPolicyRow{
		ID:                 1,
		Address:            "1",
		GroupMetadata:      "1",
		PolicyMetadata:     "1",
		Threshold:          1,
		VotingPeriod:       uint64(time.Hour.Seconds()),
		MinExecutionPeriod: 0,
	}, groupRows[0])

	var memberRows []dbtypes.GroupMemberRow
	err = suite.module.db.Sqlx.Select(&memberRows,
		`SELECT * FROM group_member where group_id = 1`,
	)
	suite.Require().NoError(err)
	suite.Require().Len(memberRows, 1)
	suite.Require().Equal(dbtypes.GroupMemberRow{
		Address:        "1",
		GroupID:        1,
		Weight:         1,
		MemberMetadata: "1",
	}, memberRows[0])
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgSubmitProposal() {
	suite.insertTestData(1, time.Now())
	tx := suite.newTestTx("1", "", 0, "", 1, 0)
	msg := suite.newTestMsgProposal(0)

	err := suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	expectedMsg :=
		"[{\"@type\": \"/cosmos.group.v1.MsgUpdateGroupMetadata\", \"admin\": \"1\", \"group_id\": 1, \"metadata\": \"\"}]"

	var proposalRows []dbtypes.GroupProposalRow
	err = suite.module.db.Sqlx.Select(
		&proposalRows,
		`SELECT * FROM group_proposal where group_id = 1`,
	)
	suite.Require().NoError(err)
	suite.Require().Len(proposalRows, 1)
	suite.Require().Equal(dbtypes.GroupProposalRow{
		ID:               1,
		GroupID:          1,
		ProposalMetadata: "",
		Proposer:         "1",
		Status:           group.PROPOSAL_STATUS_SUBMITTED.String(),
		ExecutorResult:   group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN.String(),
		Messages:         expectedMsg,
		TxHash:           dbtypes.ToNullString(""),
		BlockHeight:      1,
	}, proposalRows[0])

	suite.assertVotesCount(0)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgSubmitProposal_TryExec() {
	suite.insertTestData(1, time.Now())
	tx := suite.newTestTx("1", time.Now().Format(time.RFC3339), 1, "", 1, 0)
	msg := suite.newTestMsgProposal(group.Exec_EXEC_TRY)

	err := suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	suite.assertVotesCount(1)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgVote_OnlySubmittedStatus() {
	suite.insertTestData(1, time.Now())
	suite.insertTestProposal(group.PROPOSAL_STATUS_REJECTED)
	tx := suite.newTestTx("1", "", 0, "", 0, 0)
	msg := suite.newTestMsgVote(0, 0)

	err := suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NotNil(err)
	suite.Require().Equal("error while voting - proposal status is not PROPOSAL_STATUS_SUBMITTED", err.Error())

	suite.assertVotesCount(0)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgVote_UpdateTallyAccepted_TryExec() {
	timestamp := time.Date(2022, time.January, 1, 1, 1, 1, 0, time.FixedZone("", 0))
	suite.insertTestData(1, timestamp)
	suite.insertTestProposal(group.PROPOSAL_STATUS_SUBMITTED)
	tx := suite.newTestTx("1", timestamp.Format(time.RFC3339), 0, "", 0, group.PROPOSAL_EXECUTOR_RESULT_SUCCESS)
	msg := suite.newTestMsgVote(group.VOTE_OPTION_YES, group.Exec_EXEC_TRY)

	err := suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	var voteRows []dbtypes.GroupProposalVoteRow
	err = suite.module.db.Sqlx.Select(
		&voteRows,
		`SELECT * FROM group_proposal_vote where proposal_id = 1`,
	)
	suite.Require().NoError(err)
	suite.Require().Len(voteRows, 1)
	suite.Require().Equal(dbtypes.GroupProposalVoteRow{
		ProposalID:   1,
		GroupID:      1,
		Voter:        "1",
		VoteOption:   group.VOTE_OPTION_YES.String(),
		VoteMetadata: "1",
		SubmitTime:   timestamp,
	}, voteRows[0])

	var proposalStatus string
	var executorResult string
	err = suite.module.db.Sqlx.QueryRow(
		`SELECT status, executor_result from group_proposal WHERE id = 1`,
	).Scan(&proposalStatus, &executorResult)
	suite.Require().NoError(err)
	suite.Require().Equal(group.PROPOSAL_STATUS_ACCEPTED.String(), proposalStatus)
	suite.Require().Equal(group.PROPOSAL_EXECUTOR_RESULT_SUCCESS.String(), executorResult)
}

func (suite *GroupModuleTestSuite) TestGroup_UpdateProposalTallyResult_Rejected() {
	suite.insertTestData(1, time.Now())
	suite.insertTestProposal(group.PROPOSAL_STATUS_SUBMITTED)
	tx := suite.newTestTx("1", time.Now().Format(time.RFC3339), 0, "", 0, 0)
	msg := suite.newTestMsgVote(group.VOTE_OPTION_NO, 0)

	err := suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	var voteOption string
	err = suite.module.db.Sqlx.QueryRow(
		`SELECT vote_option from group_proposal_vote WHERE proposal_id = 1`,
	).Scan(&voteOption)
	suite.Require().NoError(err)
	suite.Require().Equal(group.VOTE_OPTION_NO.String(), voteOption)

	var proposalStatus string
	err = suite.module.db.Sqlx.QueryRow(
		`SELECT status from group_proposal WHERE id = 1`,
	).Scan(&proposalStatus)
	suite.Require().NoError(err)
	suite.Require().Equal(group.PROPOSAL_STATUS_REJECTED.String(), proposalStatus)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgWithdrawProposal() {

}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgExec_OnlyAcceptedStatus() {

}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgExec_OnlyPassedMinExecutionTime() {

}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgUpdateGroupMembers() {

}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgUpgateGroupMetadata() {

}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgUpdateGroupPolicyMetadata() {

}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgUpdateGroupDecisionPolicy() {

}

// todo test periodic operation

func (suite *GroupModuleTestSuite) newTestTx(
	txHash string,
	timestamp string,
	groupID uint64,
	groupAddress string,
	proposalID uint64,
	executorResult group.ProposalExecutorResult,
) *juno.Tx {
	events := make([]abcitypes.Event, 0)

	if groupID != 0 {
		eventCreateGroup, err := sdk.TypedEventToEvent(
			&group.EventCreateGroup{GroupId: groupID},
		)
		suite.Require().NoError(err)

		events = append(events, abcitypes.Event(eventCreateGroup))
	}

	if groupAddress != "" {
		eventCreateGroupPolicy, err := sdk.TypedEventToEvent(
			&group.EventCreateGroupPolicy{Address: groupAddress},
		)
		suite.Require().NoError(err)

		events = append(events, abcitypes.Event(eventCreateGroupPolicy))
	}

	if proposalID != 0 {
		eventSubmitProposal, err := sdk.TypedEventToEvent(
			&group.EventSubmitProposal{ProposalId: proposalID},
		)
		suite.Require().NoError(err)

		events = append(events, abcitypes.Event(eventSubmitProposal))
	}

	if executorResult != group.PROPOSAL_EXECUTOR_RESULT_UNSPECIFIED {
		eventExec, err := sdk.TypedEventToEvent(&group.EventExec{
			Result: executorResult,
		})
		suite.Require().NoError(err)

		events = append(events, abcitypes.Event(eventExec))
	}

	txLog := sdk.ABCIMessageLogs{
		{MsgIndex: 0, Events: sdk.StringifyEvents(events)},
	}

	txResponse := sdk.TxResponse{
		TxHash:    txHash,
		Logs:      txLog,
		Timestamp: timestamp,
		Height:    1,
	}

	return &juno.Tx{TxResponse: &txResponse}
}

func (suite *GroupModuleTestSuite) newTestMsgProposal(tryExec group.Exec) group.MsgSubmitProposal {
	proposalJson := `{
		"group_policy_address": "1",
		"proposers": [
			"1"
		],
		"metadata": "",
		"messages": [
			{
				"@type": "/cosmos.group.v1.MsgUpdateGroupMetadata",
				"admin": "1",
				"group_id": 1,
				"metadata": ""
			}
		]
	}`
	var proposal group.MsgSubmitProposal
	err := json.Unmarshal([]byte(proposalJson), &proposal)
	suite.Require().NoError(err)

	proposal.Exec = tryExec
	return proposal
}

func (*GroupModuleTestSuite) newTestMsgVote(voteOption group.VoteOption, tryExec group.Exec) group.MsgVote {
	return group.MsgVote{
		ProposalId: 1,
		Voter:      "1",
		Option:     voteOption,
		Metadata:   "1",
		Exec:       tryExec,
	}
}

func (suite *GroupModuleTestSuite) insertTestData(threshold uint64, timestamp time.Time) {
	_, err := suite.module.db.Sql.Exec(
		`INSERT INTO block (height, hash, timestamp) VALUES (1, '1', $1)`,
		timestamp,
	)
	suite.Require().NoError(err)

	_, err = suite.module.db.Sql.Exec(
		`INSERT INTO transaction (hash, height, success, signatures)
		 VALUES ('1', 1, true, '{"1"}')`,
	)
	suite.Require().NoError(err)

	_, err = suite.module.db.Sql.Exec(
		`INSERT INTO group_with_policy VALUES (1, '1', '', '', $1, 1, 0)`,
		threshold,
	)
	suite.Require().NoError(err)

	_, err = suite.module.db.Sql.Exec(
		`INSERT INTO group_member VALUES (1, '1', '1', '1')`,
	)
	suite.Require().NoError(err)
}

func (suite *GroupModuleTestSuite) insertTestProposal(status group.ProposalStatus) {
	_, err := suite.module.db.Sql.Exec(
		`INSERT INTO group_proposal
		VALUES (1, 1, '1', '1', $1, 'PROPOSAL_EXECUTOR_RESULT_NOT_RUN', '1', '1', null)`,
		status.String(),
	)
	suite.Require().NoError(err)
}

func (suite *GroupModuleTestSuite) assertVotesCount(expectedCount int) {
	var votesCount int
	err := suite.module.db.Sqlx.QueryRow(
		`SELECT COUNT(*) from group_proposal_vote`,
	).Scan(&votesCount)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedCount, votesCount)
}
