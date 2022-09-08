package main

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/forbole/juno/v3/cmd"
	initcmd "github.com/forbole/juno/v3/cmd/init"
	parsecmd "github.com/forbole/juno/v3/cmd/parse"
	"github.com/forbole/juno/v3/modules/messages"

	actionscmd "github.com/forbole/bdjuno/v2/cmd/actions"
	databasemigratecmd "github.com/forbole/bdjuno/v2/cmd/database-migrate"
	parsegenesiscmd "github.com/forbole/bdjuno/v2/cmd/parse-genesis"
	"github.com/forbole/bdjuno/v2/workers"
	startcmd "github.com/forbole/juno/v3/cmd/start"

	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules"
	"github.com/forbole/bdjuno/v2/types/config"
	parsetypes "github.com/forbole/juno/v3/cmd/parse/types"

	cudosapp "github.com/CudoVentures/cudos-node/app"
)

func main() {
	parseCfg := parsetypes.NewConfig().
		WithDBBuilder(database.Builder).
		WithEncodingConfigBuilder(config.MakeEncodingConfig(getBasicManagers())).
		WithRegistrar(modules.NewRegistrar(getAddressesParser()))

	cfg := cmd.NewConfig("bdjuno").WithParseConfig(parseCfg)

	// Run the command
	rootCmd := cmd.RootCmd(cfg.GetName())

	pcmd := parsecmd.NewParseCmd(cfg.GetParseConfig())
	pcmd.PreRunE = workers.GetStartWorkersPrerunE(pcmd.PreRunE, cfg.GetParseConfig())

	rootCmd.AddCommand(
		cmd.VersionCmd(),
		initcmd.NewInitCmd(cfg.GetInitConfig()),
		pcmd,
		parsegenesiscmd.NewParseGenesisCmd(cfg.GetParseConfig()),
		actionscmd.NewActionsCmd(cfg.GetParseConfig()),
		databasemigratecmd.NewDatabaseMigrateCmd(cfg.GetParseConfig()),
		startcmd.NewStartCmd(cfg.GetParseConfig()),
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
