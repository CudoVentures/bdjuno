package database_test

import (
	"time"

	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/types"
	"github.com/lib/pq"
)

func (suite *DbTestSuite) TestCosmWasm_SaveMsgStoreCodeData() {
	txHash := "txhash#1"
	insertDummyTransaction(suite, txHash)
	msgIndex := 0
	sender := "cudos1a326k254fukx9jlp0h3fwcr2ymjgludzum67dv"
	instantiatePermission := "{}"
	resultCodeID := "1"
	success := true

	err := suite.database.SaveMsgStoreCodeData(
		types.NewMsgStoreCodeData(
			txHash,
			sender,
			msgIndex,
			success,
			instantiatePermission,
			resultCodeID,
		),
	)
	suite.Require().NoError(err)

	var rows []dbtypes.CosmwasmStoreRow
	err = suite.database.Sqlx.Select(&rows, "SELECT * FROM cosmwasm_store")
	suite.Require().NoError(err)

	suite.Require().Equal(1, len(rows))
	suite.Require().Equal(dbtypes.CosmwasmStoreRow{
		TransactionHash:       txHash,
		Index:                 msgIndex,
		Sender:                sender,
		InstantiatePermission: instantiatePermission,
		ResultCodeID:          resultCodeID,
		Success:               success,
	}, rows[0])

	success = false
	resultCodeID = "2"

	err = suite.database.SaveMsgStoreCodeData(
		types.NewMsgStoreCodeData(
			txHash,
			sender,
			msgIndex,
			success,
			instantiatePermission,
			resultCodeID,
		),
	)
	suite.Require().NoError(err)

	rows = []dbtypes.CosmwasmStoreRow{}
	err = suite.database.Sqlx.Select(&rows, "SELECT * FROM cosmwasm_store")
	suite.Require().NoError(err)

	suite.Require().Equal(1, len(rows))
	suite.Require().Equal(dbtypes.CosmwasmStoreRow{
		TransactionHash:       txHash,
		Index:                 msgIndex,
		Sender:                sender,
		InstantiatePermission: instantiatePermission,
		ResultCodeID:          resultCodeID,
		Success:               success,
	}, rows[0])

	err = suite.database.Sqlx.Select(&rows, "DELETE FROM cosmwasm_store")
	suite.Require().NoError(err)

	deleteDummyTransaction(suite, txHash)
}

func (suite *DbTestSuite) TestCosmWasm_SaveMsgInstantiateContractData() {
	txHash := "txhash#1"
	insertDummyTransaction(suite, txHash)
	msgIndex := 0
	sender := "cudos1a326k254fukx9jlp0h3fwcr2ymjgludzum67dv"
	success := true
	admin := "cudos1a326k254fukx9jlp0h3fwcr2ymjgludzum67dv"
	funds := "[]"
	label := "broken code"
	codeID := "1"
	resultContractAddress := "cudos1glw29zgkz4uh3u7amkunkhxvnkrwd20dvvlerk"

	err := suite.database.SaveMsgInstantiateContractData(
		types.NewMsgInstantiateContractData(
			txHash,
			sender,
			msgIndex,
			success,
			admin,
			funds,
			label,
			codeID,
			resultContractAddress,
		),
	)
	suite.Require().NoError(err)

	rows := []dbtypes.CosmwasmInstantiateRow{}
	err = suite.database.Sqlx.Select(&rows, "SELECT * FROM cosmwasm_instantiate")
	suite.Require().NoError(err)

	suite.Require().Equal(1, len(rows))
	suite.Require().Equal(dbtypes.CosmwasmInstantiateRow{
		TransactionHash:       txHash,
		Index:                 msgIndex,
		Admin:                 admin,
		Funds:                 funds,
		Label:                 label,
		Sender:                sender,
		CodeID:                codeID,
		ResultContractAddress: resultContractAddress,
		Success:               success,
	}, rows[0])

	admin = ""
	label = "very broken code"
	codeID = "3"

	err = suite.database.SaveMsgInstantiateContractData(
		types.NewMsgInstantiateContractData(
			txHash,
			sender,
			msgIndex,
			success,
			admin,
			funds,
			label,
			codeID,
			resultContractAddress,
		),
	)
	suite.Require().NoError(err)

	rows = []dbtypes.CosmwasmInstantiateRow{}
	err = suite.database.Sqlx.Select(&rows, "SELECT * FROM cosmwasm_instantiate")
	suite.Require().NoError(err)

	suite.Require().Equal(1, len(rows))
	suite.Require().Equal(dbtypes.CosmwasmInstantiateRow{
		TransactionHash:       txHash,
		Index:                 msgIndex,
		Admin:                 admin,
		Funds:                 funds,
		Label:                 label,
		Sender:                sender,
		CodeID:                codeID,
		ResultContractAddress: resultContractAddress,
		Success:               success,
	}, rows[0])

	err = suite.database.Sqlx.Select(&rows, "DELETE FROM cosmwasm_instantiate")
	suite.Require().NoError(err)

	deleteDummyTransaction(suite, txHash)
}

