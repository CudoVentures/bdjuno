package cw20token

import (
	"context"
	"os"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	cudosnodesimapp "github.com/CudoVentures/cudos-node/simapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	ps "github.com/forbole/bdjuno/v2/utils"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"
)

func GetPubSubClient(cfgBytes []byte) *ps.GooglePubSubClient {
	cfg, err := ParseConfig(cfgBytes)
	if err != nil {
		panic(err)
	}

	client, err := ps.NewGooglePubSubClient(context.Background(), cfg.ProjectID, cfg.SubID)
	if err != nil {
		panic(err)
	}

	return client
}

func GetWasmKeeper(homePath string, db tmdb.DB) *wasmkeeper.Keeper {
	app := cudosnodesimapp.NewSimApp(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{},
		homePath, 0, simapp.MakeTestEncodingConfig(), simapp.EmptyAppOptions{},
	)

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
		app.CapabilityKeeper.ScopeToModule(wasm.ModuleName),
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		"wasmDir",
		wasm.DefaultWasmConfig(),
		"supportedFeatures",
	)

	return &keeper
}
