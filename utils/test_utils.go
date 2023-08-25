package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	dbconfig "github.com/forbole/juno/v5/database/config"
	"github.com/forbole/juno/v5/logging"
	"github.com/forbole/juno/v5/parser"

	junodb "github.com/forbole/juno/v5/database"

	"github.com/forbole/bdjuno/v4/database"

	"cosmossdk.io/simapp/params"
)

func NewTestDb(schema string) (*database.Db, error) {
	connStr := fmt.Sprintf("postgresql://bdjuno:password@localhost:6433/bdjuno?sslmode=disable&search_path=%s", schema)
	dbCfg := dbconfig.NewDatabaseConfig(
		connStr,
		"",
		"",
		"",
		"",
		-1,
		-1,
		100000,
		100,
	)

	cdc := params.MakeTestEncodingConfig()

	db, err := database.Builder(junodb.NewContext(dbCfg, &cdc, logging.DefaultLogger()))
	if err != nil {
		return nil, err
	}

	bigDipperDb, ok := (db).(*database.Db)
	if !ok {
		return nil, errors.New("error while making new test db instance")
	}

	err = bigDipperDb.ExecuteTx(func(dbTx *database.DbTx) error {
		_, err = bigDipperDb.SQL.Exec(fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE;`, schema))
		if err != nil {
			return err
		}

		_, err = bigDipperDb.SQL.Exec(fmt.Sprintf(`CREATE SCHEMA %s;`, schema))
		if err != nil {
			return err
		}

		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
		defer cancelFunc()

		err = database.ExecuteMigrations(ctx, &parser.Context{Database: db})
		if err != nil {
			return err
		}

		// Create a default partition for the transaction table
		_, err = bigDipperDb.SQL.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.transaction_default PARTITION OF transaction DEFAULT;`, schema))
		if err != nil {
			return err
		}

		return nil
	})

	return bigDipperDb, err
}