func (suite *DbTestSuite) TestCosmWasm_SaveMsgExecuteContractData() {
	txHash := "txhash#1"
	insertDummyTransaction(suite, txHash)
	msgIndex := 0
	sender := "cudos1a326k254fukx9jlp0h3fwcr2ymjgludzum67dv"
	success := true
	method := "issue_denom"
	arguments := "{}"
	funds := "[]"
	contract := "cudos1glw29zgkz4uh3u7amkunkhxvnkrwd20dvvlerk"

	err := suite.database.SaveMsgExecuteContractData(
		types.NewMsgExecuteContractData(
			txHash,
			sender,
			msgIndex,
			success,
			method,
			arguments,
			funds,
			contract,
		),
	)
	suite.Require().NoError(err)

	rows := []dbtypes.CosmwasmExecuteRow{}
	err = suite.database.Sqlx.Select(&rows, "SELECT * FROM cosmwasm_execute")
	suite.Require().NoError(err)

	suite.Require().Equal(1, len(rows))
	suite.Require().Equal(dbtypes.CosmwasmExecuteRow{
		TransactionHash: txHash,
		Index:           msgIndex,
		Method:          method,
		Arguments:       arguments,
		Funds:           funds,
		Sender:          sender,
		Contract:        contract,
		Success:         success,
	}, rows[0])

	method = "delete_denom"
	success = false

	err = suite.database.SaveMsgExecuteContractData(
		types.NewMsgExecuteContractData(
			txHash,
			sender,
			msgIndex,
			success,
			method,
			arguments,
			funds,
			contract,
		),
	)
	suite.Require().NoError(err)

	rows = []dbtypes.CosmwasmExecuteRow{}
	err = suite.database.Sqlx.Select(&rows, "SELECT * FROM cosmwasm_execute")
	suite.Require().NoError(err)

	suite.Require().Equal(1, len(rows))
	suite.Require().Equal(dbtypes.CosmwasmExecuteRow{
		TransactionHash: txHash,
		Index:           msgIndex,
		Method:          method,
		Arguments:       arguments,
		Funds:           funds,
		Sender:          sender,
		Contract:        contract,
		Success:         success,
	}, rows[0])

	err = suite.database.Sqlx.Select(&rows, "DELETE FROM cosmwasm_execute")
	suite.Require().NoError(err)

	deleteDummyTransaction(suite, txHash)
}

