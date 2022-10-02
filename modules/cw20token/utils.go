package cw20token

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	cudosnodesimapp "github.com/CudoVentures/cudos-node/simapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/cw20token/source"
	"github.com/forbole/bdjuno/v2/types"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"
)

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

func (m *Module) saveToken(dbTx *database.DbTx, contractAddress string) error {
	tokenInfo, err := getTokenInfo(dbTx, contractAddress, m.source)
	if err != nil {
		return err
	}

	if err := dbTx.SaveToken(tokenInfo); err != nil {
		return err
	}

	if err := dbTx.SaveTokenBalances(tokenInfo.Balances); err != nil {
		return err
	}
	return nil
}

func getTokenInfo(dbTx *database.DbTx, contractAddress string, source source.Source) (*types.TokenInfo, error) {
	block, err := dbTx.GetLastBlock()
	if err != nil {
		return nil, err
	}
	state, err := source.AllContractState(contractAddress, block.Height)
	if err != nil {
		return nil, err
	}

	tokenInfo := types.TokenInfo{}
	for _, s := range state {
		key := string(s.Key)

		if key == "token_info" {
			if err := json.Unmarshal(s.Value, &tokenInfo); err != nil {
				return nil, err
			}
			continue
		}

		if strings.Contains(key, "balance") {
			balance, err := strconv.ParseUint(strings.ReplaceAll(string(s.Value), "\"", ""), 10, 64)
			if err != nil {
				return nil, err
			}

			addressIndex := strings.Index(key, "cudos")
			address := key[addressIndex:]
			// todo test tokenInfo.Balances
			tokenInfo.Balances = append(tokenInfo.Balances, types.TokenBalance{Address: address, Amount: balance})
		}
	}

	tokenInfo.Address = contractAddress
	return &tokenInfo, nil
}
