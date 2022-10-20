package utils

import (
	"fmt"
	"strings"
	"time"

	abcitypes "github.com/tendermint/tendermint/abci/types"

	juno "github.com/forbole/juno/v2/types"

	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

type TxBuilder struct {
	events    []abcitypes.Event
	errors    []string
	timestamp time.Time
	txHash    string
	height    uint64
}

func NewTx(timestamp time.Time, txHash string, height uint64) *TxBuilder {
	return &TxBuilder{timestamp: timestamp, txHash: txHash, height: height}
}

func (b *TxBuilder) WithEventCreateGroup(groupID uint64, address string) *TxBuilder {
	if address == "" {
		b.errors = append(b.errors, "error while building testTx: empty group address")

	}
	eventCreateGroup, err := sdk.TypedEventToEvent(&group.EventCreateGroup{GroupId: groupID})
	if err != nil {
		b.errors = append(b.errors, err.Error())
	}

	eventCreateGroupPolicy, err := sdk.TypedEventToEvent(&group.EventCreateGroupPolicy{Address: address})
	if err != nil {
		b.errors = append(b.errors, err.Error())
	}

	b.events = append(b.events, abcitypes.Event(eventCreateGroup), abcitypes.Event(eventCreateGroupPolicy))
	return b
}

func (builder *TxBuilder) WithEventSubmitProposal(proposalID uint64) *TxBuilder {
	eventSubmitProposal, err := sdk.TypedEventToEvent(&group.EventSubmitProposal{ProposalId: proposalID})
	if err != nil {
		builder.errors = append(builder.errors, err.Error())
	}

	builder.events = append(builder.events, abcitypes.Event(eventSubmitProposal))
	return builder
}

func (builder *TxBuilder) WithEventExec(result group.ProposalExecutorResult) *TxBuilder {
	eventExec, err := sdk.TypedEventToEvent(&group.EventExec{Result: result, Logs: "1"})
	if err != nil {
		builder.errors = append(builder.errors, err.Error())
	}

	builder.events = append(builder.events, abcitypes.Event(eventExec))
	return builder
}

func (builder *TxBuilder) WithEventVote() *TxBuilder {
	eventVote, err := sdk.TypedEventToEvent(&group.EventVote{ProposalId: 1})
	if err != nil {
		builder.errors = append(builder.errors, err.Error())
	}

	builder.events = append(builder.events, abcitypes.Event(eventVote))
	return builder
}

func (builder *TxBuilder) WithEventWithdrawProposal() *TxBuilder {
	eventWithdraw, err := sdk.TypedEventToEvent(&group.EventWithdrawProposal{ProposalId: 1})
	if err != nil {
		builder.errors = append(builder.errors, err.Error())
	}

	builder.events = append(builder.events, abcitypes.Event(eventWithdraw))
	return builder
}

func (builder *TxBuilder) WithEventInstantiateContract(contractAddr string) *TxBuilder {
	eventInstantiateContract := sdk.NewEvent(wasm.EventTypeInstantiate, sdk.NewAttribute(wasm.AttributeKeyContractAddr, contractAddr))
	builder.events = append(builder.events, abcitypes.Event(eventInstantiateContract))
	return builder
}

func (builder *TxBuilder) Build() (*juno.Tx, error) {
	if len(builder.errors) > 0 {
		return &juno.Tx{}, fmt.Errorf(`error while building testTx: %s`, strings.Join(builder.errors, "\n"))
	}
	txLog := sdk.ABCIMessageLogs{{MsgIndex: 0, Events: sdk.StringifyEvents(builder.events)}}
	txResponse := sdk.TxResponse{
		TxHash:    builder.txHash,
		Logs:      txLog,
		Timestamp: builder.timestamp.Format(time.RFC3339),
		Height:    int64(builder.height),
	}

	return &juno.Tx{TxResponse: &txResponse}, nil
}
