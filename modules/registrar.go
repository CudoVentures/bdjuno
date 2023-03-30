package modules

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/tendermint/tendermint/libs/log"
	"gopkg.in/yaml.v2"

	"github.com/forbole/juno/v2/modules/pruning"
	"github.com/forbole/juno/v2/modules/telemetry"

	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/forbole/juno/v2/node/remote"

	"github.com/forbole/bdjuno/v2/modules/cosmwasm"
	"github.com/forbole/bdjuno/v2/modules/cudomint"
	"github.com/forbole/bdjuno/v2/modules/gravity"
	"github.com/forbole/bdjuno/v2/modules/history"
	"github.com/forbole/bdjuno/v2/modules/marketplace"
	"github.com/forbole/bdjuno/v2/modules/nft"
	"github.com/forbole/bdjuno/v2/modules/slashing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/forbole/juno/v2/node/local"

	jmodules "github.com/forbole/juno/v2/modules"
	"github.com/forbole/juno/v2/modules/messages"
	"github.com/forbole/juno/v2/modules/registrar"

	"github.com/forbole/bdjuno/v2/utils"

	nodeconfig "github.com/forbole/juno/v2/node/config"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/auth"
	"github.com/forbole/bdjuno/v2/modules/bank"
	banksource "github.com/forbole/bdjuno/v2/modules/bank/source"
	localbanksource "github.com/forbole/bdjuno/v2/modules/bank/source/local"
	remotebanksource "github.com/forbole/bdjuno/v2/modules/bank/source/remote"
	"github.com/forbole/bdjuno/v2/modules/consensus"
	"github.com/forbole/bdjuno/v2/modules/distribution"
	"github.com/forbole/bdjuno/v2/modules/feegrant"

	"github.com/forbole/bdjuno/v2/client/cryptoCompare"
	"github.com/forbole/bdjuno/v2/modules/cw20token"
	cw20tokensource "github.com/forbole/bdjuno/v2/modules/cw20token/source"
	localcw20tokensource "github.com/forbole/bdjuno/v2/modules/cw20token/source/local"
	remotecw20tokensource "github.com/forbole/bdjuno/v2/modules/cw20token/source/remote"
	distrsource "github.com/forbole/bdjuno/v2/modules/distribution/source"
	localdistrsource "github.com/forbole/bdjuno/v2/modules/distribution/source/local"
	remotedistrsource "github.com/forbole/bdjuno/v2/modules/distribution/source/remote"
	"github.com/forbole/bdjuno/v2/modules/gov"
	govsource "github.com/forbole/bdjuno/v2/modules/gov/source"
	localgovsource "github.com/forbole/bdjuno/v2/modules/gov/source/local"
	remotegovsource "github.com/forbole/bdjuno/v2/modules/gov/source/remote"
	"github.com/forbole/bdjuno/v2/modules/group"
	"github.com/forbole/bdjuno/v2/modules/modules"
	"github.com/forbole/bdjuno/v2/modules/pricefeed"
	slashingsource "github.com/forbole/bdjuno/v2/modules/slashing/source"
	localslashingsource "github.com/forbole/bdjuno/v2/modules/slashing/source/local"
	remoteslashingsource "github.com/forbole/bdjuno/v2/modules/slashing/source/remote"
	"github.com/forbole/bdjuno/v2/modules/staking"
	stakingsource "github.com/forbole/bdjuno/v2/modules/staking/source"
	localstakingsource "github.com/forbole/bdjuno/v2/modules/staking/source/local"
	remotestakingsource "github.com/forbole/bdjuno/v2/modules/staking/source/remote"
)

// UniqueAddressesParser returns a wrapper around the given parser that removes all duplicated addresses
func UniqueAddressesParser(parser messages.MessageAddressesParser) messages.MessageAddressesParser {
	return func(cdc codec.Codec, msg sdk.Msg) ([]string, error) {
		addresses, err := parser(cdc, msg)
		if err != nil {
			return nil, err
		}

		return utils.RemoveDuplicateValues(addresses), nil
	}
}

// --------------------------------------------------------------------------------------------------------------------

var (
	_ registrar.Registrar = &Registrar{}
)

// Registrar represents the modules.Registrar that allows to register all modules that are supported by BigDipper
type Registrar struct {
	parser messages.MessageAddressesParser
}

// NewRegistrar allows to build a new Registrar instance
func NewRegistrar(parser messages.MessageAddressesParser) *Registrar {
	return &Registrar{
		parser: UniqueAddressesParser(parser),
	}
}

