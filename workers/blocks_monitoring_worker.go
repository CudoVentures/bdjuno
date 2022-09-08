package workers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/database/types"
	parsetypes "github.com/forbole/juno/v3/cmd/parse/types"
	"github.com/forbole/juno/v3/parser"
	"github.com/jmoiron/sqlx"
)

type blocksMonitoringWorker struct {
	baseWorker
}

func (bmw blocksMonitoringWorker) Name() string {
	return "blocks_monitoring_worker"
}

func (bmw blocksMonitoringWorker) Start(ctx context.Context, parseCfg *parsetypes.Config, parseCtx *parser.Context, storage keyValueStorage, interval time.Duration) {
	bmw.baseWorker.Start(ctx, bmw.Name(), bmw.monitorBlocks, parseCfg, parseCtx, storage, interval)
}

func (bmw blocksMonitoringWorker) monitorBlocks(parseCfg *parsetypes.Config, parseCtx *parser.Context, storage keyValueStorage) error {

	lastMonitoredBlockHeightVal, err := storage.GetOrDefaultValue(blocksMonitoringLastBlockHeight, "0")
	if err != nil {
		return fmt.Errorf("error while getting worker storage key '%s': %s", blocksMonitoringLastBlockHeight, err)
	}

	lastMonitoredBlockHeight, err := strconv.ParseInt(lastMonitoredBlockHeightVal, 10, 64)
	if err != nil {
		return fmt.Errorf("error while parsing last monitored block height value %s: %s", lastMonitoredBlockHeightVal, err)
	}

	latestStoredBlockHeight, err := getLatestStoredBlockHeight(parseCtx)
	if err != nil {
		return fmt.Errorf("error while getting latest stored block height: %s", err)
	}

	currentTime := time.Now().UnixNano() / 1000

	if latestStoredBlockHeight > lastMonitoredBlockHeight {

		lastMonitoredBlockHeightVal = strconv.FormatInt(latestStoredBlockHeight, 10)
		if err := storage.SetValue(blocksMonitoringLastBlockHeight, lastMonitoredBlockHeightVal); err != nil {
			return fmt.Errorf("error while storing last monitored block height value '%s': %s", lastMonitoredBlockHeightVal, err)
		}

		lastMonitoredBlockTimeVal := strconv.FormatInt(currentTime, 10)
		if err := storage.SetValue(blocksMonitoringLastBlockTime, lastMonitoredBlockTimeVal); err != nil {
			return fmt.Errorf("error while storing last monitored block time value '%s': %s", lastMonitoredBlockTimeVal, err)
		}

		return nil
	}

	lastMonitoredBlockTimeVal, err := storage.GetValue(blocksMonitoringLastBlockTime)
	if err != nil {
		return fmt.Errorf("error while getting last monitored block time value '%s': %s", blocksMonitoringLastBlockTime, err)
	}

	lastMonitoredBlockTime, err := strconv.ParseInt(lastMonitoredBlockTimeVal, 10, 64)
	if err != nil {
		return fmt.Errorf("error while parsing last monitored block time value %s: %s", lastMonitoredBlockTimeVal, err)
	}

	if currentTime-lastMonitoredBlockTime > 60 {
		os.Exit(-1)
	}

	return nil
}

func getLatestStoredBlockHeight(parseCtx *parser.Context) (int64, error) {
	var rows []types.BlockRow
	db := database.Cast(parseCtx.Database)
	if err := db.Sqlx.Select(&rows, sqlx.Rebind(sqlx.DOLLAR, "SELECT MAX(height) AS height FROM block")); err != nil {
		return 0, err
	}

	if len(rows) == 0 {
		return 0, errors.New("failed to find latest block height")
	}

	return rows[0].Height, nil
}

const (
	blocksMonitoringLastBlockHeight = "blocks_monitoring_last_block_height"
	blocksMonitoringLastBlockTime   = "blocks_monitoring_last_block_time"
)
