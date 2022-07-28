package group

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/lib/pq"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/forbole/bdjuno/v2/database"
	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/utils"
)

var (
	timestamp      = time.Date(2022, time.January, 1, 1, 1, 1, 0, time.FixedZone("", 0))
	statusDefault  = group.PROPOSAL_STATUS_SUBMITTED
	statusAccepted = group.PROPOSAL_STATUS_ACCEPTED
	statusRejected = group.PROPOSAL_STATUS_REJECTED
	resultDefault  = group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN
	resultSuccess  = group.PROPOSAL_EXECUTOR_RESULT_SUCCESS
	execTry        = group.Exec_EXEC_TRY
	voteYes        = group.VOTE_OPTION_YES
	voteNo         = group.VOTE_OPTION_NO
	one            = uint64(1)
	oneStr         = "1"
	two            = uint64(2)
	twoStr         = "2"
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
	db, err := utils.NewTestDb("groupTest")
	suite.Require().NoError(err)
	suite.module = NewModule(simapp.MakeTestEncodingConfig().Marshaler, db)
	suite.db = db

	_, err = db.Sql.Exec(`INSERT INTO group_with_policy VALUES ($1, $2, $3, $4, $5, $6, 0)`, one, oneStr, oneStr, oneStr, one, one)
	suite.Require().NoError(err)

	_, err = db.Sql.Exec(`INSERT INTO group_member VALUES ($1, $2, $3, $4, $5)`, one, oneStr, oneStr, oneStr, timestamp)
	suite.Require().NoError(err)

	_, err = db.Sql.Exec(`INSERT INTO block (height, hash, timestamp) VALUES ($1, $2, $3)`, one, oneStr, timestamp)
	suite.Require().NoError(err)

	_, err = db.Sql.Exec(`INSERT INTO transaction (hash, height, success, signatures) VALUES ($1, $2, true, $3)`, oneStr, one, pq.Array([]string{oneStr}))
	suite.Require().NoError(err)

	_, err = db.Sql.Exec(`INSERT INTO group_proposal VALUES ($1, $2, $3, $4, $5, $6, null, null, null, $7, $8, $9, null)`, one, one, oneStr, oneStr, statusDefault.String(), resultDefault.String(), oneStr, oneStr, timestamp)
	suite.Require().NoError(err)

}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgCreateGroupWithPolicy() {
	tx, err := utils.NewTestTx(timestamp, oneStr, one).WithEventCreateGroup(two, twoStr).Build()
	suite.Require().NoError(err)

	decisionPolicy, err := codectypes.NewAnyWithValue(group.NewThresholdDecisionPolicy(twoStr, time.Second*time.Duration(two), 0))
	suite.Require().NoError(err)

	msg := group.MsgCreateGroupWithPolicy{
		Admin:               twoStr,
		Members:             []group.MemberRequest{{Address: twoStr, Weight: twoStr, Metadata: twoStr}},
		GroupMetadata:       twoStr,
		GroupPolicyMetadata: twoStr,
		DecisionPolicy:      decisionPolicy,
	}

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	expectedGroup := []dbtypes.GroupRow{{
		ID:                 two,
		Address:            twoStr,
		GroupMetadata:      twoStr,
		PolicyMetadata:     twoStr,
		Threshold:          two,
		VotingPeriod:       two,
		MinExecutionPeriod: 0,
	}}
	var actualGroup []dbtypes.GroupRow
	err = suite.db.Sqlx.Select(&actualGroup, `SELECT * FROM group_with_policy where id = $1`, two)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedGroup, actualGroup)

	expectedMember := []dbtypes.GroupMemberRow{{
		Address:  twoStr,
		GroupID:  two,
		Weight:   two,
		Metadata: twoStr,
		AddTime:  timestamp,
	}}
	var actualMember []dbtypes.GroupMemberRow
	err = suite.db.Sqlx.Select(&actualMember, `SELECT * FROM group_member where group_id = $1`, two)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedMember, actualMember)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgSubmitProposal() {
	tx, err := utils.NewTestTx(timestamp, oneStr, one).WithEventSubmitProposal(two).Build()
	suite.Require().NoError(err)

	msg, err := group.NewMsgSubmitProposal(oneStr, []string{twoStr}, []types.Msg{}, twoStr, 0)
	suite.Require().NoError(err)

	err = suite.module.HandleMsg(0, msg, tx)
	suite.Require().NoError(err)

	expectedProposal := []dbtypes.GroupProposalRow{{
		ID:               two,
		GroupID:          one,
		ProposalMetadata: twoStr,
		Proposer:         twoStr,
		Status:           statusDefault.String(),
		ExecutorResult:   resultDefault.String(),
		Messages:         "[]",
		BlockHeight:      int64(one),
		SubmitTime:       timestamp,
	}}
	var actualProposal []dbtypes.GroupProposalRow
	err = suite.db.Sqlx.Select(&actualProposal, `SELECT * FROM group_proposal where id = $1`, two)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedProposal, actualProposal)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgSubmitProposal_TryExec() {
	tx, err := utils.NewTestTx(timestamp, oneStr, one).WithEventSubmitProposal(two).Build()
	suite.Require().NoError(err)

	msg, err := group.NewMsgSubmitProposal(oneStr, []string{twoStr}, []types.Msg{}, twoStr, execTry)
	suite.Require().NoError(err)

	err = suite.module.HandleMsg(0, msg, tx)
	suite.Require().Equal("error while getting EventVote", err.Error())
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgVote_UpdateStatusToAccepted() {
	tx, err := utils.NewTestTx(timestamp, oneStr, one).WithEventVote().Build()
	suite.Require().NoError(err)

	msg := group.MsgVote{ProposalId: one, Voter: oneStr, Option: voteYes}

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	expectedVote := []dbtypes.GroupProposalVoteRow{{
		ProposalID: one,
		GroupID:    one,
		Voter:      oneStr,
		VoteOption: voteYes.String(),
		SubmitTime: timestamp,
	}}
	var actualVote []dbtypes.GroupProposalVoteRow
	err = suite.db.Sqlx.Select(&actualVote, `SELECT * FROM group_proposal_vote where proposal_id = $1`, one)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedVote, actualVote)

	expectedStatus := statusAccepted.String()
	var actualStatus string
	err = suite.db.Sqlx.QueryRow(`SELECT status from group_proposal WHERE id = $1`, one).Scan(&actualStatus)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedStatus, actualStatus)
}

