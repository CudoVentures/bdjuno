package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/forbole/juno/v2/cmd/parse"
	dbconfig "github.com/forbole/juno/v2/database/config"
	"github.com/forbole/juno/v2/logging"

	junodb "github.com/forbole/juno/v2/database"

	"github.com/forbole/bdjuno/v2/database"

	"github.com/cosmos/cosmos-sdk/simapp"
)

func NewTestDb(schema string) (*database.Db, error) {
	dbCfg := dbconfig.NewDatabaseConfig(
		"bdjuno",
		"localhost",
		6433,
		"bdjuno",
		"password",
		"",
		schema,
		-1,
		-1,
	)

	cdc := simapp.MakeTestEncodingConfig()

	db, err := database.Builder(junodb.NewContext(dbCfg, &cdc, logging.DefaultLogger()))
	if err != nil {
		return nil, err
	}

	bigDipperDb, ok := (db).(*database.Db)
	if !ok {
		return nil, errors.New("error while making new test db instance")
	}

	err = bigDipperDb.ExecuteTx(func(dbTx *database.DbTx) error {
		_, err = bigDipperDb.Sql.Exec(fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE;`, schema))
		if err != nil {
			return err
		}

		_, err = bigDipperDb.Sql.Exec(fmt.Sprintf(`CREATE SCHEMA %s;`, schema))
		if err != nil {
			return err
		}

		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
		defer cancelFunc()

		err = database.ExecuteMigrations(ctx, &parse.Context{Database: db})
		if err != nil {
			return err
		}

		return nil
	})

	return bigDipperDb, err
}
