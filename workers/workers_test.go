package workers

import (
	"context"
	"testing"
	"time"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/juno/v2/cmd/parse"
	"github.com/stretchr/testify/require"
)

var instancesCount int

type mockWorker struct{}

func (mw mockWorker) Name() string {
	return "mockWorker"
}

func (mw mockWorker) Start(ctx context.Context, parseCfg *parse.Config, parseCtx *parse.Context,
	storage keyValueStorage, interval time.Duration) {
	instancesCount++
}

func TestStartWorkers(t *testing.T) {
	cfg := workersConfig{
		Workers: []workerConfig{
			{
				Name:     "mockWorker",
				Interval: "1s",
			},
		},
	}
	var parseCfg parse.Config
	parseCtx := parse.Context{
		Database: &database.Db{},
	}

	workers := []worker{
		mockWorker{},
		mockWorker{},
		mockWorker{},
	}

	err := startWorkers(context.Background(), workers, cfg, &parseCfg, &parseCtx)

	require.NoError(t, err)
	require.Equal(t, len(workers), instancesCount)
}

func TestBaseWorkerStart(t *testing.T) {
	var executionsCount int
	job := func(parseCfg *parse.Config, parseCtx *parse.Context, storage keyValueStorage) error {
		executionsCount += 1
		return nil
	}
	bw := baseWorker{}
	bw.Start(context.Background(), "test", job, nil, nil, nil, 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	require.Greater(t, executionsCount, 2)
}
