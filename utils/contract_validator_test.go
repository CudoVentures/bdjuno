package utils

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContractValidator_CW20(t *testing.T) {
	for testName, tc := range map[string]struct {
		wasmPath string
		wantErr  error
	}{
		"valid": {
			wasmPath: "../testdata/cw20_base.wasm",
		},
		"invalid": {
			wasmPath: "../testdata/alpha.wasm",
			wantErr:  fmt.Errorf("Error parsing into type alpha::msg::ExecuteMsg: unknown variant `transfer`, expected `increment` or `reset`"),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			contractWasm, err := os.ReadFile(tc.wasmPath)
			require.NoError(t, err)

			haveErr := ValidateContract(contractWasm, CW20)
			require.Equal(t, tc.wantErr, haveErr)
		})
	}
}
