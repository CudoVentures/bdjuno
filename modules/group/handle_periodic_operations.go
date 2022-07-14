package group

import (
	"time"

	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/utils"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
)

func (m *Module) RegisterPeriodicOperations(scheduler *gocron.Scheduler) error {
	log.Debug().Str("module", "group").Msg("setting up periodic tasks")

	if _, err := scheduler.Every(1).Hour().Do(func() {
		utils.WatchMethod(m.checkForExpiredGroupProposals)
	}); err != nil {
		return err
	}

	return nil
}

func (m *Module) checkForExpiredGroupProposals() error {
	return m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		proposals, err := dbTx.GetActiveGroupProposalsDecisionPolicies()
		if err != nil {
			return err
		}

		expiredProposals := make([]uint64, 0)

		for _, p := range proposals {
			block, err := m.db.GetLastBlock()
			if err != nil {
				return err
			}
			votingPeriod := time.Second * time.Duration(p.VotingPeriod)
			if p.SubmitTime.Add(votingPeriod).After(block.Timestamp) {
				expiredProposals = append(expiredProposals, p.ID)
			}
		}

		return dbTx.UpdateGroupProposalStatus(
			expiredProposals, group.PROPOSAL_STATUS_REJECTED.String(),
		)
	})
}
