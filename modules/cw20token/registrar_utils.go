package cw20token

import (
	"os"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	cudosapp "github.com/CudoVentures/cudos-node/app"
	"github.com/CudoVentures/cudos-node/simapp"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func GetWasmKeeper(homePath string, db dbm.DB) *wasmkeeper.Keeper {
	cudosapp.SetConfig()

	app := simapp.NewSimApp(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, homePath, 0, simapp.MakeTestEncodingConfig(), simapp.EmptyAppOptions{},
	)

	// todo update cudos-node simapp
	// add iterator feature
	// cudosapp.SetConfig() for token prefix (must be called before initializing the simapp, otherwise it's already sealed)
	// add time to keeper.SetParams() header cuz nil time panics (we also add wasm defaultParams, which may be called in NewSimApp())
	// then we can remove this function
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
		homePath,
		wasm.DefaultWasmConfig(),
		"iterator,staking,stargate",
	)

	keeper.SetParams(app.NewContext(false, tmproto.Header{Time: time.Now()}), wasmtypes.DefaultParams())

	return &keeper
}
