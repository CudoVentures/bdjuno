package group

import (
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
	block, err := m.db.GetLastBlock()
	if err != nil {
		return err
	}

	return m.db.UpdateGroupProposalsExpiration(block.Timestamp)
}
