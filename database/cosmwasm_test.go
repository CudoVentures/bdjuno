package database_test

import (
	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/types"
)

func (suite *DbTestSuite) TestCosmWasm_SaveMsgStoreCodeData() {
	txHash := "hash#1"
	msgIndex := 0
	sender := "cudos1a326k254fukx9jlp0h3fwcr2ymjgludzum67dv"
	instantiatePermission := "{}"
	resultCodeId := "1"
	success := true

	err := suite.database.SaveMsgStoreCodeData(
		types.NewMsgStoreCodeData(
			txHash,
			sender,
			msgIndex,
			success,
			instantiatePermission,
			resultCodeId,
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
		ResultCodeId:          resultCodeId,
		Success:               success,
	}, rows[0])

	success = false
	resultCodeId = "2"

	err = suite.database.SaveMsgStoreCodeData(
		types.NewMsgStoreCodeData(
			txHash,
			sender,
			msgIndex,
			success,
			instantiatePermission,
			resultCodeId,
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
		ResultCodeId:          resultCodeId,
		Success:               success,
	}, rows[0])
}

func (suite *DbTestSuite) TestCosmWasm_SaveMsgInstantiateContractData() {
	txHash := "hash#1"
	msgIndex := 0
	sender := "cudos1a326k254fukx9jlp0h3fwcr2ymjgludzum67dv"
	success := true
	admin := "cudos1a326k254fukx9jlp0h3fwcr2ymjgludzum67dv"
	funds := "[]"
	label := "broken code"
	codeId := "1"
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
			codeId,
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
		CodeId:                codeId,
		ResultContractAddress: resultContractAddress,
		Success:               success,
	}, rows[0])

	admin = ""
	label = "very broken code"
	codeId = "3"

	err = suite.database.SaveMsgInstantiateContractData(
		types.NewMsgInstantiateContractData(
			txHash,
			sender,
			msgIndex,
			success,
			admin,
			funds,
			label,
			codeId,
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
		CodeId:                codeId,
		ResultContractAddress: resultContractAddress,
		Success:               success,
	}, rows[0])
}

func (suite *DbTestSuite) TestCosmWasm_SaveMsgExecuteContractData() {
	txHash := "hash#1"
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
}

func (suite *DbTestSuite) TestCosmWasm_SaveMsgMigrateContactData() {
	txHash := "hash#1"
	msgIndex := 0
	sender := "cudos1a326k254fukx9jlp0h3fwcr2ymjgludzum67dv"
	success := true
	contract := "cudos1glw29zgkz4uh3u7amkunkhxvnkrwd20dvvlerk"
	codeId := "100"
	arguments := "{}"

	err := suite.database.SaveMsgMigrateContactData(
		types.NewMsgMigrateContractData(
			txHash,
			sender,
			msgIndex,
			success,
			contract,
			codeId,
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
		CodeId:          codeId,
		Arguments:       arguments,
		Success:         success,
	}, rows[0])

	codeId = "9999999"

	err = suite.database.SaveMsgMigrateContactData(
		types.NewMsgMigrateContractData(
			txHash,
			sender,
			msgIndex,
			success,
			contract,
			codeId,
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
		CodeId:          codeId,
		Arguments:       arguments,
		Success:         success,
	}, rows[0])
}

func (suite *DbTestSuite) TestCosmWasm_SaveMsgUpdateAdminData() {
	txHash := "hash#1"
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
}

func (suite *DbTestSuite) TestCosmWasm_SaveMsgClearAdminData() {
	txHash := "hash#1"
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
}
