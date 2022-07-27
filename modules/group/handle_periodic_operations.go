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
		utils.WatchMethod(m.checkProposalExpirations)
	}); err != nil {
		return err
	}

	return nil
}

func (m *Module) checkProposalExpirations() error {
	return m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		proposals, err := dbTx.GetAllActiveProposals()
		if err != nil {
			return err
		}

		block, err := dbTx.GetLastBlock()
		if err != nil {
			return err
		}

		expiredProposals := make([]uint64, 0)
		for _, p := range proposals {
			votingPeriod := p.SubmitTime.Add(time.Second * time.Duration(p.VotingPeriod))
			if block.Timestamp.After(votingPeriod) {
				expiredProposals = append(expiredProposals, p.ID)
			}
		}

		return dbTx.UpdateProposalStatuses(expiredProposals, group.PROPOSAL_STATUS_REJECTED.String())
	})
}
