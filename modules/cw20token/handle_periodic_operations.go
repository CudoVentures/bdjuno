package cw20token

import (
	"encoding/json"
	"time"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/utils"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
)

func (m *Module) RegisterPeriodicOperations(scheduler *gocron.Scheduler) error {
	log.Debug().Str("module", "cw20token").Msg("setting up periodic tasks")

	if _, err := scheduler.Every(1).Hour().Do(func() {
		utils.WatchMethod(m.removeExpiredAllowances)
	}); err != nil {
		return err
	}

	return nil
}

func (m *Module) removeExpiredAllowances() error {
	return m.db.ExecuteTx(func(dbTx *database.DbTx) error {
		allowances, err := dbTx.GetAllAllowances()
		if err != nil {
			return err
		}

		block, err := dbTx.GetLastBlock()
		if err != nil {
			return err
		}

		for _, a := range allowances {
			expires := struct {
				AtHeight struct{ Height int64 }   `json:"at_height,omitempty"`
				AtTime   struct{ Time time.Time } `json:"at_time,omitempty"`
			}{}

			if err := json.Unmarshal([]byte(a.Expires), &expires); err != nil {
				return err
			}

			heightExpired := expires.AtHeight.Height > 0 && block.Height >= expires.AtHeight.Height
			timeExpired := !expires.AtTime.Time.IsZero() && !block.Timestamp.Before(expires.AtTime.Time)
			if heightExpired || timeExpired {
				if err := dbTx.SaveAllowance(a.Token, a.Owner, a.Spender, "0", ""); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
