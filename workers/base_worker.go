package workers

import (
	"context"
	"fmt"
	"time"

	parsecmdtypes "github.com/forbole/juno/v5/cmd/parse/types"
	"github.com/forbole/juno/v5/parser"
)

type job func(parseCfg *parsecmdtypes.Config, parseCtx *parser.Context, storage keyValueStorage) error

type baseWorker struct{}

func (bw baseWorker) Start(ctx context.Context, name string, j job, parseCfg *parsecmdtypes.Config, parseCtx *parser.Context, storage keyValueStorage, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := j(parseCfg, parseCtx, storage); err != nil {
					parseCfg.GetLogger().Error(fmt.Errorf("job from worker '%s' failed: %s", name, err).Error(), name)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
