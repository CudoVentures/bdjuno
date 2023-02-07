package utils

import (
	"testing"
	"time"

	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
	marketplace "github.com/CudoVentures/cudos-node/x/marketplace/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

const (
	num           = uint64(1)
	str           = "1"
	index         = uint32(0)
	resultDefault = group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN
)

var (
	timestamp = time.Date(2222, 1, 1, 1, 0, 0, 0, time.UTC)
	wantLogs  = sdk.ABCIMessageLogs{{MsgIndex: 0, Events: sdk.StringifyEvents([]abcitypes.Event{
		abcitypes.Event(sdk.NewEvent(
			"cosmos.group.v1.EventCreateGroup",
			sdk.NewAttribute("group_id", str),
		)),
		abcitypes.Event(sdk.NewEvent(
			"cosmos.group.v1.EventCreateGroupPolicy",
			sdk.NewAttribute("address", str),
		)),
		abcitypes.Event(sdk.NewEvent(
			"cosmos.group.v1.EventSubmitProposal",
			sdk.NewAttribute("proposal_id", str),
		)),
		abcitypes.Event(sdk.NewEvent(
			"cosmos.group.v1.EventExec",
			sdk.NewAttribute("result", resultDefault.String()),
			sdk.NewAttribute("logs", str),
		)),
		abcitypes.Event(sdk.NewEvent(
			"cosmos.group.v1.EventVote",
			sdk.NewAttribute("proposal_id", str),
		)),
		abcitypes.Event(sdk.NewEvent(
			"cosmos.group.v1.EventWithdrawProposal",
			sdk.NewAttribute("proposal_id", str),
		)),
		abcitypes.Event(sdk.NewEvent(
			wasm.EventTypeInstantiate,
			sdk.NewAttribute(wasm.AttributeKeyContractAddr, str),
		)),
		abcitypes.Event(sdk.NewEvent(
			wasm.WasmModuleEventType,
			sdk.NewAttribute("action", str),
		)),
		abcitypes.Event(sdk.NewEvent(
			marketplace.EventPublishAuctionType,
			sdk.NewAttribute(marketplace.AttributeAuctionID, str),
			sdk.NewAttribute(marketplace.AttributeStartTime, timestamp.Format(time.RFC3339)),
			sdk.NewAttribute(marketplace.AttributeEndTime, timestamp.Format(time.RFC3339)),
			sdk.NewAttribute(marketplace.AttributeAuctionInfo, str),
		)),
		abcitypes.Event(sdk.NewEvent(
			marketplace.EventBuyNftType,
			sdk.NewAttribute(marketplace.AttributeAuctionID, str),
			sdk.NewAttribute(marketplace.AttributeKeyTokenID, str),
			sdk.NewAttribute(marketplace.AttributeKeyDenomID, str),
			sdk.NewAttribute(marketplace.AttributeKeyBuyer, str),
		)),
	})}}
)

func TestMockTxBuilder_Build(t *testing.T) {
	tx := NewMockTxBuilder(t, timestamp, str, num).WithEventCreateGroup(num, str).WithEventSubmitProposal(num).WithEventExec(resultDefault).WithEventVote().WithEventWithdrawProposal().WithEventInstantiateContract(str).WithEventWasmAction(str).WithEventPublishAuction(num, timestamp, timestamp, str).WithEventBuyNftFromAuction(num, num, str, str).Build()

	require.Equal(t, wantLogs, tx.Logs)

	actualTxHash := tx.TxHash
	require.Equal(t, str, actualTxHash)

	actualHeight := tx.Height
	require.Equal(t, int64(num), actualHeight)
}
