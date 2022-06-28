package database_test

import (
	"github.com/forbole/bdjuno/v2/database"
	_ "github.com/lib/pq"
)

var expectedAppliedMigrations = []database.Migration{
	{ID: int64(1), Name: "000-initial_schema.sql", CreatedAt: int64(0)},
	{ID: int64(2), Name: "001-workers_storage.sql", CreatedAt: int64(0)},
	{ID: int64(3), Name: "002-inflation_calculation.sql", CreatedAt: int64(0)},
}

func (suite *DbTestSuite) TestExecuteMigrations() {
	var rows []database.Migration
	suite.Require().NoError(suite.database.Sqlx.Select(&rows, `SELECT id, name FROM migrations`))
	suite.Require().Equal(expectedAppliedMigrations, rows)
}