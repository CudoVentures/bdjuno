package utils

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"

	wasmutils "github.com/CosmWasm/wasmd/x/wasm/client/utils"
	wasmvm "github.com/CosmWasm/wasmvm"
	wasmvmapi "github.com/CosmWasm/wasmvm/api"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
)

type cwStandard int

const (
	CW20 cwStandard = iota
	CW721
	CW1155
)

type contract struct {
	msgs    []string
	queries []string
}

var interfaces = map[cwStandard]contract{
	CW20: {
		msgs: []string{
			`{"transfer":{"recipient":"cudos1","amount":"0"}}`,
			// todo add all must functions
		},
		queries: []string{
			`{"all_accounts":{}}`,
			`{"balance":{"address":"cudos1"}}`,
		},
	},
	CW721: {
		msgs: []string{
			`{"":{}}`,
		},
		queries: []string{
			`{"":{}}`,
		},
	},
	CW1155: {
		msgs: []string{
			`{"":{}}`,
		},
		queries: []string{
			`{"":{}}`,
		},
	},
}

const (
	SUPPORTED_FEATURES = "iterator,staking,stargate"
	MEMORY_LIMIT       = 32
	CACHE_SIZE         = 100
	GAS_LIMIT          = 100_000_000_000_000
	DEBUG              = false

	// this error is allowed because both initial_balances and mint are optional
	// so we don't have a good way to fund, that's why we transfer 0 amount
	ALLOWED_ERR = "Invalid zero amount"
)

var (
	gasMeter  = wasmvmapi.NewMockGasMeter(100_000_000_000_000)
	store     = wasmvmapi.NewLookup(gasMeter)
	api       = wasmvmapi.NewMockAPI()
	querier   = wasmvmapi.DefaultQuerier(wasmvmapi.MOCK_CONTRACT_ADDR, nil)
	env       = wasmvmapi.MockEnv()
	info      = wasmvmapi.MockInfo("creator", nil)
	deserCost = wasmvmtypes.UFraction{1, 1}
)

func ValidateContract(contract wasmvm.WasmCode, cw cwStandard) error {
	tmpdir := os.TempDir()
	vm, err := wasmvm.NewVM(tmpdir, SUPPORTED_FEATURES, MEMORY_LIMIT, DEBUG, CACHE_SIZE)
	if err != nil {
		return err
	}

	defer func() {
		vm.Cleanup()
		os.RemoveAll(tmpdir)
	}()

	// wasmCode coming from indexed MsgStoreCode is compressed
	// in order to be runnable it needs to be decompressed
	if wasmutils.IsGzip(contract) {
		if contract, err = decompress(contract); err != nil {
			return err
		}
	}

	checksum, err := vm.Create(contract)
	if err != nil {
		return err
	}

	for _, m := range interfaces[cw].msgs {
		if _, _, err := vm.Execute(checksum, env, info, []byte(m), store, *api, querier, gasMeter, GAS_LIMIT, deserCost); err != nil && err.Error() != ALLOWED_ERR {
			return err
		}
	}

	for _, q := range interfaces[cw].queries {
		if _, _, err := vm.Query(checksum, env, []byte(q), store, *api, querier, gasMeter, GAS_LIMIT, deserCost); err != nil {
			return err
		}
	}

	return nil
}

func decompress(data wasmvm.WasmCode) (wasmvm.WasmCode, error) {
	var buf bytes.Buffer
	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	defer gr.Close()

	data, err = ioutil.ReadAll(gr)
	if err != nil {
		return nil, err
	}

	buf.Write(data)
	return buf.Bytes(), nil
}
