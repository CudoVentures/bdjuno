package cudomint

import (
	"encoding/json"
	"fmt"

	"github.com/forbole/bdjuno/v2/types"
	tmtypes "github.com/tendermint/tendermint/types"

	cudoMintTypes "github.com/CudoVentures/cudos-node/x/cudoMint/types"
	"github.com/rs/zerolog/log"
)

// HandleGenesis implements modules.Module
func (m *Module) HandleGenesis(doc *tmtypes.GenesisDoc, appState map[string]json.RawMessage) error {
	log.Debug().Str("module", "cudomint").Msg("parsing genesis")

	// Read the genesis state
	var genState cudoMintTypes.GenesisState
	err := m.cdc.UnmarshalJSON(appState[cudoMintTypes.ModuleName], &genState)
	if err != nil {
		return fmt.Errorf("error while reading mint genesis data: %s", err)
	}

	// Save the params
	err = m.db.SaveMintParams(types.NewMintParams(genState, doc.InitialHeight))
	if err != nil {
		return fmt.Errorf("error while storing genesis mint params: %s", err)
	}

	return nil
}