func (suite *GroupModuleTestSuite) TestGroup_HangleMsgVote_UpdateStatusToRejected() {
	tx, err := utils.NewTestTx(timestamp, oneStr, one).WithEventVote().Build()
	suite.Require().NoError(err)

	msg := group.MsgVote{ProposalId: one, Voter: oneStr, Option: voteNo}

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	expectedVote := voteNo.String()
	var actualVote string
	err = suite.db.Sqlx.QueryRow(`SELECT vote_option from group_proposal_vote WHERE proposal_id = $1`, one).Scan(&actualVote)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedVote, actualVote)

	expectedStatus := statusRejected.String()
	var actualStatus string
	err = suite.db.Sqlx.QueryRow(`SELECT status from group_proposal WHERE id = $1`, one).Scan(&actualStatus)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedStatus, actualStatus)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgVote_TryExec() {
	tx, err := utils.NewTestTx(timestamp, oneStr, one).WithEventVote().Build()
	suite.Require().NoError(err)

	msg := group.MsgVote{ProposalId: one, Voter: oneStr, Option: voteYes, Exec: execTry}

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().Equal("error while getting EventExec", err.Error())
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgExec() {
	tx, err := utils.NewTestTx(timestamp, oneStr, one).WithEventExec(resultSuccess).Build()
	suite.Require().NoError(err)

	msg := group.MsgExec{ProposalId: one, Executor: oneStr}

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	expectedProposal := []dbtypes.GroupProposalRow{{
		ID:               one,
		GroupID:          one,
		ProposalMetadata: oneStr,
		Proposer:         oneStr,
		Status:           statusDefault.String(),
		ExecutorResult:   resultSuccess.String(),
		Executor:         dbtypes.ToNullString(oneStr),
		ExecutionTime:    sql.NullTime{Time: timestamp, Valid: true},
		ExecutionLog:     dbtypes.ToNullString(oneStr),
		Messages:         oneStr,
		TxHash:           dbtypes.ToNullString(oneStr),
		BlockHeight:      int64(one),
		SubmitTime:       timestamp,
	}}
	var actualProposal []dbtypes.GroupProposalRow
	err = suite.db.Sqlx.Select(&actualProposal, `SELECT * FROM group_proposal WHERE ID = $1`, one)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedProposal, actualProposal)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgExec_HandleMsgUpdateGroup() {
	tx, err := utils.NewTestTx(timestamp, oneStr, one).WithEventSubmitProposal(two).WithEventVote().WithEventExec(resultSuccess).Build()
	suite.Require().NoError(err)

	msgJson := fmt.Sprintf(`{"group_policy_address": "%[1]d","proposers": ["%[1]d"],"metadata": "","messages": [{"@type": "/cosmos.group.v1.MsgUpdateGroupMetadata","admin": "%[1]d","group_id": %[1]d,"metadata": "%[2]d"},{"@type": "/cosmos.group.v1.MsgUpdateGroupPolicyMetadata","admin": "%[1]d","group_id": %[1]d,"metadata": "%[2]d"},{"@type": "/cosmos.group.v1.MsgUpdateGroupMembers","admin": "%[1]d","group_id": "%[1]d","member_updates": [{ "weight": "0", "address": "%[1]d", "metadata": "%[2]d" },{ "weight": "%[2]d", "address": "%[2]d", "metadata": "%[2]d" }]},{"@type": "/cosmos.group.v1.MsgUpdateGroupPolicyDecisionPolicy","admin": "%[1]d","group_id": %[1]d,"decision_policy": {"@type":"/cosmos.group.v1.ThresholdDecisionPolicy","threshold":"%[2]d","windows": {"voting_period": "%[2]d", "min_execution_period": "%[2]d"}}}]}`, one, two)
	msg := group.MsgSubmitProposal{Exec: execTry}
	err = json.Unmarshal([]byte(msgJson), &msg)
	suite.Require().NoError(err)

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	expectedGroup := []dbtypes.GroupRow{{
		ID:                 one,
		Address:            oneStr,
		GroupMetadata:      twoStr,
		PolicyMetadata:     twoStr,
		Threshold:          two,
		VotingPeriod:       two,
		MinExecutionPeriod: two,
	}}
	var actualGroup []dbtypes.GroupRow
	err = suite.db.Sqlx.Select(&actualGroup, `SELECT * FROM group_with_policy where id = $1`, one)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedGroup, actualGroup)

	expectedMember := []dbtypes.GroupMemberRow{{
		Address:  twoStr,
		GroupID:  one,
		Weight:   two,
		Metadata: twoStr,
		AddTime:  timestamp,
	}}
	var actualMember []dbtypes.GroupMemberRow
	err = suite.db.Sqlx.Select(&actualMember, `SELECT * FROM group_member WHERE weight > 0 AND group_id = $1`, one)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedMember, actualMember)

	expectedStatuses := []string{statusAccepted.String(), group.PROPOSAL_STATUS_ABORTED.String()}
	var actualStatuses []string
	err = suite.db.Sqlx.Select(&actualStatuses, `SELECT status FROM group_proposal`)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedStatuses, actualStatuses)
}

