package fake

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	wasmapp "github.com/CosmWasm/wasmd/app"
	"github.com/forbole/bdjuno/v2/modules/cw20token/source"
	"github.com/forbole/bdjuno/v2/modules/cw20token/source/query"
	"github.com/forbole/bdjuno/v2/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasm "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ source.Source = &FakeSource{}
)

type FakeSource struct {
	TokenAddr string
	app       *wasmapp.WasmApp
	ctx       sdk.Context
	q         *query.QueryHandler
}

func SetupFakeSource(t *testing.T, token types.TokenInfo) (*FakeSource, error) {
	if len(token.Balances) == 0 {
		return nil, fmt.Errorf("token has no balances")
	}

	app := wasmapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})
	k := wasmapp.NewTestSupport(t, app).WasmKeeper()
	q := query.QueryHandlerLocal(wasmkeeper.Querier(&k).SmartContractState)

	s := &FakeSource{
		app: app,
		ctx: ctx,
		q:   q,
	}

	cw20Code, err := os.ReadFile("../../testdata/cw20_base.wasm")
	if err != nil {
		return nil, err
	}

	signer := token.Balances[0].Address

	resStoreCode, err := s.StoreCode(func(msg *wasm.MsgStoreCode) {
		msg.WASMByteCode = cw20Code
		msg.Sender = signer
	})
	if err != nil {
		return nil, err
	}

	msgRaw, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}

	_, err = s.Instantiate(func(m *wasm.MsgInstantiateContract) {
		m.Sender = signer
		m.Funds = sdk.Coins{}
		m.Msg = msgRaw
		m.CodeID = resStoreCode.CodeID
	})

	return s, err
}

func (s *FakeSource) StoreCode(mutators ...func(msg *wasm.MsgStoreCode)) (*wasm.MsgStoreCodeResponse, error) {
	msg := wasm.MsgStoreCodeFixture(mutators...)

	res := wasm.MsgStoreCodeResponse{}
	err := s.handleMsg(msg, &res)

	return &res, err
}

func (s *FakeSource) Instantiate(mutators ...func(msg *wasm.MsgInstantiateContract)) (*wasm.MsgInstantiateContractResponse, error) {
	msg := wasm.MsgInstantiateContractFixture(mutators...)

	res := wasm.MsgInstantiateContractResponse{}
	if err := s.handleMsg(msg, &res); err != nil {
		return nil, err
	}

	s.TokenAddr = res.Address
	return &res, nil
}

func (s *FakeSource) Execute(mutators ...func(msg *wasm.MsgExecuteContract)) (*wasm.MsgExecuteContractResponse, error) {
	msg := wasm.MsgExecuteContractFixture(mutators...)

	res := wasm.MsgExecuteContractResponse{}
	err := s.handleMsg(msg, &res)

	return &res, err
}

func (s *FakeSource) handleMsg(msg sdk.Msg, dest codec.ProtoMarshaler) error {
	if dest == nil {
		return fmt.Errorf("dest is nil")
	}

	res, err := s.app.MsgServiceRouter().Handler(msg)(s.ctx, msg)
	if err != nil {
		return err
	}

	return s.app.AppCodec().Unmarshal(res.Data, dest)
}

func (s *FakeSource) TokenInfo(tokenAddr string, height int64) (types.TokenInfo, error) {
	return s.q.TokenInfo(s.ctx, tokenAddr, height)
}

func (s *FakeSource) AllBalances(tokenAddr string, height int64) ([]types.TokenBalance, error) {
	return s.q.AllBalances(s.ctx, tokenAddr, height)
}

func (s *FakeSource) Balance(tokenAddr string, address string, height int64) (uint64, error) {
	return s.q.Balance(s.ctx, tokenAddr, address, height)
}

func (s *FakeSource) TotalSupply(tokenAddr string, height int64) (uint64, error) {
	return s.q.TotalSupply(s.ctx, tokenAddr, height)
}
