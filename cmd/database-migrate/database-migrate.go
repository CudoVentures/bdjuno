package migratedb

import (
	"context"
	"fmt"
	"time"

	"github.com/forbole/bdjuno/v4/database"
	parsecmdtypes "github.com/forbole/juno/v5/cmd/parse/types"
	"github.com/forbole/juno/v5/types/config"
	"github.com/spf13/cobra"
)

// NewDatabaseMigrateCmd returns the Cobra command allowing to migrate the db up to latest scheme
func NewDatabaseMigrateCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "database migrate",
		Short:   "Migrates database to latest scheme from database/scheme folder",
		Example: "bdjuno database migrate",
		PreRunE: parsecmdtypes.ReadConfigPreRunE(parseConfig),
		RunE: func(cmd *cobra.Command, args []string) error {
			parseCtx, err := parsecmdtypes.GetParserContext(config.Cfg, parseConfig)
			if err != nil {
				return err
			}

			ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*120)
			defer cancelFunc()
			if err := database.ExecuteMigrations(ctx, parseCtx); err != nil {
				return fmt.Errorf("failed to execute migrations: %s", err)
			}

			return nil
		},
	}
}
