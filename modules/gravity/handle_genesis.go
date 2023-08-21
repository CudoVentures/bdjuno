package gravity

import (
	"encoding/json"
	"fmt"

	gravityTypes "github.com/althea-net/cosmos-gravity-bridge/module/x/gravity/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/rs/zerolog/log"
)

func (m *Module) HandleGenesis(doc *tmtypes.GenesisDoc, _ map[string]json.RawMessage) error {
	log.Debug().Str("module", "gravity").Msg("parsing genesis")

	type genesis struct {
		Gravity struct {
			DelegateKeys []*gravityTypes.MsgSetOrchestratorAddress `json:"delegate_keys,omitempty"`
		} `json:"gravity"`
	}

	var state genesis
	if err := json.Unmarshal(doc.AppState, &state); err != nil {
		return fmt.Errorf("error while unmarshalling gravity state: %v", err)
	}

	for _, delegateKey := range state.Gravity.DelegateKeys {
		if err := m.db.SaveOrchestrator(delegateKey.Orchestrator); err != nil {
			return fmt.Errorf("saving orchestrator failed: %v", err)
		}
	}

	return nil
}
