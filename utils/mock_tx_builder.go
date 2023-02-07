package utils

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"

	juno "github.com/forbole/juno/v2/types"

	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
	marketplace "github.com/CudoVentures/cudos-node/x/marketplace/types"
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
	e := abcitypes.Event(sdk.NewEvent(
		"cosmos.group.v1.EventCreateGroup",
		sdk.NewAttribute("group_id", strconv.FormatUint(groupID, 10)),
	))

	e2 := abcitypes.Event(sdk.NewEvent(
		"cosmos.group.v1.EventCreateGroupPolicy",
		sdk.NewAttribute("address", address),
	))

	b.events = append(b.events, e, e2)
	return b
}

func (b *MockTxBuilder) WithEventSubmitProposal(proposalID uint64) *MockTxBuilder {
	e := abcitypes.Event(sdk.NewEvent(
		"cosmos.group.v1.EventSubmitProposal",
		sdk.NewAttribute("proposal_id", strconv.FormatUint(proposalID, 10))),
	)

	b.events = append(b.events, e)
	return b
}

func (b *MockTxBuilder) WithEventExec(result group.ProposalExecutorResult) *MockTxBuilder {
	e := abcitypes.Event(sdk.NewEvent(
		"cosmos.group.v1.EventExec",
		sdk.NewAttribute("result", result.String()),
		sdk.NewAttribute("logs", "1"),
	))

	b.events = append(b.events, e)
	return b
}

func (b *MockTxBuilder) WithEventVote() *MockTxBuilder {
	e := abcitypes.Event(sdk.NewEvent(
		"cosmos.group.v1.EventVote",
		sdk.NewAttribute("proposal_id", "1"),
	))

	b.events = append(b.events, e)
	return b
}

func (b *MockTxBuilder) WithEventWithdrawProposal() *MockTxBuilder {
	e := abcitypes.Event(sdk.NewEvent(
		"cosmos.group.v1.EventWithdrawProposal",
		sdk.NewAttribute("proposal_id", "1"),
	))

	b.events = append(b.events, e)
	return b
}

func (b *MockTxBuilder) WithEventInstantiateContract(contractAddr string) *MockTxBuilder {
	e := abcitypes.Event(sdk.NewEvent(
		wasm.EventTypeInstantiate,
		sdk.NewAttribute(wasm.AttributeKeyContractAddr, contractAddr),
	))

	b.events = append(b.events, e)
	return b
}

func (b *MockTxBuilder) WithEventWasmAction(msgType string) *MockTxBuilder {
	e := abcitypes.Event(sdk.NewEvent(
		wasm.WasmModuleEventType,
		sdk.NewAttribute("action", msgType),
	))

	b.events = append(b.events, e)
	return b
}

func (b *MockTxBuilder) WithEventPublishAuction(auctionID uint64, startTime time.Time, endTime time.Time, auctionInfo string) *MockTxBuilder {
	e := abcitypes.Event(sdk.NewEvent(
		marketplace.EventPublishAuctionType,
		sdk.NewAttribute(marketplace.AttributeAuctionID, strconv.FormatUint(auctionID, 10)),
		sdk.NewAttribute(marketplace.AttributeStartTime, startTime.Format(time.RFC3339)),
		sdk.NewAttribute(marketplace.AttributeEndTime, endTime.Format(time.RFC3339)),
		sdk.NewAttribute(marketplace.AttributeAuctionInfo, auctionInfo),
	))

	b.events = append(b.events, e)
	return b
}

func (b *MockTxBuilder) WithEventBuyNftFromAuction(auctionID uint64, tokenID uint64, denomID string, buyer string) *MockTxBuilder {
	e := abcitypes.Event(sdk.NewEvent(
		marketplace.EventBuyNftType,
		sdk.NewAttribute(marketplace.AttributeAuctionID, strconv.FormatUint(auctionID, 10)),
		sdk.NewAttribute(marketplace.AttributeKeyTokenID, strconv.FormatUint(tokenID, 10)),
		sdk.NewAttribute(marketplace.AttributeKeyDenomID, denomID),
		sdk.NewAttribute(marketplace.AttributeKeyBuyer, buyer),
	))

	b.events = append(b.events, e)
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
