package group

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-co-op/gocron"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/forbole/bdjuno/v2/database"
	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/utils"
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
	db := utils.NewTestDb(&suite.Suite, "groupTest")
	suite.module = NewModule(simapp.MakeTestEncodingConfig().Marshaler, db)
	suite.db = db

	_, err := db.Sql.Exec(`INSERT INTO group_with_policy VALUES (1, '1', '', '', 1, 1, 0)`)
	suite.Require().NoError(err)

	_, err = db.Sql.Exec(`INSERT INTO group_member VALUES (1, '1', '1', '1')`)
	suite.Require().NoError(err)

	_, err = db.Sql.Exec(`INSERT INTO block (height, hash, timestamp) VALUES (1, '1', NOW())`)
	suite.Require().NoError(err)

	_, err = db.Sql.Exec(`INSERT INTO transaction (hash, height, success, signatures) VALUES ('1', 1, true, ARRAY['1'])`)
	suite.Require().NoError(err)

	_, err = db.Sql.Exec(`INSERT INTO group_proposal VALUES (1, 1, '1', '1', 'PROPOSAL_STATUS_SUBMITTED', 'PROPOSAL_EXECUTOR_RESULT_NOT_RUN', null, null, null, '1', '1', $1, null)`, testTimestamp())
	suite.Require().NoError(err)

}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgCreateGroupWithPolicy() {
	groupID := uint64(2)
	tx, err := utils.NewTestTx(testTimestamp()).WithEventCreateGroup(groupID, "1").Build()
	suite.Require().NoError(err)

	decisionPolicy, err := codectypes.NewAnyWithValue(group.NewThresholdDecisionPolicy("1", time.Hour, 0))
	suite.Require().NoError(err)

	msg := group.MsgCreateGroupWithPolicy{
		Admin:               "1",
		Members:             []group.MemberRequest{{Address: "1", Weight: "1", Metadata: "1"}},
		GroupMetadata:       "1",
		GroupPolicyMetadata: "1",
		DecisionPolicy:      decisionPolicy,
	}

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	expectedGroup := []dbtypes.GroupRow{{
		ID:                 groupID,
		Address:            "1",
		GroupMetadata:      "1",
		PolicyMetadata:     "1",
		Threshold:          1,
		VotingPeriod:       uint64(time.Hour.Seconds()),
		MinExecutionPeriod: 0,
	}}
	var actualGroup []dbtypes.GroupRow
	err = suite.db.Sqlx.Select(&actualGroup, `SELECT * FROM group_with_policy where id = $1`, groupID)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedGroup, actualGroup)

	expectedMember := []dbtypes.GroupMemberRow{{
		Address:  "1",
		GroupID:  groupID,
		Weight:   1,
		Metadata: "1",
	}}
	var actualMember []dbtypes.GroupMemberRow
	err = suite.db.Sqlx.Select(&actualMember, `SELECT address, group_id, weight, metadata FROM group_member where group_id = $1`, groupID)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedMember, actualMember)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgSubmitProposal() {
	proposalID := uint64(2)
	tx, err := utils.NewTestTx(testTimestamp()).WithEventSubmitProposal(proposalID).Build()
	suite.Require().NoError(err)

	msg, err := group.NewMsgSubmitProposal("1", []string{"1"}, []types.Msg{}, "1", 0)
	suite.Require().NoError(err)

	err = suite.module.HandleMsg(0, msg, tx)
	suite.Require().NoError(err)

	expectedProposal := []dbtypes.GroupProposalRow{{
		ID:               proposalID,
		GroupID:          1,
		ProposalMetadata: "1",
		Proposer:         "1",
		Status:           group.PROPOSAL_STATUS_SUBMITTED.String(),
		ExecutorResult:   group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN.String(),
		Messages:         "[]",
		BlockHeight:      1,
		SubmitTime:       testTimestamp(),
	}}
	var actualProposal []dbtypes.GroupProposalRow
	err = suite.db.Sqlx.Select(&actualProposal, `SELECT * FROM group_proposal where id = $1`, proposalID)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedProposal, actualProposal)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgSubmitProposal_TryExec() {
	tx, err := utils.NewTestTx(testTimestamp()).WithEventSubmitProposal(1).WithEventExec(group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN).Build()
	suite.Require().NoError(err)

	msg, err := group.NewMsgSubmitProposal("1", []string{"1"}, []types.Msg{}, "1", group.Exec_EXEC_TRY)
	suite.Require().NoError(err)

	err = suite.module.HandleMsg(0, msg, tx)
	suite.Require().Equal("error while getting EventVote", err.Error())
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgVote_UpdateStatusToAccepted() {
	tx, err := utils.NewTestTx(testTimestamp()).WithEventVote().WithEventExec(group.PROPOSAL_EXECUTOR_RESULT_SUCCESS).Build()
	suite.Require().NoError(err)

	proposalID := uint64(1)
	msg := group.MsgVote{ProposalId: proposalID, Voter: "1", Option: group.VOTE_OPTION_YES}

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	expectedVote := []dbtypes.GroupProposalVoteRow{{
		ProposalID: proposalID,
		GroupID:    1,
		Voter:      "1",
		VoteOption: group.VOTE_OPTION_YES.String(),
		SubmitTime: testTimestamp(),
	}}
	var actualVote []dbtypes.GroupProposalVoteRow
	err = suite.db.Sqlx.Select(&actualVote, `SELECT * FROM group_proposal_vote where proposal_id = $1`, proposalID)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedVote, actualVote)
}

