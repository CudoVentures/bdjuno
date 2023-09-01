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
	{ID: int64(1), Name: "00-initial_schema.sql", CreatedAt: int64(0)},
	{ID: int64(2), Name: "01-cosmos.sql", CreatedAt: int64(0)},
	{ID: int64(3), Name: "02-auth.sql", CreatedAt: int64(0)},
	{ID: int64(4), Name: "03-bank.sql", CreatedAt: int64(0)},
	{ID: int64(5), Name: "04-staking.sql", CreatedAt: int64(0)},
	{ID: int64(6), Name: "05-consensus.sql", CreatedAt: int64(0)},
	{ID: int64(7), Name: "06-mint.sql", CreatedAt: int64(0)},
	{ID: int64(8), Name: "07-distribution.sql", CreatedAt: int64(0)},
	{ID: int64(9), Name: "08-pricefeed.sql", CreatedAt: int64(0)},
	{ID: int64(10), Name: "09-gov.sql", CreatedAt: int64(0)},
	{ID: int64(11), Name: "10-modules.sql", CreatedAt: int64(0)},
	{ID: int64(12), Name: "11-slashing.sql", CreatedAt: int64(0)},
	{ID: int64(13), Name: "12-feegrant.sql", CreatedAt: int64(0)},
	{ID: int64(14), Name: "13-upgrade.sql", CreatedAt: int64(0)},
	{ID: int64(15), Name: "14-cosmwasm.sql", CreatedAt: int64(0)},
	{ID: int64(16), Name: "15-gravity.sql", CreatedAt: int64(0)},
	{ID: int64(17), Name: "16-workers_storage.sql", CreatedAt: int64(0)},
	{ID: int64(18), Name: "17-inflation_calculation.sql", CreatedAt: int64(0)},
	{ID: int64(19), Name: "18-nft_module.sql", CreatedAt: int64(0)},
	{ID: int64(20), Name: "19-distinct_message_query_func.sql", CreatedAt: int64(0)},
	{ID: int64(21), Name: "20-group_module.sql", CreatedAt: int64(0)},
	{ID: int64(22), Name: "21-marketplace_module.sql", CreatedAt: int64(0)},
	{ID: int64(23), Name: "22-cw20token_module.sql", CreatedAt: int64(0)},
	{ID: int64(24), Name: "23-block_parsed_data.sql", CreatedAt: int64(0)},
	{ID: int64(25), Name: "24-cw20token_update.sql", CreatedAt: int64(0)},
	{ID: int64(26), Name: "25-nft-uniq-id.sql", CreatedAt: int64(0)},
	{ID: int64(27), Name: "26-nft-migrate-uniq-id-values.sql", CreatedAt: int64(0)},
	{ID: int64(28), Name: "27-marketplace-nft-id-column-unique.sql", CreatedAt: int64(0)},
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
