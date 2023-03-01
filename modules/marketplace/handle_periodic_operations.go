package marketplace

import (
	"github.com/forbole/bdjuno/v2/client/coingecko"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
)

func (m *Module) RegisterPeriodicOperations(scheduler *gocron.Scheduler) error {
	log.Debug().Str("module", "marketplace").Msg("setting up periodic tasks")

	if _, err := scheduler.Every(1).Minute().Do(func() {
		utils.WatchMethod(m.fetchCudosPrice)
	}); err != nil {
		return err
	}

	return nil
}

func (m *Module) fetchCudosPrice() error {
	usdPrice, err := coingecko.GetCUDOSPrice("usd")
	if err != nil {
		return err
	}

	btcPrice, err := coingecko.GetCUDOSPrice("btc")
	if err != nil {
		return err
	}

	m.cudosPrice = types.CudosPrice{USD: usdPrice, BTC: btcPrice}
	return nil
}
