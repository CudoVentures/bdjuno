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
	"github.com/forbole/bdjuno/v2/types"
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

func (suite *GroupModuleTestSuite) TestGroup_CreateGroupWithPolicy() {
	decisionPolicy, _ := codectypes.NewAnyWithValue(
		group.NewThresholdDecisionPolicy("1", time.Hour, 0),
	)

	msg := group.MsgCreateGroupWithPolicy{
		Admin: "admin",
		Members: []group.MemberRequest{
			{Address: "cudos1", Weight: "1", Metadata: "1"},
		},
		GroupMetadata:       "1",
		GroupPolicyMetadata: "1",
		GroupPolicyAsAdmin:  true,
		DecisionPolicy:      decisionPolicy,
	}

	tx := suite.newTestTx("1", "", 1, "cudos1", 0, 0)

	err := suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	var groupRows []dbtypes.GroupWithPolicyRow
	err = suite.module.db.Sqlx.
		Select(&groupRows, "SELECT * FROM group_with_policy where id = 1")
	suite.Require().NoError(err)
	suite.Require().Len(groupRows, 1)
	suite.Require().Equal(dbtypes.GroupWithPolicyRow{
		ID:                 1,
		Address:            "cudos1",
		GroupMetadata:      "1",
		PolicyMetadata:     "1",
		Threshold:          1,
		VotingPeriod:       uint64(time.Hour.Seconds()),
		MinExecutionPeriod: 0,
	}, groupRows[0])

	var memberRows []dbtypes.GroupMemberRow
	err = suite.module.db.Sqlx.
		Select(&memberRows, "SELECT * FROM group_member where group_id = 1")
	suite.Require().NoError(err)
	suite.Require().Len(memberRows, 1)
	suite.Require().Equal(dbtypes.GroupMemberRow{
		Address:        "cudos1",
		GroupID:        1,
		Weight:         1,
		MemberMetadata: "1",
	}, memberRows[0])
}

func (suite *GroupModuleTestSuite) TestGroup_CreateGroupProposal() {
	suite.insertTestData(1)
	proposal := suite.newTestProposalMsg(0)
	tx := suite.newTestTx("1", "", 0, "", 1, 0)

	err := suite.module.HandleMsg(0, &proposal, tx)
	suite.Require().NoError(err)

	expectedMsg :=
		"[{\"@type\": \"/cosmos.group.v1.MsgUpdateGroupMetadata\", \"admin\": \"cudos1\", \"group_id\": 1, \"metadata\": \"\"}]"

	var proposalRows []dbtypes.GroupProposalRow
	err = suite.module.db.Sqlx.
		Select(&proposalRows, "SELECT * FROM group_proposal where group_id = 1")
	suite.Require().NoError(err)
	suite.Require().Len(proposalRows, 1)
	suite.Require().Equal(dbtypes.GroupProposalRow{
		ID:               1,
		GroupID:          1,
		ProposalMetadata: "",
		Proposer:         "cudos1",
		Status:           group.PROPOSAL_STATUS_SUBMITTED.String(),
		ExecutorResult:   group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN.String(),
		Messages:         expectedMsg,
		TxHash:           dbtypes.ToNullString(""),
		BlockHeight:      1,
	}, proposalRows[0])

	var votesCount int
	err = suite.module.db.Sqlx.
		QueryRow(`SELECT COUNT(*) from group_proposal_vote`).
		Scan(&votesCount)
	suite.Require().NoError(err)
	suite.Require().Equal(0, votesCount)
}

func (suite *GroupModuleTestSuite) TestGroup_CreateGroupProposal_TryExecEnoughThreshold() {
	suite.insertTestData(1)
	proposal := suite.newTestProposalMsg(group.Exec_EXEC_TRY)
	tx := suite.newTestTx("1", "", 0, "", 1, group.PROPOSAL_EXECUTOR_RESULT_SUCCESS)

	err := suite.module.HandleMsg(0, &proposal, tx)
	suite.Require().NoError(err)

	expectedMsg :=
		"[{\"@type\": \"/cosmos.group.v1.MsgUpdateGroupMetadata\", \"admin\": \"cudos1\", \"group_id\": 1, \"metadata\": \"\"}]"

	var proposalRows []dbtypes.GroupProposalRow
	err = suite.module.db.Sqlx.
		Select(&proposalRows, "SELECT * FROM group_proposal where group_id = 1")
	suite.Require().NoError(err)
	suite.Require().Len(proposalRows, 1)
	suite.Require().Equal(dbtypes.GroupProposalRow{
		ID:               1,
		GroupID:          1,
		ProposalMetadata: "",
		Proposer:         "cudos1",
		Status:           group.PROPOSAL_STATUS_ABORTED.String(),
		ExecutorResult:   group.PROPOSAL_EXECUTOR_RESULT_SUCCESS.String(),
		Messages:         expectedMsg,
		TxHash:           dbtypes.ToNullString("1"),
		BlockHeight:      1,
	}, proposalRows[0])

	var votesCount int
	err = suite.module.db.Sqlx.
		QueryRow(`SELECT COUNT(*) from group_proposal_vote`).
		Scan(&votesCount)
	suite.Require().NoError(err)
	suite.Require().Equal(1, votesCount)
}

func (suite *GroupModuleTestSuite) TestGroup_CreateGroupProposal_TryExecNotEnoughThreshold() {

}

func (suite *GroupModuleTestSuite) TestGroup_CreateGroupProposalVote() {

}

func (suite *GroupModuleTestSuite) TestGroup_CreateGroupProposalVote_TryExecEnoughThreshold() {

}

func (suite *GroupModuleTestSuite) TestGroup_CreateGroupProposalVote_TryExecNotEnoughThreshold() {

}

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

func (suite *GroupModuleTestSuite) insertTestData(threshold uint64) {
	_, err := suite.module.db.Sql.Exec(
		`INSERT INTO block (height, hash, timestamp) VALUES (1, '1', NOW())`,
	)
	suite.Require().NoError(err)

	_, err = suite.module.db.Sql.Exec(
		`INSERT INTO transaction (hash, height, success, signatures)
		 VALUES ('1', 1, true, '{"1"}')`,
	)
	suite.Require().NoError(err)

	err = suite.module.db.SaveGroupWithPolicy(
		types.NewGroupWithPolicy(1, "cudos1", "", "", threshold, 1, 0),
	)
	suite.Require().NoError(err)

	members := []group.MemberRequest{
		{Address: "cudos1", Weight: "1", Metadata: "1"},
	}
	err = suite.module.db.SaveGroupMembers(members, 1)
	suite.Require().NoError(err)
}

func (*GroupModuleTestSuite) newTestProposalMsg(tryExec group.Exec) group.MsgSubmitProposal {
	proposalJson := `{
		"group_policy_address": "cudos1",
		"proposers": [
			"cudos1"
		],
		"metadata": "",
		"messages": [
			{
				"@type": "/cosmos.group.v1.MsgUpdateGroupMetadata",
				"admin": "cudos1",
				"group_id": 1,
				"metadata": ""
			}
		]
	}`
	var proposal group.MsgSubmitProposal
	json.Unmarshal([]byte(proposalJson), &proposal)
	proposal.Exec = tryExec
	return proposal
}
