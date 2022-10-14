package cw20token

import (
	"context"
	"os"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	cudosapp "github.com/CudoVentures/cudos-node/app"
	csimapp "github.com/CudoVentures/cudos-node/simapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/forbole/bdjuno/v2/utils/pubsub"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func GetPubSubClient(cfgBytes []byte) *pubsub.GooglePubSubClient {
	cfg, err := ParseConfig(cfgBytes)
	if err != nil {
		panic(err)
	}

	client, err := pubsub.NewGooglePubSubClient(context.Background(), cfg.ProjectID, cfg.SubID)
	if err != nil {
		panic(err)
	}

	return client
}

func GetWasmKeeper(homePath string, db dbm.DB) *wasmkeeper.Keeper {
	app := csimapp.NewSimApp(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, homePath, 0, csimapp.MakeTestEncodingConfig(), simapp.EmptyAppOptions{},
	)

	cudosapp.SetConfig()

	keeper := wasmkeeper.NewKeeper(
		app.AppCodec(),
		app.GetKey(wasm.StoreKey),
		app.GetSubspace(wasm.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.DistrKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		nil,
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		homePath,
		wasm.DefaultWasmConfig(),
		"iterator,staking,stargate",
	)

	keeper.SetParams(app.NewContext(false, tmproto.Header{Time: time.Now()}), wasmtypes.DefaultParams())

	return &keeper
}
