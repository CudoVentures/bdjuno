package main

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/forbole/juno/v2/cmd"
	initcmd "github.com/forbole/juno/v2/cmd/init"
	parsecmd "github.com/forbole/juno/v2/cmd/parse"
	"github.com/forbole/juno/v2/modules/messages"

	actionscmd "github.com/forbole/bdjuno/v2/cmd/actions"
	databasemigratecmd "github.com/forbole/bdjuno/v2/cmd/database-migrate"
	fixcmd "github.com/forbole/bdjuno/v2/cmd/fix"
	migratecmd "github.com/forbole/bdjuno/v2/cmd/migrate"
	parsegenesiscmd "github.com/forbole/bdjuno/v2/cmd/parse-genesis"
	"github.com/forbole/bdjuno/v2/workers"

	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules"
	"github.com/forbole/bdjuno/v2/types/config"

	cudosapp "github.com/CudoVentures/cudos-node/app"
)

func main() {
	parseCfg := parsecmd.NewConfig().
		WithDBBuilder(database.Builder).
		WithEncodingConfigBuilder(config.MakeEncodingConfig(getBasicManagers())).
		WithRegistrar(modules.NewRegistrar(getAddressesParser()))

	cfg := cmd.NewConfig("bdjuno").
		WithParseConfig(parseCfg)

	// Run the command
	rootCmd := cmd.RootCmd(cfg.GetName())

	pcmd := parsecmd.ParseCmd(cfg.GetParseConfig())
	pcmd.PreRunE = workers.GetStartWorkersPrerunE(pcmd.PreRunE, cfg.GetParseConfig())

	rootCmd.AddCommand(
		cmd.VersionCmd(),
		initcmd.InitCmd(cfg.GetInitConfig()),
		pcmd,
		migratecmd.NewMigrateCmd(),
		fixcmd.NewFixCmd(cfg.GetParseConfig()),
		parsegenesiscmd.NewParseGenesisCmd(cfg.GetParseConfig()),
		actionscmd.NewActionsCmd(cfg.GetParseConfig()),
		databasemigratecmd.NewDatabaseMigrateCmd(cfg.GetParseConfig()),
	)

	executor := cmd.PrepareRootCmd(cfg.GetName(), rootCmd)
	if err := executor.Execute(); err != nil {
		panic(err)
	}
}

// getBasicManagers returns the various basic managers that are used to register the encoding to
// support custom messages.
// This should be edited by custom implementations if needed.
func getBasicManagers() []module.BasicManager {
	return []module.BasicManager{
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
