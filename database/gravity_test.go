package database_test

import dbtypes "github.com/forbole/bdjuno/v2/database/types"

const (
	testOrchestrator1 = "cudos1a326k254fukx9jlp0h3fwcr2ymjgludzum67d1"
	testOrchestrator2 = "cudos1a326k254fukx9jlp0h3fwcr2ymjgludzum67d2"
)

func (suite *DbTestSuite) TestGravity() {
	testGravity_GetOrchestratorsCount(suite, 0)
	testGravity_SaveOrchestrator(suite)
	testGravity_GetOrchestratorsCount(suite, 2)
	testGravity_SaveMsgSendToCosmosClaim(suite)
	testGravity_GetGravityTransactionVotes(suite)
	testGravity_SetGravityTransactionConsensus(suite)
}

func testGravity_SaveOrchestrator(suite *DbTestSuite) {
	err := suite.database.SaveOrchestrator(testOrchestrator1)
	suite.Require().NoError(err)
	err = suite.database.SaveOrchestrator(testOrchestrator1)
	suite.Require().NoError(err)

	err = suite.database.SaveOrchestrator(testOrchestrator2)
	suite.Require().NoError(err)

	var rows []dbtypes.GravityOrchestratorRow
	err = suite.database.Sqlx.Select(&rows, "SELECT * FROM gravity_orchestrator ORDER BY address ASC")
	suite.Require().NoError(err)

	suite.Require().Equal([]dbtypes.GravityOrchestratorRow{
		{
			Address: testOrchestrator1,
		},
		{
			Address: testOrchestrator2,
		},
	}, rows)
}

func testGravity_GetOrchestratorsCount(suite *DbTestSuite, expectedValue int) {
	count, err := suite.database.GetOrchestratorsCount()
	suite.Require().NoError(err)
	suite.Require().Equal(expectedValue, count)
}

func testGravity_SaveMsgSendToCosmosClaim(suite *DbTestSuite) {
	txHash := "txhash#31337"
	insertDummyTransaction(suite, txHash)

	err := suite.database.SaveMsgSendToCosmosClaim(txHash, "SendToCosmosClaim", "1", "me", "you")
	suite.Require().NotNil(err)

	err = suite.database.SaveMsgSendToCosmosClaim(txHash, "SendToCosmosClaim", "1", "me", testOrchestrator1)
	suite.Require().Nil(err)
	err = suite.database.SaveMsgSendToCosmosClaim(txHash, "SendToCosmosClaim", "1", "me", testOrchestrator1)
	suite.Require().Nil(err)

	err = suite.database.SaveMsgSendToCosmosClaim(txHash, "SendToCosmosClaim", "2", "me", testOrchestrator1)
	suite.Require().Nil(err)
}

func testGravity_GetGravityTransactionVotes(suite *DbTestSuite) {
	votes, err := suite.database.GetGravityTransactionVotes("1")
	suite.Require().Nil(err)
	suite.Require().Equal(2, votes)

	votes, err = suite.database.GetGravityTransactionVotes("2")
	suite.Require().Nil(err)
	suite.Require().Equal(1, votes)
}

func testGravity_SetGravityTransactionConsensus(suite *DbTestSuite) {
	err := suite.database.SetGravityTransactionConsensus("1", true)
	suite.Require().Nil(err)

	var rows []dbtypes.GravityTransactionRow
	err = suite.database.Sqlx.Select(&rows, "SELECT * FROM gravity_transaction ORDER BY attestation_id ASC")
	suite.Require().Nil(err)
	suite.Require().Equal("1", rows[0].AttestationID)
	suite.Require().Equal(true, rows[0].Consensus)
	suite.Require().Equal("2", rows[1].AttestationID)
	suite.Require().Equal(false, rows[1].Consensus)
}
