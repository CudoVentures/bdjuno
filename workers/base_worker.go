package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/forbole/juno/v2/cmd/parse"
)

type job func(parseCfg *parse.Config, parseCtx *parse.Context, storage keyValueStorage) error

type baseWorker struct{}

func (bw baseWorker) Start(ctx context.Context, name string, j job, parseCfg *parse.Config, parseCtx *parse.Context, storage keyValueStorage, interval time.Duration) {
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
