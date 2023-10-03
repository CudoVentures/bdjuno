package set_up

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/forbole/bdjuno/v4/integration_tests/types"

	// Use effects
	_ "github.com/lib/pq"
)

const (
	dbHost     = "postgres"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "12345"
	dbName     = "bdjuno_test_db"
)

var (
	db                *sql.DB
	averageBlockTime  = 7
	cudosHome         = "/tmp/cudos-test-data"
	cmd               = "cudos-noded"
	Denom             = "acudos"
	txCmd             = "tx"
	queryCmd          = "q"
	from              = "from"
	authModule        = "auth"
	homeFlag          = fmt.Sprintf("--home=%s", cudosHome)
	keyringFlag       = "--keyring-backend=test"
	gasAdjustmentFlag = "--gas-adjustment=1.3"
	gasFlag           = "--gas=8000000"
	feesFlag          = fmt.Sprintf("--fees=1000000000000000000%s", Denom)
	gasPricesFlag     = fmt.Sprintf("--gas-prices=5000000000000%s", Denom)

	GetFlag = func(key string, value string) string {
		return fmt.Sprintf("--%s=%s", key, value)
	}
)

func waitForAvgBlockTime() {
	time.Sleep(time.Duration(averageBlockTime) * time.Second)
}

func getAddress(user string) (string, error) {
	args := []string{
		"keys", "show", user, "-a", keyringFlag,
	}
	return executeCommand(args...)
}

func executeCommand(args ...string) (string, error) {
	args = append(args, []string{homeFlag}...)
	output, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %v; output: %s", err, output)
	}
	return strings.TrimSpace(string(output)), nil
}

func ExecuteTxCommandWithFees(signerAddress string, args ...string) (string, error) {
	additionalFlags := []string{GetFlag(from, signerAddress), keyringFlag, feesFlag, "-y"}
	args = append(args, additionalFlags...)
	args = append([]string{txCmd}, args...)
	return executeCommand(args...)
}

func ExecuteTxCommandWithGas(signerAddress string, args ...string) (string, error) {
	additionalFlags := []string{GetFlag(from, signerAddress), keyringFlag, gasAdjustmentFlag, gasFlag, gasPricesFlag, "-y"}
	args = append(args, additionalFlags...)
	args = append([]string{txCmd}, args...)
	return executeCommand(args...)
}

func IsTxSuccess(result string) (string, uint64, error) {
	txResult := extractTxResult(result)

	if txResult.Code != 0 || txResult.TxHash == "" {
		return "", 0, fmt.Errorf("invalid tx")
	}

	// Waiting for the successful Tx to be included in a block
	waitForAvgBlockTime()

	// Requerying to confirm it is included in a block
	result, err := executeCommand([]string{
		queryCmd, "tx", txResult.TxHash,
	}...)
	if err != nil {
		return "", 0, err
	}

	txResult = extractTxResult(result)
	if txResult.Height > 0 && txResult.Code == 0 {
		return txResult.TxHash, txResult.Height, nil
	}

	return "", 0, fmt.Errorf(txResult.RawLog)
}

func GetTestUserAddress(userNumber int) string {
	result, err := getAddress(fmt.Sprintf("testUser%s", strconv.Itoa(userNumber)))
	if err != nil {
		return ""
	}
	return result
}

func GetTestAdminAddress() string {
	result, err := getAddress("test-admin")
	if err != nil {
		return ""
	}
	return result
}

func SaveToTempFile(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	tmpFile, err := ioutil.TempFile("", "prefix-*.json")
	if err != nil {
		return "", err
	}

	_, err = tmpFile.Write(data)
	if err != nil {
		tmpFile.Close() // ensure closure of the file in case of error
		return "", err
	}

	return tmpFile.Name(), tmpFile.Close()
}

func extractAddressFromQueryResult(queryResultString string) string {
	var address string
	re := regexp.MustCompile(`address: (\w+)`)
	match := re.FindStringSubmatch(queryResultString)
	if len(match) > 1 {
		address = match[1]
	}
	return address
}

func extractTxResult(result string) types.TxResult {
	var txRes types.TxResult

	// Extract code
	codeRegex := regexp.MustCompile(`code: (\d+)`)
	codeMatches := codeRegex.FindStringSubmatch(result)
	if len(codeMatches) > 1 {
		code, _ := strconv.Atoi(codeMatches[1])
		txRes.Code = code
	}

	// Extract height
	heightRegex := regexp.MustCompile(`height: "(\d+)"`)
	heightMatches := heightRegex.FindStringSubmatch(result)
	if len(heightMatches) > 1 {
		height, _ := strconv.ParseUint(heightMatches[1], 10, 64)
		txRes.Height = height
	}

	// Extract txhash
	txhashRegex := regexp.MustCompile(`txhash: (\w+)`)
	txhashMatches := txhashRegex.FindStringSubmatch(result)
	if len(txhashMatches) > 1 {
		txRes.TxHash = txhashMatches[1]
	}

	// Extract logs
	rawLogRegex := regexp.MustCompile(`(?s)raw_log: '(.*?)'`)
	rawLogMatches := rawLogRegex.FindStringSubmatch(result)
	if len(rawLogMatches) > 1 {
		txRes.RawLog = rawLogMatches[1]
	}

	return txRes
}

func setupDB() {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)
	db, _ = sql.Open("postgres", connStr)
}

func QueryDatabase(query string, args ...interface{}) *sql.Row {
	if db == nil {
		setupDB()
	}
	return db.QueryRow(query, args...)
}

func QueryDatabaseMultiRows(query string, args ...interface{}) (*sql.Rows, error) {
	if db == nil {
		setupDB()
	}
	return db.Query(query, args...)
}

func IsParsedToTheDb(txHash string, blockHeight uint64) bool {
	var exists bool

	// Waiting for the successful Tx to be parsed to the DB
	waitForAvgBlockTime()

	// Check if the block is parsed to the DB
	err := QueryDatabase(`SELECT EXISTS(SELECT 1 FROM block WHERE height = $1)`, blockHeight).Scan(&exists)
	if err != nil || !exists {
		return false
	}

	// Check if the tx is parsed to the DB
	err = QueryDatabase(`SELECT EXISTS(SELECT 1 FROM transaction WHERE hash = $1)`, txHash).Scan(&exists)
	if err != nil || !exists {
		return false
	}

	return exists
}

func GetModuleAccountAddress(moduleName string) (string, error) {
	var address string

	args := []string{
		queryCmd,
		authModule,
		"module-account",
		moduleName,
	}

	result, err := executeCommand(args...)
	if err != nil {
		return address, err
	}
	address = extractAddressFromQueryResult(result)
	return address, nil
}
