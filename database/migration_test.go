package database_test

import (
	"github.com/forbole/bdjuno/v2/database"
	_ "github.com/lib/pq"
	_ "github.com/proullon/ramsql/driver"
)

func (suite *DbTestSuite) TestExecuteMigrations() {
	var rows []database.Migration
	suite.Require().NoError(suite.database.Sqlx.Select(&rows, `SELECT * FROM migrations`))
	suite.Require().Equal(int64(1), rows[0].ID)
	suite.Require().Equal("000-initial_schema.sql", rows[0].Name)
}