func (suite *GroupModuleTestSuite) TestGroup_HangleMsgVote_UpdateStatusToRejected() {
	tx, err := utils.NewTestTx(testTimestamp()).WithEventVote().Build()
	suite.Require().NoError(err)

	proposalID := uint64(1)
	msg := group.MsgVote{ProposalId: proposalID, Voter: "1", Option: group.VOTE_OPTION_NO}

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	expectedVote := group.VOTE_OPTION_NO.String()
	var actualVote string
	err = suite.db.Sqlx.QueryRow(`SELECT vote_option from group_proposal_vote WHERE proposal_id = $1`, proposalID).Scan(&actualVote)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedVote, actualVote)

	expectedStatus := group.PROPOSAL_STATUS_REJECTED.String()
	var actualStatus string
	err = suite.db.Sqlx.QueryRow(`SELECT status from group_proposal WHERE id = $1`, proposalID).Scan(&actualStatus)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedStatus, actualStatus)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgVote_TryExec() {
	tx, err := utils.NewTestTx(testTimestamp()).WithEventExec(group.PROPOSAL_EXECUTOR_RESULT_SUCCESS).WithEventVote().Build()
	suite.Require().NoError(err)

	proposalID := uint64(1)
	msg := group.MsgVote{ProposalId: proposalID, Voter: "1", Option: group.VOTE_OPTION_YES, Exec: group.Exec_EXEC_TRY}

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	expectedStatus := group.PROPOSAL_STATUS_ACCEPTED.String()
	expectedResult := group.PROPOSAL_EXECUTOR_RESULT_SUCCESS.String()
	var actualStatus string
	var actualResult string
	err = suite.db.Sqlx.QueryRow(`SELECT status, executor_result from group_proposal WHERE id = $1`, proposalID).Scan(&actualStatus, &actualResult)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedStatus, actualStatus)
	suite.Require().Equal(expectedResult, actualResult)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgExec() {
	tx, err := utils.NewTestTx(testTimestamp()).WithEventExec(group.PROPOSAL_EXECUTOR_RESULT_SUCCESS).Build()
	suite.Require().NoError(err)

	proposalID := uint64(1)
	msg := group.MsgExec{ProposalId: proposalID, Executor: "1"}

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	expectedProposal := []dbtypes.GroupProposalRow{{
		ID:               proposalID,
		GroupID:          1,
		ProposalMetadata: "1",
		Proposer:         "1",
		Status:           group.PROPOSAL_STATUS_SUBMITTED.String(),
		ExecutorResult:   group.PROPOSAL_EXECUTOR_RESULT_SUCCESS.String(),
		Executor:         dbtypes.ToNullString("1"),
		ExecutionTime:    sql.NullTime{Time: testTimestamp(), Valid: true},
		ExecutionLog:     dbtypes.ToNullString("1"),
		Messages:         "1",
		TxHash:           dbtypes.ToNullString("1"),
		BlockHeight:      1,
		SubmitTime:       testTimestamp(),
	}}
	var actualProposal []dbtypes.GroupProposalRow
	err = suite.db.Sqlx.Select(&actualProposal, `SELECT * FROM group_proposal WHERE ID = $1`, proposalID)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedProposal, actualProposal)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgExec_HandleMsgUpdateGroup() {
	proposalID := uint64(2)
	tx, err := utils.NewTestTx(testTimestamp()).WithEventSubmitProposal(proposalID).WithEventExec(group.PROPOSAL_EXECUTOR_RESULT_SUCCESS).WithEventVote().Build()
	suite.Require().NoError(err)

	proposalJson := `{"group_policy_address": "1","proposers": ["1"],"metadata": "","messages": [{"@type": "/cosmos.group.v1.MsgUpdateGroupMetadata","admin": "1","group_id": 1,"metadata": "2"},{"@type": "/cosmos.group.v1.MsgUpdateGroupPolicyMetadata","admin": "1","group_id": 1,"metadata": "2"},{"@type": "/cosmos.group.v1.MsgUpdateGroupMembers","admin": "1","group_id": "1","member_updates": [{ "weight": "0", "address": "1", "metadata": "2" },{ "weight": "2", "address": "2", "metadata": "2" }]},{"@type": "/cosmos.group.v1.MsgUpdateGroupPolicyDecisionPolicy","admin": "1","group_id": 1,"decision_policy": {"@type":"/cosmos.group.v1.ThresholdDecisionPolicy","threshold":"2","windows": {"voting_period": "2", "min_execution_period": "2"}}}]}`
	msg := group.MsgSubmitProposal{Exec: group.Exec_EXEC_TRY}
	err = json.Unmarshal([]byte(proposalJson), &msg)
	suite.Require().NoError(err)

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	expectedGroup := []dbtypes.GroupRow{{
		ID:                 1,
		Address:            "1",
		GroupMetadata:      "2",
		PolicyMetadata:     "2",
		Threshold:          2,
		VotingPeriod:       2,
		MinExecutionPeriod: 2,
	}}
	var actualGroup []dbtypes.GroupRow
	err = suite.db.Sqlx.Select(&actualGroup, `SELECT * FROM group_with_policy where id = 1`)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedGroup, actualGroup)

	expectedMember := []dbtypes.GroupMemberRow{{
		Address:  "2",
		GroupID:  1,
		Weight:   2,
		Metadata: "2",
	}}
	var actualMember []dbtypes.GroupMemberRow
	err = suite.db.Sqlx.Select(&actualMember, `SELECT address, group_id, weight, metadata FROM group_member WHERE weight > 0 AND group_id = 1`)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedMember, actualMember)

	expectedStatuses := []string{group.PROPOSAL_STATUS_ACCEPTED.String(), group.PROPOSAL_STATUS_ABORTED.String()}
	var actualStatuses []string
	err = suite.db.Sqlx.Select(&actualStatuses, `SELECT status FROM group_proposal`)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedStatuses, actualStatuses)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgWithdrawProposal() {
	tx, err := utils.NewTestTx(testTimestamp()).WithEventWithdrawProposal().Build()
	suite.Require().NoError(err)

	proposalID := uint64(1)
	msg := group.MsgWithdrawProposal{ProposalId: proposalID, Address: "1"}

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	expectedStatus := group.PROPOSAL_STATUS_WITHDRAWN.String()
	var actualStatus string
	err = suite.db.Sqlx.QueryRow(`SELECT status FROM group_proposal WHERE id = $1`, proposalID).Scan(&actualStatus)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedStatus, actualStatus)
}

