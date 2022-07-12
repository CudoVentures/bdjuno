package group

import (
	"testing"

	"github.com/tendermint/tendermint/abci/types"

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

func (suite *GroupModuleTestSuite) SetupSuite() {
	db, cdc := utils.NewTestDb(&suite.Suite, "groupTest")
	suite.module = NewModule(cdc, db)
}

func (suite *GroupModuleTestSuite) TestGroup_CreateGroupWithPolicy() {
	decisionPolicy, _ := codectypes.NewAnyWithValue(
		group.NewThresholdDecisionPolicy("1", 10000, 0),
	)

	msg := group.MsgCreateGroupWithPolicy{
		Admin: "admin",
		Members: []group.MemberRequest{
			{Address: "cudos1", Weight: "1", Metadata: "1"},
		},
		GroupMetadata:       "1",
		GroupPolicyMetadata: "1'",
		GroupPolicyAsAdmin:  true,
		DecisionPolicy:      decisionPolicy,
	}

	tx := newTestTx("1", 1, "cudos1", 0, 0)

	err := suite.module.HandleMsg(0, &msg, tx)
	suite.Require().NoError(err)

	var rows []dbtypes.GroupWithPolicyRow
	err = suite.module.db.Sqlx.Select(&rows, "SELECT * FROM group_with_policy")
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(rows))
}

func newTestTx(
	txHash string,
	groupID uint64,
	groupAddress string,
	proposalID uint64,
	executorResult group.ProposalExecutorResult,
) *juno.Tx {
	events := make([]types.Event, 0)

	if groupID != 0 {
		eventCreateGroup, _ := sdk.TypedEventToEvent(
			&group.EventCreateGroup{GroupId: groupID},
		)
		events = append(events, types.Event(eventCreateGroup))
	}

	if groupAddress != "" {
		eventCreateGroupPolicy, _ := sdk.TypedEventToEvent(
			&group.EventCreateGroupPolicy{Address: groupAddress},
		)
		events = append(events, types.Event(eventCreateGroupPolicy))
	}

	if proposalID != 0 {
		eventSubmitProposal, _ := sdk.TypedEventToEvent(
			&group.EventSubmitProposal{ProposalId: proposalID},
		)
		events = append(events, types.Event(eventSubmitProposal))
	}

	if executorResult != group.PROPOSAL_EXECUTOR_RESULT_UNSPECIFIED {
		eventExec, _ := sdk.TypedEventToEvent(&group.EventExec{
			Result: executorResult,
		})

		events = append(events, types.Event(eventExec))
	}

	txLog := sdk.ABCIMessageLogs{
		{MsgIndex: 0, Events: sdk.StringifyEvents(events)},
	}

	txResponse := sdk.TxResponse{
		TxHash: txHash,
		Logs:   txLog,
	}

	return &juno.Tx{TxResponse: &txResponse}
}