func (suite *DbTestSuite) TestCosmWasm_SaveMsgMigrateContactData() {
	txHash := "txhash#1"
	insertDummyTransaction(suite, txHash)
	msgIndex := 0
	sender := "cudos1a326k254fukx9jlp0h3fwcr2ymjgludzum67dv"
	success := true
	contract := "cudos1glw29zgkz4uh3u7amkunkhxvnkrwd20dvvlerk"
	codeID := "100"
	arguments := "{}"

	err := suite.database.SaveMsgMigrateContactData(
		types.NewMsgMigrateContractData(
			txHash,
			sender,
			msgIndex,
			success,
			contract,
			codeID,
			arguments,
		),
	)
	suite.Require().NoError(err)

	rows := []dbtypes.CosmwasmMigrateRow{}
	err = suite.database.Sqlx.Select(&rows, "SELECT * FROM cosmwasm_migrate")
	suite.Require().NoError(err)

	suite.Require().Equal(1, len(rows))
	suite.Require().Equal(dbtypes.CosmwasmMigrateRow{
		TransactionHash: txHash,
		Index:           msgIndex,
		Sender:          sender,
		Contract:        contract,
		CodeID:          codeID,
		Arguments:       arguments,
		Success:         success,
	}, rows[0])

	codeID = "9999999"

	err = suite.database.SaveMsgMigrateContactData(
		types.NewMsgMigrateContractData(
			txHash,
			sender,
			msgIndex,
			success,
			contract,
			codeID,
			arguments,
		),
	)
	suite.Require().NoError(err)

	rows = []dbtypes.CosmwasmMigrateRow{}
	err = suite.database.Sqlx.Select(&rows, "SELECT * FROM cosmwasm_migrate")
	suite.Require().NoError(err)

	suite.Require().Equal(1, len(rows))
	suite.Require().Equal(dbtypes.CosmwasmMigrateRow{
		TransactionHash: txHash,
		Index:           msgIndex,
		Sender:          sender,
		Contract:        contract,
		CodeID:          codeID,
		Arguments:       arguments,
		Success:         success,
	}, rows[0])

	err = suite.database.Sqlx.Select(&rows, "DELETE FROM cosmwasm_migrate")
	suite.Require().NoError(err)

	deleteDummyTransaction(suite, txHash)
}

func (suite *DbTestSuite) TestCosmWasm_SaveMsgUpdateAdminData() {
	txHash := "txhash#1"
	insertDummyTransaction(suite, txHash)
	msgIndex := 0
	sender := "cudos1a326k254fukx9jlp0h3fwcr2ymjgludzum67dv"
	success := true
	contract := "cudos1glw29zgkz4uh3u7amkunkhxvnkrwd20dvvlerk"
	newAdmin := "cudos1glw29zgkz4uh3u7amkunkhxvnkrwd20dvv12da"

	err := suite.database.SaveMsgUpdateAdminData(
		types.NewMsgUpdateAdminData(
			txHash,
			sender,
			msgIndex,
			success,
			contract,
			newAdmin,
		),
	)
	suite.Require().NoError(err)

	rows := []dbtypes.CosmwasmUpdateAdminRow{}
	err = suite.database.Sqlx.Select(&rows, "SELECT * FROM cosmwasm_update_admin")
	suite.Require().NoError(err)

	suite.Require().Equal(1, len(rows))
	suite.Require().Equal(dbtypes.CosmwasmUpdateAdminRow{
		TransactionHash: txHash,
		Index:           msgIndex,
		Sender:          sender,
		Contract:        contract,
		NewAdmin:        newAdmin,
		Success:         success,
	}, rows[0])

	contract = "cudos1glw29zgkz4uh3u7amkunkhxvnkrwd20dvvl4cb"
	newAdmin = "cudos1glw29zgkz4uh3u7amkunkhxvnkrwd20dvv19fe"

	err = suite.database.SaveMsgUpdateAdminData(
		types.NewMsgUpdateAdminData(
			txHash,
			sender,
			msgIndex,
			success,
			contract,
			newAdmin,
		),
	)
	suite.Require().NoError(err)

	rows = []dbtypes.CosmwasmUpdateAdminRow{}
	err = suite.database.Sqlx.Select(&rows, "SELECT * FROM cosmwasm_update_admin")
	suite.Require().NoError(err)

	suite.Require().Equal(1, len(rows))
	suite.Require().Equal(dbtypes.CosmwasmUpdateAdminRow{
		TransactionHash: txHash,
		Index:           msgIndex,
		Sender:          sender,
		Contract:        contract,
		NewAdmin:        newAdmin,
		Success:         success,
	}, rows[0])

	err = suite.database.Sqlx.Select(&rows, "DELETE FROM cosmwasm_update_admin")
	suite.Require().NoError(err)

	deleteDummyTransaction(suite, txHash)
}

