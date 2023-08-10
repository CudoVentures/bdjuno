package mockutils

import (
	"testing"
	"time"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	juno "github.com/forbole/juno/v5/types"

	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

type MockTxBuilder struct {
	events    []abcitypes.Event
	timestamp time.Time
	txHash    string
	height    uint64
	t         *testing.T
}

func NewMockTxBuilder(t *testing.T, timestamp time.Time, txHash string, height uint64) *MockTxBuilder {
	return &MockTxBuilder{timestamp: timestamp, txHash: txHash, height: height, t: t}
}

func (b *MockTxBuilder) WithEventCreateGroup(groupID uint64, address string) *MockTxBuilder {
	require.NotEmpty(b.t, address)
	e, err := sdk.TypedEventToEvent(&group.EventCreateGroup{GroupId: groupID})
	require.NoError(b.t, err)

	e2, err := sdk.TypedEventToEvent(&group.EventCreateGroupPolicy{Address: address})
	require.NoError(b.t, err)

	b.events = append(b.events, abcitypes.Event(e), abcitypes.Event(e2))
	return b
}

func (b *MockTxBuilder) WithEventSubmitProposal(proposalID uint64) *MockTxBuilder {
	e, err := sdk.TypedEventToEvent(&group.EventSubmitProposal{ProposalId: proposalID})
	require.NoError(b.t, err)

	b.events = append(b.events, abcitypes.Event(e))
	return b
}

func (b *MockTxBuilder) WithEventExec(result group.ProposalExecutorResult) *MockTxBuilder {
	e, err := sdk.TypedEventToEvent(&group.EventExec{Result: result, Logs: "1"})
	require.NoError(b.t, err)

	b.events = append(b.events, abcitypes.Event(e))
	return b
}

func (b *MockTxBuilder) WithEventVote() *MockTxBuilder {
	e, err := sdk.TypedEventToEvent(&group.EventVote{ProposalId: 1})
	require.NoError(b.t, err)

	b.events = append(b.events, abcitypes.Event(e))
	return b
}

func (b *MockTxBuilder) WithEventWithdrawProposal() *MockTxBuilder {
	e, err := sdk.TypedEventToEvent(&group.EventWithdrawProposal{ProposalId: 1})
	require.NoError(b.t, err)

	b.events = append(b.events, abcitypes.Event(e))
	return b
}

func (b *MockTxBuilder) WithEventInstantiateContract(contractAddr string) *MockTxBuilder {
	e := sdk.NewEvent(wasm.EventTypeInstantiate, sdk.NewAttribute(wasm.AttributeKeyContractAddr, contractAddr))
	b.events = append(b.events, abcitypes.Event(e))
	return b
}

func (b *MockTxBuilder) WithEventWasmAction(msgType string) *MockTxBuilder {
	e := sdk.NewEvent(wasm.WasmModuleEventType, sdk.NewAttribute("action", msgType))
	b.events = append(b.events, abcitypes.Event(e))
	return b
}

func (b *MockTxBuilder) Build() *juno.Tx {
	txLog := sdk.ABCIMessageLogs{{MsgIndex: 0, Events: sdk.StringifyEvents(b.events)}}
	txResponse := sdk.TxResponse{
		TxHash:    b.txHash,
		Logs:      txLog,
		Timestamp: b.timestamp.Format(time.RFC3339),
		Height:    int64(b.height),
	}

	return &juno.Tx{TxResponse: &txResponse}
}
