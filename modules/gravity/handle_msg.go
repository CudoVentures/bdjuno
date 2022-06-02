package gravity

import (
	"fmt"
	"strconv"

	gravityTypes "github.com/althea-net/cosmos-gravity-bridge/module/x/gravity/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/bdjuno/v2/modules/utils"
	juno "github.com/forbole/juno/v2/types"
)

const requiredConsensusPercent = 60

func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	if len(tx.Logs) == 0 {
		return nil
	}

	if cosmosMsg, ok := msg.(*gravityTypes.MsgSendToCosmosClaim); ok {
		return m.handleMsgSendToCosmosClaim(index, cosmosMsg, tx)
	}
	return nil
}

func (m *Module) handleMsgSendToCosmosClaim(index int, msg *gravityTypes.MsgSendToCosmosClaim, tx *juno.Tx) error {
	attestationID := utils.GetValueFromLogs(uint32(index), tx.Logs, sdk.EventTypeMessage, gravityTypes.AttributeKeyAttestationID)
	if len(attestationID) == 0 {
		return fmt.Errorf("attestation id not found: %+v", msg)
	}

	attestationID = strconv.QuoteToASCII(attestationID)

	if err := m.db.SaveMsgSendToCosmosClaim(tx.TxHash, msg.Type(), attestationID, msg.CosmosReceiver, msg.Orchestrator, tx.Height); err != nil {
		return fmt.Errorf("failed to save send to cosmos claim %+v: %v", msg, err)
	}

	votes, err := m.db.GetGravityTransactionVotes(attestationID)
	if err != nil {
		return fmt.Errorf("failed to get gravity transaction votes: %v", err)
	}

	orchestratorCount, err := m.db.GetOrchestratorsCount()
	if err != nil {
		return fmt.Errorf("failed to get orchestrators count: %v", err)
	}

	if isConsensusReached(votes, orchestratorCount) {
		if err := m.db.SetGravityTransactionConsensus(attestationID, true); err != nil {
			return fmt.Errorf("setting gravity transaction consensus failed")
		}
	}

	return nil
}

func isConsensusReached(votes, orchestratorsCount int) bool {
	consensusPercent := (float64(votes) * float64(100)) / float64(orchestratorsCount)
	return consensusPercent >= requiredConsensusPercent
}
