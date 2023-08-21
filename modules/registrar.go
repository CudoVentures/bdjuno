package modules

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/forbole/bdjuno/v4/modules/actions"
	"github.com/forbole/bdjuno/v4/modules/types"

	"github.com/forbole/juno/v5/modules/pruning"
	"github.com/forbole/juno/v5/modules/telemetry"

	"github.com/forbole/bdjuno/v4/modules/slashing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	jmodules "github.com/forbole/juno/v5/modules"
	"github.com/forbole/juno/v5/modules/messages"
	"github.com/forbole/juno/v5/modules/registrar"

	"github.com/forbole/bdjuno/v4/utils"

	"github.com/forbole/bdjuno/v4/database"
	"github.com/forbole/bdjuno/v4/modules/auth"
	"github.com/forbole/bdjuno/v4/modules/bank"
	"github.com/forbole/bdjuno/v4/modules/consensus"
	"github.com/forbole/bdjuno/v4/modules/distribution"
	"github.com/forbole/bdjuno/v4/modules/feegrant"
	"github.com/forbole/bdjuno/v4/modules/gravity"

	"github.com/forbole/bdjuno/v4/modules/gov"
	"github.com/forbole/bdjuno/v4/modules/modules"
	"github.com/forbole/bdjuno/v4/modules/pricefeed"
	"github.com/forbole/bdjuno/v4/modules/staking"
	"github.com/forbole/bdjuno/v4/modules/upgrade"

	"github.com/forbole/bdjuno/v4/client/cryptocompare"
	"github.com/forbole/bdjuno/v4/modules/cosmwasm"
	"github.com/forbole/bdjuno/v4/modules/cudomint"
	"github.com/forbole/bdjuno/v4/modules/group"
	"github.com/forbole/bdjuno/v4/modules/history"
	"github.com/forbole/bdjuno/v4/modules/marketplace"
	"github.com/forbole/bdjuno/v4/modules/nft"
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
	cdc := ctx.EncodingConfig.Codec
	db := database.Cast(ctx.Database)

	sources, err := types.BuildSources(ctx.JunoConfig.Node, ctx.EncodingConfig)
	if err != nil {
		panic(err)
	}

	var cryptoCompareConfig cryptocompare.Config
	configBytes, err := ctx.JunoConfig.GetBytes()
	if err != nil {
		panic(fmt.Errorf("failed to get bytes from JunoConfig: %s", err))
	}
	if err := yaml.Unmarshal(configBytes, &cryptoCompareConfig); err != nil {
		panic(fmt.Errorf("failed to parse cryptoCompare config: %s", err))
	}

	cryptoCompareClient := cryptocompare.NewClient(&cryptoCompareConfig)

	actionsModule := actions.NewModule(cdc, ctx.JunoConfig, ctx.EncodingConfig)
	authModule := auth.NewModule(r.parser, cdc, db)
	bankModule := bank.NewModule(r.parser, sources.BankSource, cdc, db)
	consensusModule := consensus.NewModule(db)
	distrModule := distribution.NewModule(sources.DistrSource, cdc, db)
	feegrantModule := feegrant.NewModule(cdc, db)
	historyModule := history.NewModule(ctx.JunoConfig.Chain, r.parser, cdc, db)
	cudoMintModule := cudomint.NewModule(cdc, db, configBytes)
	slashingModule := slashing.NewModule(sources.SlashingSource, cdc, db)
	stakingModule := staking.NewModule(sources.StakingSource, cdc, db)
	govModule := gov.NewModule(sources.GovSource, authModule, distrModule, slashingModule, stakingModule, cdc, db)
	cosmwasmModule := cosmwasm.NewModule(cdc, db)
	gravityModule := gravity.NewModule(cdc, db)
	nftModule := nft.NewModule(cdc, db)
	groupModule := group.NewModule(cdc, db)
	marketplaceModule := marketplace.NewModule(cdc, db, configBytes, cryptoCompareClient)
	upgradeModule := upgrade.NewModule(db, stakingModule)

	return []jmodules.Module{
		messages.NewModule(r.parser, cdc, ctx.Database),
		telemetry.NewModule(ctx.JunoConfig),
		pruning.NewModule(ctx.JunoConfig, db, ctx.Logger),

		actionsModule,
		authModule,
		bankModule,
		consensusModule,
		distrModule,
		feegrantModule,
		govModule,
		historyModule,
		cudoMintModule,
		modules.NewModule(ctx.JunoConfig.Chain, db),
		pricefeed.NewModule(ctx.JunoConfig, cryptoCompareClient, historyModule, cdc, db),
		slashingModule,
		stakingModule,
		cosmwasmModule,
		gravityModule,
		marketplaceModule,
		nftModule,
		groupModule,
		upgradeModule,
	}
}
