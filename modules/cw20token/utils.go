package cw20token

import (
	"os"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	cudosnodesimapp "github.com/CudoVentures/cudos-node/simapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/types"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"
)

func (m *Module) saveTokenInfo(dbTx *database.DbTx, contract string, codeID uint64, height int64) error {
	tokenInfo, err := m.source.GetTokenInfo(contract, height)
	if err != nil {
		return err
	}

	tokenInfo.Address = contract
	tokenInfo.CodeID = codeID

	if err := dbTx.SaveTokenInfo(tokenInfo); err != nil {
		return err
	}

	if err := dbTx.SaveTokenBalances(contract, tokenInfo.Balances); err != nil {
		return err
	}

	return nil
}

func (m *Module) saveBalances(dbTx *database.DbTx, contract string, sender string, msg *types.MsgTokenExecute, height int64) error {
	balances := []*types.TokenBalance{}

	if msg.Owner != "" {
		sender = msg.Owner
	}
	balances = append(balances, &types.TokenBalance{Address: sender})

	if msg.Recipient != "" {
		balances = append(balances, &types.TokenBalance{Address: msg.Recipient})
	}

	if msg.Contract != "" {
		balances = append(balances, &types.TokenBalance{Address: msg.Contract})
	}

	for _, a := range balances {
		balance, err := m.source.GetBalance(contract, a.Address, height)
		if err != nil {
			return err
		}

		a.Amount = balance
	}

	return dbTx.SaveTokenBalances(contract, balances)
}

func (m *Module) saveCirculatingSupply(dbTx *database.DbTx, contract string, height int64) error {
	supply, err := m.source.GetCirculatingSupply(contract, height)
	if err != nil {
		return err
	}

	return dbTx.UpdateTokenCirculatingSupply(contract, supply)
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