// BuildModules implements modules.Registrar
func (r *Registrar) BuildModules(ctx registrar.Context) jmodules.Modules {
	cdc := ctx.EncodingConfig.Marshaler
	db := database.Cast(ctx.Database)

	sources, err := BuildSources(ctx.JunoConfig.Node, ctx.EncodingConfig)
	if err != nil {
		panic(err)
	}

	var cryptoCompareConfig cryptoCompare.Config
	if err := yaml.Unmarshal(ctx.JunoConfig.GetBytes(), &cryptoCompareConfig); err != nil {
		panic(fmt.Errorf("failed to parse cudomint config: %s", err))
	}

	authModule := auth.NewModule(r.parser, cdc, db)
	bankModule := bank.NewModule(r.parser, sources.BankSource, cdc, db)
	consensusModule := consensus.NewModule(db)
	distrModule := distribution.NewModule(sources.DistrSource, cdc, db)
	feegrantModule := feegrant.NewModule(cdc, db)
	historyModule := history.NewModule(ctx.JunoConfig.Chain, r.parser, cdc, db)
	cudoMintModule := cudomint.NewModule(cdc, db, ctx.JunoConfig.GetBytes())
	slashingModule := slashing.NewModule(sources.SlashingSource, cdc, db)
	stakingModule := staking.NewModule(sources.StakingSource, slashingModule, authModule, cdc, db)
	govModule := gov.NewModule(sources.GovSource, authModule, distrModule, slashingModule, stakingModule, cdc, db)
	cosmwasmModule := cosmwasm.NewModule(cdc, db)
	gravityModule := gravity.NewModule(cdc, db)
	nftModule := nft.NewModule(cdc, db)
	groupModule := group.NewModule(cdc, db)
	marketplaceModule := marketplace.NewModule(cdc, db, ctx.JunoConfig.GetBytes(), cryptoCompareConfig)
	cw20tokenModule := cw20token.NewModule(cdc, db, sources.CW20TokenSource)

	return []jmodules.Module{
		messages.NewModule(r.parser, cdc, ctx.Database),
		telemetry.NewModule(ctx.JunoConfig),
		pruning.NewModule(ctx.JunoConfig, db, ctx.Logger),

		authModule,
		bankModule,
		consensusModule,
		distrModule,
		feegrantModule,
		govModule,
		historyModule,
		cudoMintModule,
		modules.NewModule(ctx.JunoConfig.Chain, db),
		pricefeed.NewModule(ctx.JunoConfig, cryptoCompareConfig, historyModule, cdc, db),
		slashingModule,
		stakingModule,
		cosmwasmModule,
		gravityModule,
		nftModule,
		groupModule,
		marketplaceModule,
		cw20tokenModule,
	}
}

type Sources struct {
	BankSource      banksource.Source
	DistrSource     distrsource.Source
	GovSource       govsource.Source
	SlashingSource  slashingsource.Source
	StakingSource   stakingsource.Source
	CW20TokenSource cw20tokensource.Source
}

func BuildSources(nodeCfg nodeconfig.Config, encodingConfig *params.EncodingConfig) (*Sources, error) {
	switch cfg := nodeCfg.Details.(type) {
	case *remote.Details:
		return buildRemoteSources(cfg)
	case *local.Details:
		return buildLocalSources(cfg, encodingConfig)

	default:
		return nil, fmt.Errorf("invalid configuration type: %T", cfg)
	}
}

func buildLocalSources(cfg *local.Details, encodingConfig *params.EncodingConfig) (*Sources, error) {
	source, err := local.NewSource(cfg.Home, encodingConfig)
	if err != nil {
		return nil, err
	}

	app := simapp.NewSimApp(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)), source.StoreDB, nil, true, map[int64]bool{},
		cfg.Home, 0, simapp.MakeTestEncodingConfig(), simapp.EmptyAppOptions{},
	)

	sources := &Sources{
		BankSource:      localbanksource.NewSource(source, banktypes.QueryServer(app.BankKeeper)),
		DistrSource:     localdistrsource.NewSource(source, distrtypes.QueryServer(app.DistrKeeper)),
		GovSource:       localgovsource.NewSource(source, govtypes.QueryServer(app.GovKeeper)),
		SlashingSource:  localslashingsource.NewSource(source, slashingtypes.QueryServer(app.SlashingKeeper)),
		StakingSource:   localstakingsource.NewSource(source, stakingkeeper.Querier{Keeper: app.StakingKeeper}),
		CW20TokenSource: localcw20tokensource.NewSource(source, wasmkeeper.Querier(cw20token.GetWasmKeeper(cfg.Home, source.StoreDB))),
	}

	// Mount and initialize the stores
	err = source.MountKVStores(app, "keys")
	if err != nil {
		return nil, err
	}

	err = source.MountTransientStores(app, "tkeys")
	if err != nil {
		return nil, err
	}

	err = source.MountMemoryStores(app, "memKeys")
	if err != nil {
		return nil, err
	}

	err = source.InitStores()
	if err != nil {
		return nil, err
	}

	return sources, nil
}

func buildRemoteSources(cfg *remote.Details) (*Sources, error) {
	source, err := remote.NewSource(cfg.GRPC)
	if err != nil {
		return nil, fmt.Errorf("error while creating remote source: %s", err)
	}

	return &Sources{
		BankSource:      remotebanksource.NewSource(source, banktypes.NewQueryClient(source.GrpcConn)),
		DistrSource:     remotedistrsource.NewSource(source, distrtypes.NewQueryClient(source.GrpcConn)),
		GovSource:       remotegovsource.NewSource(source, govtypes.NewQueryClient(source.GrpcConn)),
		SlashingSource:  remoteslashingsource.NewSource(source, slashingtypes.NewQueryClient(source.GrpcConn)),
		StakingSource:   remotestakingsource.NewSource(source, stakingtypes.NewQueryClient(source.GrpcConn)),
		CW20TokenSource: remotecw20tokensource.NewSource(source, wasmtypes.NewQueryClient(source.GrpcConn)),
	}, nil
}
