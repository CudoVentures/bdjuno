package workers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/juno/v2/cmd/parse"
	"github.com/forbole/juno/v2/parser"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

type PersistentPreRunE func(cmd *cobra.Command, args []string) error

type keyValueStorage interface {
	SetValue(key, value string) error
	GetValue(key string) (string, error)
	GetOrDefaultValue(key, defaultValue string) (string, error)
}

const startHeightKey = "start_height"

type fixBlocksWorker struct {
	baseWorker
}

func (fbw fixBlocksWorker) Name() string {
	return "fix_blocks_worker"
}

func (fbw fixBlocksWorker) Start(ctx context.Context, parseCfg *parse.Config, parseCtx *parse.Context, storage keyValueStorage, interval time.Duration) {
	fbw.baseWorker.Start(ctx, fbw.Name(), fbw.fixBlocks, parseCfg, parseCtx, storage, interval)
}

func (fbw fixBlocksWorker) fixBlocks(parseCfg *parse.Config, parseCtx *parse.Context, storage keyValueStorage) error {
	workerCtx := parser.NewContext(parseCtx.EncodingConfig.Marshaler, nil, parseCtx.Node, parseCtx.Database, parseCtx.Logger, parseCtx.Modules)
	worker := parser.NewWorker(0, workerCtx)

	latestHeight, err := parseCtx.Node.LatestHeight()
	if err != nil {
		return fmt.Errorf("error while getting chain latest block height: %s", err)
	}

	isSynced, err := parseCtx.Database.HasBlock(latestHeight - 10)
	if err != nil {
		return fmt.Errorf("error while checking if synced: %s", err)
	}

	if !isSynced {
		parseCtx.Logger.Info("Not synced - fix blocks worker will skip")
		return nil
	}

	latestHeight-- // This worker should not compete with the main parsing worker

	startHeightVal, err := storage.GetOrDefaultValue(startHeightKey, "0")
	if err != nil {
		return fmt.Errorf("error while getting worker storage key '%s': %s", startHeightKey, err)
	}

	startHeight, err := strconv.ParseInt(startHeightVal, 10, 64)
	if err != nil {
		return fmt.Errorf("error while parsing start height '%s': %s", startHeightVal, err)
	}

	if startHeight == 0 {
		startHeight, err = getGenesisMaxInitialHeight(parseCtx)
		if err != nil {
			return fmt.Errorf("error while getting genesis max initial height: %s", err)
		}
	}

	parseCtx.Logger.Info(fmt.Sprintf("Refetching missing blocks and transactions from height %d... \n", startHeight))

	for ; startHeight <= latestHeight; startHeight++ {
		if err := worker.ProcessIfNotExists(startHeight); err != nil {
			parseCtx.Logger.Error(fmt.Sprintf("Error while re-fetching block %d: %s", startHeight, err))
			break
		}
	}

	latestHeightVal := strconv.FormatInt(latestHeight, 10)
	if err := storage.SetValue(startHeightKey, latestHeightVal); err != nil {
		return fmt.Errorf("error while storing latest height in worker storage '%s': %s", latestHeightVal, err)
	}

	return nil
}

func getGenesisMaxInitialHeight(parseCtx *parse.Context) (int64, error) {
	var rows []types.GenesisRow
	db := database.Cast(parseCtx.Database)
	if err := db.Sqlx.Select(&rows, sqlx.Rebind(sqlx.DOLLAR, "SELECT MAX(initial_height) AS initial_height FROM genesis")); err != nil {
		return 0, err
	}

	if len(rows) == 0 {
		return 0, errors.New("failed to find genesis initial height")
	}

	return rows[0].InitialHeight, nil
}