func (suite *DbTestSuite) TestCosmWasm_SaveMsgClearAdminData() {
	txHash := "txhash#1"
	insertDummyTransaction(suite, txHash)
	msgIndex := 0
	sender := "cudos1a326k254fukx9jlp0h3fwcr2ymjgludzum67dv"
	success := true
	contract := "cudos1glw29zgkz4uh3u7amkunkhxvnkrwd20dvvlerk"

	err := suite.database.SaveMsgClearAdminData(
		types.NewClearAdminData(
			txHash,
			sender,
			msgIndex,
			success,
			contract,
		),
	)
	suite.Require().NoError(err)

	rows := []dbtypes.CosmwasmClearAdminRow{}
	err = suite.database.Sqlx.Select(&rows, "SELECT * FROM cosmwasm_clear_admin")
	suite.Require().NoError(err)

	suite.Require().Equal(1, len(rows))
	suite.Require().Equal(dbtypes.CosmwasmClearAdminRow{
		TransactionHash: txHash,
		Index:           msgIndex,
		Sender:          sender,
		Contract:        contract,
		Success:         success,
	}, rows[0])

	sender = "cudos1a326k254fukx9jlp0h3fwcr2ymjgludzum12ba"

	err = suite.database.SaveMsgClearAdminData(
		types.NewClearAdminData(
			txHash,
			sender,
			msgIndex,
			success,
			contract,
		),
	)
	suite.Require().NoError(err)

	rows = []dbtypes.CosmwasmClearAdminRow{}
	err = suite.database.Sqlx.Select(&rows, "SELECT * FROM cosmwasm_clear_admin")
	suite.Require().NoError(err)

	suite.Require().Equal(1, len(rows))
	suite.Require().Equal(dbtypes.CosmwasmClearAdminRow{
		TransactionHash: txHash,
		Index:           msgIndex,
		Sender:          sender,
		Contract:        contract,
		Success:         success,
	}, rows[0])

	err = suite.database.Sqlx.Select(&rows, "DELETE FROM cosmwasm_clear_admin")
	suite.Require().NoError(err)

	deleteDummyTransaction(suite, txHash)
}

func insertDummyTransaction(suite *DbTestSuite, txHash string) {
	height := 1
	insertDummyBlock(suite, height)
	success := true
	messages := "[]"
	memo := ""
	signatures := []string{"sign1"}
	signerInfos := "[]"
	fee := "{}"
	_, err := suite.database.Sqlx.Exec(`INSERT INTO transaction (hash, height, success, messages, memo, signatures, signer_infos, fee) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8
	) ON CONFLICT DO NOTHING`, txHash, height, success, messages, memo, pq.Array(signatures), signerInfos, fee)
	suite.Require().NoError(err)
}

func insertDummyBlock(suite *DbTestSuite, height int) {
	hash := "some hash"
	now := time.Now()
	_, err := suite.database.Sqlx.Exec(`INSERT INTO block (height, hash, timestamp) VALUES (
		$1, $2, $3) ON CONFLICT DO NOTHING`, height, hash, &now)
	suite.Require().NoError(err)
}

func deleteDummyTransaction(suite *DbTestSuite, txHash string) {
	_, err := suite.database.Sqlx.Exec(`DELETE FROM transaction WHERE hash = $1`, txHash)
	suite.Require().NoError(err)
	deleteDummyBlock(suite, 1)
}

func deleteDummyBlock(suite *DbTestSuite, height int) {
	_, err := suite.database.Sqlx.Exec(`DELETE FROM block WHERE height = $1`, height)
	suite.Require().NoError(err)
}
