package database_test

import (
	"testing"

	"github.com/forbole/bdjuno/v4/database"
	dbtypes "github.com/forbole/bdjuno/v4/database/types"
	utils "github.com/forbole/bdjuno/v4/utils"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

var expectedAppliedMigrations = []dbtypes.Migration{
	{ID: int64(1), Name: "000-initial_schema.sql", CreatedAt: int64(0)},
	{ID: int64(2), Name: "001-workers_storage.sql", CreatedAt: int64(0)},
	{ID: int64(3), Name: "002-inflation_calculation.sql", CreatedAt: int64(0)},
	{ID: int64(4), Name: "003-nft_module.sql", CreatedAt: int64(0)},
	{ID: int64(5), Name: "004-distinct_message_query_func.sql", CreatedAt: int64(0)},
	{ID: int64(6), Name: "005-group_module.sql", CreatedAt: int64(0)},
	{ID: int64(7), Name: "006-marketplace_module.sql", CreatedAt: int64(0)},
	{ID: int64(8), Name: "007-cw20token_module.sql", CreatedAt: int64(0)},
	{ID: int64(9), Name: "008-block_parsed_data.sql", CreatedAt: int64(0)},
	{ID: int64(10), Name: "009-cw20token_update.sql", CreatedAt: int64(0)},
	{ID: int64(11), Name: "010-nft-uniq-id.sql", CreatedAt: int64(0)},
	{ID: int64(12), Name: "011-nft-migrate-uniq-id-values.sql", CreatedAt: int64(0)},
	{ID: int64(13), Name: "012-marketplace-nft-id-column-unique.sql", CreatedAt: int64(0)},
	{ID: int64(14), Name: "013-migrate-to-bdjuno-v4-0-0.sql", CreatedAt: int64(0)},
}

type MigrationTestSuite struct {
	suite.Suite
	db *database.Db
}

func TestGroupModuleTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}

func (suite *MigrationTestSuite) SetupTest() {
	db, err := utils.NewTestDb("migrationTest")
	suite.Require().NoError(err)
	suite.db = db
}

func (suite *MigrationTestSuite) TestExecuteMigrations() {
	var rows []dbtypes.Migration
	suite.Require().NoError(suite.db.Sqlx.Select(&rows, `SELECT id, name FROM migrations`))
	suite.Require().Equal(expectedAppliedMigrations, rows)
}
