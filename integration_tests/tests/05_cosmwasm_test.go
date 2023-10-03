package test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	node "github.com/CudoVentures/cudos-node/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbTypes "github.com/forbole/bdjuno/v4/database/types"
	config "github.com/forbole/bdjuno/v4/integration_tests/set_up"
	"github.com/stretchr/testify/require"
)

const (
	wasmModule                 = "wasm"
	expectedWasmContractCodeID = "1"
)

var (
	withCoinsFlag   = config.GetFlag("amount", smallDepositAmount)
	withAdminFlag   = config.GetFlag("admin", CudosAdmin)
	contractAddress string
)

func TestWasmStoreCode(t *testing.T) {

	// PREPARE
	testContractPath := filepath.Join("..", "set_up", "test_contract.wasm")
	args := []string{
		wasmModule,
		"store",
		testContractPath,
	}

	// EXECUTE
	result, err := config.ExecuteTxCommandWithGas(User1, args...)
	require.NoError(t, err)

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	// cosmwasm_store
	var resultFromDB dbTypes.CosmwasmStoreRow
	err = config.QueryDatabase(`
	SELECT 
		sender, 
		result_code_id, 
		success FROM cosmwasm_store
		WHERE transaction_hash = $1`, txHash).Scan(
		&resultFromDB.Sender,
		&resultFromDB.ResultCodeID,
		&resultFromDB.Success,
	)
	require.NoError(t, err)

	require.Equal(t, expectedWasmContractCodeID, resultFromDB.ResultCodeID)
	require.Equal(t, User1, resultFromDB.Sender)
	require.Equal(t, true, resultFromDB.Success)
}

func TestWasmInstantiateContract(t *testing.T) {

	// PREPARE
	instantiator := User1
	label := "test_contract"
	args := []string{
		wasmModule,
		"instantiate",
		expectedWasmContractCodeID,
		"{}",
		config.GetFlag("label", label),
		withCoinsFlag,
		withAdminFlag,
	}

	// EXECUTE
	result, err := config.ExecuteTxCommandWithFees(instantiator, args...)
	require.NoError(t, err)

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	// cosmwasm_instantiate
	var resultFromDB dbTypes.CosmwasmInstantiateRow
	err = config.QueryDatabase(`
	SELECT 
		admin, 
		funds,
		label,
		sender,
		result_contract_address FROM cosmwasm_instantiate
		WHERE transaction_hash = $1 
		AND code_id = $2 
		AND success = $3`, txHash, expectedWasmContractCodeID, true).Scan(
		&resultFromDB.Admin,
		&resultFromDB.Funds,
		&resultFromDB.Label,
		&resultFromDB.Sender,
		&resultFromDB.ResultContractAddress,
	)
	require.NoError(t, err)

	contractAddress = resultFromDB.ResultContractAddress
	require.NotEmpty(t, contractAddress)
	bz, err := sdk.GetFromBech32(contractAddress, node.AccountAddressPrefix)
	require.NoError(t, err)
	err = sdk.VerifyAddressFormat(bz)
	require.NoError(t, err)

	require.Equal(t, CudosAdmin, resultFromDB.Admin)
	require.Equal(t, instantiator, resultFromDB.Sender)
	require.Equal(t, label, resultFromDB.Label)

	var contractFunds sdk.Coins
	err = json.Unmarshal([]byte(resultFromDB.Funds), &contractFunds)
	require.NoError(t, err)
	require.Len(t, contractFunds, 1)
	require.Equal(t, smallDepositAmount, contractFunds[0].String())
}

func TestWasmUpdateAdmin(t *testing.T) {

	// PREPARE
	// Depending on the previous test success
	require.NotEmpty(t, contractAddress)
	intendedNewAdmin := User1
	var currentAdmin string
	err := config.QueryDatabase(`
	SELECT 
		admin FROM cosmwasm_instantiate
		WHERE result_contract_address = $1`, contractAddress).Scan(&currentAdmin)
	require.NoError(t, err)
	require.NotEqual(t, intendedNewAdmin, currentAdmin)

	args := []string{
		wasmModule,
		"set-contract-admin",
		contractAddress,
		intendedNewAdmin,
	}

	// EXECUTE
	result, err := config.ExecuteTxCommandWithFees(currentAdmin, args...)
	require.NoError(t, err)

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	err = config.QueryDatabase(`
	SELECT 
		new_admin FROM cosmwasm_update_admin
		WHERE contract = $1 
		AND transaction_hash = $2`, contractAddress, txHash).Scan(&currentAdmin)
	require.NoError(t, err)
	require.Equal(t, intendedNewAdmin, currentAdmin)
}

func TestWasmMigrateContract(t *testing.T) {

	// PREPARE
	// Depending on the previous test success
	require.NotEmpty(t, contractAddress)
	currentAdmin := User1
	args := []string{
		wasmModule,
		"migrate",
		contractAddress,
		expectedWasmContractCodeID,
		"{}",
	}

	// EXECUTE
	result, err := config.ExecuteTxCommandWithFees(currentAdmin, args...)
	require.NoError(t, err)

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	// cosmwasm_migrate
	var success bool
	err = config.QueryDatabase(`
	SELECT 
		success FROM cosmwasm_migrate
		WHERE contract = $1
		AND transaction_hash = $2`, contractAddress, txHash).Scan(&success)
	require.NoError(t, err)
	require.True(t, success)
}

func TestWasmClearAdmin(t *testing.T) {
	// PREPARE
	// Depending on the previous test success
	require.NotEmpty(t, contractAddress)
	currentAdmin := User1
	args := []string{
		wasmModule,
		"clear-contract-admin",
		contractAddress,
	}

	// EXECUTE
	result, err := config.ExecuteTxCommandWithFees(currentAdmin, args...)
	require.NoError(t, err)

	// ASSERT

	// make sure TX is included on chain
	txHash, blockHeight, err := config.IsTxSuccess(result)
	require.NoError(t, err)

	// make sure TX is parsed to DB
	exists := config.IsParsedToTheDb(txHash, blockHeight)
	require.True(t, exists)

	// cosmwasm_clear_admin
	var success bool
	err = config.QueryDatabase(`
	SELECT 
		success FROM cosmwasm_clear_admin
		WHERE contract = $1
		AND transaction_hash = $2`, contractAddress, txHash).Scan(&success)
	require.NoError(t, err)
	require.True(t, success)
}
