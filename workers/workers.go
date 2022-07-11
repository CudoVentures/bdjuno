package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/juno/v2/cmd/parse"
	"github.com/forbole/juno/v2/types/config"
	"github.com/spf13/cobra"
)

type PreRunE func(cmd *cobra.Command, args []string) error

type worker interface {
	Name() string
	Start(ctx context.Context, parseCfg *parse.Config, parseCtx *parse.Context, storage keyValueStorage, interval time.Duration)
}

var cancelWorkersCtx context.CancelFunc

var workers = []worker{
	fixBlocksWorker{},
	migrateNftsWorker{},
}

func GetStartWorkersPrerunE(origPreRunE PreRunE, parseCfg *parse.Config) PreRunE {
	return func(cmd *cobra.Command, args []string) error {
		if err := parse.ReadConfig(parseCfg)(nil, nil); err != nil {
			return err
		}

		parseCtx, err := parse.GetParsingContext(parseCfg)
		if err != nil {
			return err
		}

		cfg, err := parseConfig(config.Cfg.GetBytes())
		if err != nil {
			return err
		}

		if err := startWorkers(context.Background(), workers, cfg, parseCfg, parseCtx); err != nil {
			return err
		}

		return origPreRunE(cmd, args)
	}
}

func startWorkers(ctx context.Context, workers []worker, cfg workersConfig, parseCfg *parse.Config, parseCtx *parse.Context) error {
	var workersCtx context.Context
	workersCtx, cancelWorkersCtx = context.WithCancel(ctx)

	for _, w := range workers {
		wcfg, err := getWorkerConfig(cfg, w.Name())
		if err != nil {
			return err
		}

		interval, err := time.ParseDuration(wcfg.Interval)
		if err != nil {
			return err
		}

		w.Start(workersCtx, parseCfg, parseCtx, NewWorkersStorage(database.Cast(parseCtx.Database), w.Name()), interval)
	}

	return nil
}

func getWorkerConfig(cfg workersConfig, name string) (workerConfig, error) {
	for idx := range cfg.Workers {
		if cfg.Workers[idx].Name == name {
			return cfg.Workers[idx], nil
		}
	}
	return workerConfig{}, fmt.Errorf("worker config for %s not found", name)
}

func StopWorkers() {
	cancelWorkersCtx()
}