func (suite *GroupModuleTestSuite) TestGroup_HandlePeriodicOperations() {
	proposalID := uint64(2)
	_, err := suite.db.Sql.Exec(
		`INSERT INTO group_proposal VALUES ($1, 1, '1', '1', 'PROPOSAL_STATUS_SUBMITTED', 'PROPOSAL_EXECUTOR_RESULT_NOT_RUN', null, null, null, '1', '1', $2, null)`,
		proposalID, testTimestamp().Add(time.Second),
	)
	suite.Require().NoError(err)

	_, err = suite.db.Sql.Exec(`INSERT INTO block (height, hash, timestamp) VALUES ($1, $2, $3)`, "2", "2", testTimestamp().Add(time.Second*2))
	suite.Require().NoError(err)

	scheduler := gocron.NewScheduler(time.UTC)
	err = suite.module.RegisterPeriodicOperations(scheduler)
	suite.Require().NoError(err)
	scheduler.StartAsync()
	time.Sleep(time.Second * 5)

	expectedStatuses := []string{group.PROPOSAL_STATUS_SUBMITTED.String(), group.PROPOSAL_STATUS_REJECTED.String()}
	var actualStatuses []string
	err = suite.db.Sqlx.Select(&actualStatuses, `SELECT status FROM group_proposal`)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedStatuses, actualStatuses)
}

func testTimestamp() time.Time {
	return time.Date(2022, time.January, 1, 1, 1, 1, 0, time.FixedZone("", 0))
}
