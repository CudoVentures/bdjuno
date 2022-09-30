package cw20token

import (
	"fmt"

	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	juno "github.com/forbole/juno/v2/types"
)

func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	if len(tx.Logs) == 0 {
		return nil
	}

	switch cosmosMsg := msg.(type) {
	case *wasmTypes.MsgInstantiateContract:
		m.mu.Lock()
		defer m.mu.Unlock()
		fmt.Print(cosmosMsg)
	case *wasmTypes.MsgExecuteContract:
		return nil
	}

	return nil
}
