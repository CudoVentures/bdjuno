package group

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	abcitypes "github.com/tendermint/tendermint/abci/types"

	juno "github.com/forbole/juno/v2/types"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/forbole/bdjuno/v2/database"
	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	testutils "github.com/forbole/bdjuno/v2/utils"
)

type GroupModuleTestSuite struct {
	suite.Suite
	module *Module
	db     *database.Db
}

func TestGroupModuleTestSuite(t *testing.T) {
	suite.Run(t, new(GroupModuleTestSuite))
}

func (suite *GroupModuleTestSuite) SetupTest() {
	db, cdc := testutils.NewTestDb(&suite.Suite, "groupTest")
	suite.module = NewModule(cdc, db)
	suite.db = db
}

func (suite *GroupModuleTestSuite) TestGroup_MsgCreateGroupWithPolicy() {
	decisionPolicy, err := codectypes.NewAnyWithValue(group.NewThresholdDecisionPolicy("1", time.Hour, 0))
	suite.Require().NoError(err)
	tx := suite.newTestTx("1", "", 1, "1", 0, 0, false)
	msg := group.MsgCreateGroupWithPolicy{
		Admin:               "admin",
		Members:             []group.MemberRequest{{Address: "1", Weight: "1", Metadata: "1"}},
		GroupMetadata:       "1",
		GroupPolicyMetadata: "1",
		GroupPolicyAsAdmin:  true,
		DecisionPolicy:      decisionPolicy,
	}

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	var groupRows []dbtypes.GroupRow
	err = suite.db.Sqlx.Select(&groupRows, `SELECT * FROM group_with_policy where id = 1`)
	suite.Require().NoError(err)
	suite.Require().Len(groupRows, 1)
	suite.Require().Equal(dbtypes.GroupRow{
		ID:                 1,
		Address:            "1",
		GroupMetadata:      "1",
		PolicyMetadata:     "1",
		Threshold:          1,
		VotingPeriod:       uint64(time.Hour.Seconds()),
		MinExecutionPeriod: 0,
	}, groupRows[0])

	var memberRows []dbtypes.GroupMemberRow
	err = suite.db.Sqlx.Select(&memberRows, `SELECT address, group_id, weight, metadata FROM group_member where group_id = 1`)
	suite.Require().NoError(err)
	suite.Require().Len(memberRows, 1)
	suite.Require().Equal(dbtypes.GroupMemberRow{
		Address:  "1",
		GroupID:  1,
		Weight:   1,
		Metadata: "1",
	}, memberRows[0])
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgSubmitProposal() {
	suite.insertTestGroup(1)
	suite.insertTestBlockAndTx(1, time.Now())

	timestamp := time.Date(2022, time.January, 1, 1, 1, 1, 0, time.FixedZone("", 0))
	tx := suite.newTestTx("1", timestamp.Format(time.RFC3339), 0, "", 1, 0, false)
	msg := suite.newTestMsgProposal(0)

	err := suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	expectedMsg := "[{\"@type\": \"/cosmos.group.v1.MsgUpdateGroupMetadata\", \"admin\": \"1\", \"group_id\": 1, \"metadata\": \"2\"}, {\"@type\": \"/cosmos.group.v1.MsgUpdateGroupPolicyMetadata\", \"admin\": \"1\", \"group_id\": 1, \"metadata\": \"2\"}, {\"@type\": \"/cosmos.group.v1.MsgUpdateGroupMembers\", \"admin\": \"1\", \"group_id\": \"1\", \"member_updates\": [{\"weight\": \"0\", \"address\": \"1\", \"metadata\": \"2\"}, {\"weight\": \"2\", \"address\": \"2\", \"metadata\": \"2\"}]}, {\"@type\": \"/cosmos.group.v1.MsgUpdateGroupPolicyDecisionPolicy\", \"admin\": \"1\", \"group_id\": 1, \"decision_policy\": {\"@type\": \"/cosmos.group.v1.ThresholdDecisionPolicy\", \"windows\": {\"voting_period\": \"2\", \"min_execution_period\": \"2\"}, \"threshold\": \"2\"}}]"
	var proposalRows []dbtypes.GroupProposalRow
	err = suite.db.Sqlx.Select(&proposalRows, `SELECT * FROM group_proposal where group_id = 1`)
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
		SubmitTime:       timestamp,
	}, proposalRows[0])

	suite.assertVotesCount(0)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgSubmitProposal_TryExec() {
	timestamp := time.Now()
	suite.insertTestGroup(2)
	suite.insertTestBlockAndTx(1, timestamp)

	tx := suite.newTestTx("1", timestamp.Format(time.RFC3339), 1, "", 1, 0, true)
	msg := suite.newTestMsgProposal(group.Exec_EXEC_TRY)

	err := suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	suite.assertVotesCount(1)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgVote_RequireEventEmit() {
	timestamp := time.Now()
	suite.insertTestGroup(1)
	suite.insertTestBlockAndTx(1, timestamp)
	suite.insertTestProposal(group.PROPOSAL_STATUS_SUBMITTED)

	tx := suite.newTestTx("1", timestamp.Format(time.RFC3339), 0, "", 0, 0, false)
	msg := suite.newTestMsgVote(0, 0)

	err := suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NotNil(err)
	suite.Require().Equal("error while getting EventVote", err.Error())

	suite.assertVotesCount(0)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgVote_TryExec_UpdateStatusToAccepted() {
	timestamp := time.Date(2022, time.January, 1, 1, 1, 1, 0, time.FixedZone("", 0))
	suite.insertTestGroup(1)
	suite.insertTestBlockAndTx(1, timestamp)
	suite.insertTestProposal(group.PROPOSAL_STATUS_SUBMITTED)

	tx := suite.newTestTx("1", timestamp.Format(time.RFC3339), 0, "", 0, group.PROPOSAL_EXECUTOR_RESULT_SUCCESS, true)
	msg := suite.newTestMsgVote(group.VOTE_OPTION_YES, group.Exec_EXEC_TRY)

	err := suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	var voteRows []dbtypes.GroupProposalVoteRow
	err = suite.db.Sqlx.Select(&voteRows, `SELECT * FROM group_proposal_vote where proposal_id = 1`)
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
	err = suite.db.Sqlx.QueryRow(
		`SELECT status, executor_result from group_proposal WHERE id = 1`,
	).Scan(&proposalStatus, &executorResult)
	suite.Require().NoError(err)
	suite.Require().Equal(group.PROPOSAL_STATUS_ACCEPTED.String(), proposalStatus)
	suite.Require().Equal(group.PROPOSAL_EXECUTOR_RESULT_SUCCESS.String(), executorResult)
}

func (suite *GroupModuleTestSuite) TestGroup_HangleMsgVote_UpdateStatusToRejected() {
	suite.insertTestGroup(1)
	suite.insertTestBlockAndTx(1, time.Now())
	suite.insertTestProposal(group.PROPOSAL_STATUS_SUBMITTED)

	tx := suite.newTestTx("1", time.Now().Format(time.RFC3339), 0, "", 0, 0, true)
	msg := suite.newTestMsgVote(group.VOTE_OPTION_NO, 0)

	err := suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	var voteOption string
	err = suite.db.Sqlx.QueryRow(
		`SELECT vote_option from group_proposal_vote WHERE proposal_id = 1`,
	).Scan(&voteOption)
	suite.Require().NoError(err)
	suite.Require().Equal(group.VOTE_OPTION_NO.String(), voteOption)

	var proposalStatus string
	err = suite.db.Sqlx.QueryRow(
		`SELECT status from group_proposal WHERE id = 1`,
	).Scan(&proposalStatus)
	suite.Require().NoError(err)
	suite.Require().Equal(group.PROPOSAL_STATUS_REJECTED.String(), proposalStatus)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgExec_RequireEvent() {
	timestamp := time.Now()
	suite.insertTestGroup(1)
	suite.insertTestBlockAndTx(1, timestamp)
	suite.insertTestProposal(group.PROPOSAL_STATUS_ABORTED)
	suite.insertTestBlockAndTx(2, timestamp.Add(time.Hour+time.Second))

	tx := suite.newTestTx("1", "", 0, "", 0, 0, false)
	msg := group.MsgExec{ProposalId: 1, Executor: "1"}

	err := suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NotNil(err)
	suite.Require().Equal("error while getting EventExec", err.Error())
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgExec_HandleMsgUpdateGroup() {
	timestamp := time.Now()
	suite.insertTestGroup(1)
	suite.insertTestBlockAndTx(1, timestamp)
	suite.insertTestBlockAndTx(2, timestamp.Add(time.Hour))
	tx := suite.newTestTx("1", timestamp.Format(time.RFC3339), 0, "", 1, group.PROPOSAL_EXECUTOR_RESULT_SUCCESS, true)
	msg := suite.newTestMsgProposal(group.Exec_EXEC_TRY)

	err := suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	var txHash sql.NullString
	var executorResult string
	err = suite.db.Sqlx.QueryRow(
		`SELECT transaction_hash, executor_result from group_proposal WHERE id = 1`,
	).Scan(&txHash, &executorResult)
	suite.Require().NoError(err)
	suite.Require().Equal(dbtypes.ToNullString("1"), txHash)
	suite.Require().Equal(group.PROPOSAL_EXECUTOR_RESULT_SUCCESS.String(), executorResult)

	var groupRows []dbtypes.GroupRow
	err = suite.db.Sqlx.Select(&groupRows, `SELECT * FROM group_with_policy where id = 1`)
	suite.Require().NoError(err)
	suite.Require().Len(groupRows, 1)
	suite.Require().Equal(dbtypes.GroupRow{
		ID:                 1,
		Address:            "1",
		GroupMetadata:      "2",
		PolicyMetadata:     "2",
		Threshold:          2,
		VotingPeriod:       2,
		MinExecutionPeriod: 2,
	}, groupRows[0])

	var memberRows []dbtypes.GroupMemberRow
	err = suite.db.Sqlx.Select(&memberRows, `SELECT address, group_id, weight, metadata FROM group_member where group_id = 1 AND removed = B'0'`)
	suite.Require().NoError(err)
	suite.Require().Len(memberRows, 1)
	suite.Require().Equal(dbtypes.GroupMemberRow{
		Address:  "2",
		GroupID:  1,
		Weight:   2,
		Metadata: "2",
	}, memberRows[0])
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgWithdrawProposal() {
	// todo
}

func (suite *GroupModuleTestSuite) newTestTx(
	txHash string, timestamp string, groupID uint64, groupAddress string,
	proposalID uint64, executorResult group.ProposalExecutorResult, voteEvent bool,
) *juno.Tx {
	events := make([]abcitypes.Event, 0)

	if groupID != 0 {
		eventCreateGroup, err := sdk.TypedEventToEvent(&group.EventCreateGroup{GroupId: groupID})
		suite.Require().NoError(err)
		events = append(events, abcitypes.Event(eventCreateGroup))
	}

	if groupAddress != "" {
		eventCreateGroupPolicy, err := sdk.TypedEventToEvent(&group.EventCreateGroupPolicy{Address: groupAddress})
		suite.Require().NoError(err)
		events = append(events, abcitypes.Event(eventCreateGroupPolicy))
	}

	if proposalID != 0 {
		eventSubmitProposal, err := sdk.TypedEventToEvent(&group.EventSubmitProposal{ProposalId: proposalID})
		suite.Require().NoError(err)
		events = append(events, abcitypes.Event(eventSubmitProposal))
	}

	if executorResult != group.PROPOSAL_EXECUTOR_RESULT_UNSPECIFIED {
		eventExec, err := sdk.TypedEventToEvent(&group.EventExec{Result: executorResult})
		suite.Require().NoError(err)

		events = append(events, abcitypes.Event(eventExec))
	}

	if voteEvent {
		eventVote, err := sdk.TypedEventToEvent(&group.EventVote{ProposalId: 1})
		suite.Require().NoError(err)
		events = append(events, abcitypes.Event(eventVote))
	}

	txLog := sdk.ABCIMessageLogs{{MsgIndex: 0, Events: sdk.StringifyEvents(events)}}
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
				"metadata": "2"
			},
			{
				"@type": "/cosmos.group.v1.MsgUpdateGroupPolicyMetadata",
				"admin": "1",
				"group_id": 1,
				"metadata": "2"
			},
			{
				"@type": "/cosmos.group.v1.MsgUpdateGroupMembers",
				"admin": "1",
				"group_id": "1",
				"member_updates": [
					{ "weight": "0", "address": "1", "metadata": "2" },
					{ "weight": "2", "address": "2", "metadata": "2" }
				]
			},
			{
				"@type": "/cosmos.group.v1.MsgUpdateGroupPolicyDecisionPolicy",
				"admin": "1",
				"group_id": 1,
				"decision_policy": {
					"@type":"/cosmos.group.v1.ThresholdDecisionPolicy",
					"threshold":"2",
					"windows": {"voting_period": "2", "min_execution_period": "2"}
				}
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

func (suite *GroupModuleTestSuite) insertTestGroup(threshold uint64) {
	_, err := suite.db.Sql.Exec(`INSERT INTO group_with_policy VALUES (1, '1', '', '', $1, 3600, 0)`, threshold)
	suite.Require().NoError(err)

	_, err = suite.db.Sql.Exec(`INSERT INTO group_member VALUES (1, '1', '1', '1')`)
	suite.Require().NoError(err)
}

func (suite *GroupModuleTestSuite) insertTestBlockAndTx(height int, timestamp time.Time) {
	_, err := suite.db.Sql.Exec(
		`INSERT INTO block (height, hash, timestamp) VALUES ($1, $2, $3)`,
		height, strconv.Itoa(height), timestamp,
	)
	suite.Require().NoError(err)

	_, err = suite.db.Sql.Exec(
		`INSERT INTO transaction (hash, height, success, signatures) VALUES ($1, $2, true, '{"1"}')`,
		strconv.Itoa(height), height,
	)
	suite.Require().NoError(err)
}

func (suite *GroupModuleTestSuite) insertTestProposal(status group.ProposalStatus) {
	_, err := suite.db.Sql.Exec(
		`INSERT INTO group_proposal VALUES (1, 1, '1', '1', $1, 'PROPOSAL_EXECUTOR_RESULT_NOT_RUN', '1', '1', NOW(), null)`,
		status.String(),
	)
	suite.Require().NoError(err)
}

func (suite *GroupModuleTestSuite) assertVotesCount(expectedCount int) {
	var votesCount int
	err := suite.db.Sqlx.QueryRow(`SELECT COUNT(*) from group_proposal_vote`).Scan(&votesCount)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedCount, votesCount)
}