func (suite *GroupModuleTestSuite) TestGroup_HandleMsgWithdrawProposal() {
	tx, err := utils.NewTestTx(timestamp, oneStr, one).WithEventWithdrawProposal().Build()
	suite.Require().NoError(err)

	msg := group.MsgWithdrawProposal{ProposalId: one, Address: oneStr}

	err = suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	expectedStatus := group.PROPOSAL_STATUS_WITHDRAWN.String()
	var actualStatus string
	err = suite.db.Sqlx.QueryRow(`SELECT status FROM group_proposal WHERE id = $1`, one).Scan(&actualStatus)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedStatus, actualStatus)
}

func (suite *GroupModuleTestSuite) TestGroup_HandlePeriodicOperations() {
	_, err := suite.db.Sql.Exec(`INSERT INTO group_proposal VALUES ($1, $2, $3, $4, $5, $6, null, null, null, $7, $8, $9, null)`, two, one, oneStr, oneStr, statusDefault.String(), resultDefault.String(), oneStr, oneStr, timestamp.Add(time.Second))
	suite.Require().NoError(err)

	_, err = suite.db.Sql.Exec(`INSERT INTO block (height, hash, timestamp) VALUES ($1, $2, $3)`, two, two, timestamp.Add(time.Second*2))
	suite.Require().NoError(err)

	scheduler := gocron.NewScheduler(time.UTC)
	err = suite.module.RegisterPeriodicOperations(scheduler)
	suite.Require().NoError(err)
	scheduler.StartAsync()
	time.Sleep(time.Second * 5)

	expectedStatuses := []string{statusDefault.String(), statusRejected.String()}
	var actualStatuses []string
	err = suite.db.Sqlx.Select(&actualStatuses, `SELECT status FROM group_proposal`)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedStatuses, actualStatuses)
}
