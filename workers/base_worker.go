package workers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/database/types"
	parsetypes "github.com/forbole/juno/v3/cmd/parse/types"
	"github.com/forbole/juno/v3/parser"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

type job func(parseCfg *parsetypes.Config, parseCtx *parser.Context, storage keyValueStorage) error

type baseWorker struct{}

func (bw baseWorker) Start(ctx context.Context, name string, j job, parseCfg *parsetypes.Config, parseCtx *parser.Context, storage keyValueStorage, interval time.Duration) {
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

type PersistentPreRunE func(cmd *cobra.Command, args []string) error

type keyValueStorage interface {
	SetValue(key, value string) error
	GetValue(key string) (string, error)
	GetOrDefaultValue(key, defaultValue string) (string, error)
}

func getGenesisMaxInitialHeight(parseCtx *parser.Context) (int64, error) {
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
