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

type msg struct {
	msg     string
	wantErr string
}
type contract struct {
	msgs    []msg
	queries []string
}

var interfaces = map[cwStandard]contract{
	CW20: {
		msgs: []msg{
			{`{"transfer":{"recipient":"cudos1","amount":"0"}}`, "Invalid zero amount"},
			{`{"transfer_from":{"owner":"creator","recipient":"cudos1","amount":"0"}}`, "No allowance for this account"},
			{`{"increase_allowance":{"spender":"cudos1","amount":"0"}}`, ""},
		},
		queries: []string{
			`{"all_accounts":{}}`,
			`{"balance":{"address":"cudos1"}}`,
		},
	},
}

const (
	supportedFeatures = "iterator,staking,stargate"
	memoryLimit       = 32
	cacheSize         = 100
	gasLimit          = 100_000_000_000_000
	debugMode         = false
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
	vm, err := wasmvm.NewVM(os.TempDir(), supportedFeatures, memoryLimit, debugMode, cacheSize)
	if err != nil {
		return err
	}

	defer func() {
		vm.Cleanup()
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
		if _, _, err := vm.Execute(checksum, env, info, []byte(m.msg), store, *api, querier, gasMeter, gasLimit, deserCost); err != nil && err.Error() != m.wantErr {
			return err
		}
	}

	for _, q := range interfaces[cw].queries {
		if _, _, err := vm.Query(checksum, env, []byte(q), store, *api, querier, gasMeter, gasLimit, deserCost); err != nil {
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
