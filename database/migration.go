package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"time"

	parsecmd "github.com/forbole/juno/v2/cmd/parse"
	"github.com/jmoiron/sqlx"
)

type Migration struct {
	ID        int64  `db:"id"`
	Name      string `db:"name"`
	CreatedAt int64  `db:"created_at"`
}

const noMigrationsTablePqError = "pq: relation \"migrations\" does not exist"

//go:embed scheme
var scheme embed.FS

func ExecuteMigrations(ctx context.Context, parseCtx *parsecmd.Context) error {
	db := Cast(parseCtx.Database)

	var rows []Migration
	if err := db.Sqlx.SelectContext(ctx, &rows, "SELECT * FROM migrations"); err != nil && err.Error() != noMigrationsTablePqError {
		return err
	}

	appliedMigrations := make(map[string]struct{})
	for _, row := range rows {
		appliedMigrations[row.Name] = struct{}{}
	}

	tx, err := db.Sqlx.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %s", err)
	}

	currentlyExecutedMigrations := []string{}
	commentsRegExp := regexp.MustCompile(`/\*.*\*/`)

	if err := fs.WalkDir(scheme, "scheme", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "scheme" {
			return nil
		}

		_, fileName := filepath.Split(path)
		if _, ok := appliedMigrations[fileName]; ok {
			return nil
		}

		fileContent, err := fs.ReadFile(scheme, path)
		if err != nil {
			return err
		}

		sql := commentsRegExp.ReplaceAllString(string(fileContent), "")

		if _, err = tx.Exec(sql); err != nil {
			return err
		}

		currentlyExecutedMigrations = append(currentlyExecutedMigrations, fileName)

		return nil

	}); err != nil {
		failErr := fmt.Errorf("failed to apply migartions: %s", err)

		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("failed to rollback during: %s", failErr)
		}

		return failErr
	}

	now := time.Now().UnixNano()
	for _, executedMigration := range currentlyExecutedMigrations {
		if _, err := tx.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, "INSERT INTO migrations (name, created_at) VALUES(?, ?);"), executedMigration, now); err != nil {

			failErr := fmt.Errorf("failed to insert executed migration: %s", err)

			if err := tx.Rollback(); err != nil {
				return fmt.Errorf("failed to rollback during: %s", failErr)
			}

			return failErr
		}
	}

	if err := tx.Commit(); err != nil {
		failErr := fmt.Errorf("failed to commit migrations transaction: %s", err)

		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("failed to rollback during: %s", failErr)
		}

		return failErr
	}

	return nil
}
