package test

import (
	"path/filepath"
	"testing"

	dbTypes "github.com/forbole/bdjuno/v4/database/types"
	config "github.com/forbole/bdjuno/v4/integration_tests/set_up"
	"github.com/stretchr/testify/require"
)

const (
	wasmModule                 = "wasm"
	expectedWasmContractCodeID = "1"
)

func TestWasmMsgStoreCode(t *testing.T) {

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
