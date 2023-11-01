package main

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/forbole/juno/v5/cmd"
	initcmd "github.com/forbole/juno/v5/cmd/init"
	parsetypes "github.com/forbole/juno/v5/cmd/parse/types"
	startcmd "github.com/forbole/juno/v5/cmd/start"
	"github.com/forbole/juno/v5/modules/messages"

	cudosapp "github.com/CudoVentures/cudos-node/app"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	databasemigratecmd "github.com/forbole/bdjuno/v4/cmd/database-migrate"
	migratecmd "github.com/forbole/bdjuno/v4/cmd/migrate"
	parsegenesiscmd "github.com/forbole/bdjuno/v4/cmd/parse-genesis"
	"github.com/forbole/bdjuno/v4/workers"

	"github.com/forbole/bdjuno/v4/types/config"

	"cosmossdk.io/simapp"

	"github.com/forbole/bdjuno/v4/database"
	"github.com/forbole/bdjuno/v4/modules"
)

func main() {
	initCfg := initcmd.NewConfig().
		WithConfigCreator(config.Creator)

	parseCfg := parsetypes.NewConfig().
		WithDBBuilder(database.Builder).
		WithEncodingConfigBuilder(config.MakeEncodingConfig(getBasicManagers())).
		WithRegistrar(modules.NewRegistrar(getAddressesParser()))

	cfg := cmd.NewConfig("bdjuno").
		WithInitConfig(initCfg).
		WithParseConfig(parseCfg)

	cfgName := cfg.GetName()

	// Run the command
	rootCmd := cmd.RootCmd(cfgName)

	startcmd := startcmd.NewStartCmd(parseCfg)
	startcmd.PreRunE = workers.GetStartWorkersPrerunE(startcmd.PreRunE, parseCfg)

	rootCmd.AddCommand(
		cmd.VersionCmd(),
		initcmd.NewInitCmd(cfg.GetInitConfig()),
		migratecmd.NewMigrateCmd(cfgName, parseCfg),
		startcmd,
		parsegenesiscmd.NewParseGenesisCmd(parseCfg),
		databasemigratecmd.NewDatabaseMigrateCmd(parseCfg),
	)

	executor := cmd.PrepareRootCmd(cfgName, rootCmd)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

// getBasicManagers returns the various basic managers that are used to register the encoding to
// support custom messages.
// This should be edited by custom implementations if needed.
func getBasicManagers() []module.BasicManager {
	return []module.BasicManager{
		simapp.ModuleBasics,
		cudosapp.ModuleBasics,
		module.NewBasicManager(groupmodule.AppModuleBasic{}),
	}
}

// getAddressesParser returns the messages parser that should be used to get the users involved in
// a specific message.
// This should be edited by custom implementations if needed.
func getAddressesParser() messages.MessageAddressesParser {
	return messages.JoinMessageParsers(
		messages.CosmosMessageAddressesParser,
	)
}
