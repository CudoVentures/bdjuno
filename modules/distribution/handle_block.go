package distribution

import (
	"fmt"

	juno "github.com/forbole/juno/v2/types"

	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// HandleBlock implements modules.BlockModule
func (m *Module) HandleBlock(
	b *tmctypes.ResultBlock, _ *tmctypes.ResultBlockResults, _ []*juno.Tx, _ *tmctypes.ResultValidators,
) error {
	// Update the validator commissions amount upon reaching interval or if no commission amount is saved in db
	if m.shouldUpdateValidatorsCommissionAmounts(b.Block.Height) {
		err := m.updateValidatorsCommissionAmounts(b.Block.Height)
		if err != nil {
			return fmt.Errorf("error while updateValidatorsCommissionAmounts")
		}
	}

	// Update the delegators commissions amounts upon reaching interval or no rewards saved yet
	if m.shouldUpdateDelegatorRewardsAmounts(b.Block.Height) {
		err := m.refreshDelegatorsRewardsAmounts(b.Block.Height)
		if err != nil {
			return fmt.Errorf("error while refreshDelegatorsRewardsAmounts")
		}
	}

	return nil
}
