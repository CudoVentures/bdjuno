package database_test

import "github.com/forbole/bdjuno/v4/database"

func (suite *DbTestSuite) TestDbTx_Commit() {
	err := suite.database.ExecuteTx(func(dbTx *database.DbTx) error {
		_, err := dbTx.Exec(`INSERT INTO account VALUES ('1')`)
		suite.Require().NoError(err)

		return nil
	})
	suite.Require().NoError(err)

	var accountsCount int
	err = suite.database.Sqlx.QueryRow(`SELECT COUNT(*) from account`).Scan(&accountsCount)
	suite.Require().NoError(err)
	suite.Require().Equal(1, accountsCount)
}

func (suite *DbTestSuite) TestDbTx_Rollback() {
	err := suite.database.ExecuteTx(func(dbTx *database.DbTx) error {
		_, err := dbTx.Exec(`INSERT INTO account VALUES ('1')`)
		suite.Require().NoError(err)

		_, err = dbTx.Exec(`INSERT INTO account VALUES (null)`)
		return err
	})
	suite.Require().NotNil(err)

	var accountsCount int
	err = suite.database.Sqlx.QueryRow(`SELECT COUNT(*) from account`).Scan(&accountsCount)
	suite.Require().NoError(err)
	suite.Require().Equal(0, accountsCount)
}
