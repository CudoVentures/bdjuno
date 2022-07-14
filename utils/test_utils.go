package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/juno/v2/cmd/parse"
	dbconfig "github.com/forbole/juno/v2/database/config"
	"github.com/forbole/juno/v2/logging"

	junodb "github.com/forbole/juno/v2/database"

	"github.com/forbole/bdjuno/v2/database"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/stretchr/testify/suite"
)

func NewTestDb(suite *suite.Suite, schema string) (*database.Db, codec.Codec) {
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

	db, err := database.Builder(
		junodb.NewContext(dbCfg, &cdc, logging.DefaultLogger()),
	)
	suite.Require().NoError(err)

	bigDipperDb, ok := (db).(*database.Db)
	suite.Require().True(ok)

	_, err = bigDipperDb.Sql.Exec(fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE;`, schema))
	suite.Require().NoError(err)

	_, err = bigDipperDb.Sql.Exec(fmt.Sprintf(`CREATE SCHEMA %s;`, schema))
	suite.Require().NoError(err)

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelFunc()

	err = database.ExecuteMigrations(ctx, &parse.Context{Database: db})
	suite.Require().NoError(err)

	return bigDipperDb, cdc.Marshaler
}
