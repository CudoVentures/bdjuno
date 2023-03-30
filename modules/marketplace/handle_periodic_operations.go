package marketplace

import (
	"github.com/forbole/bdjuno/v2/client/cryptoCompare"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
)

func (m *Module) RegisterPeriodicOperations(scheduler *gocron.Scheduler) error {
	log.Debug().Str("module", "marketplace").Msg("setting up periodic tasks")

	if _, err := scheduler.Every(15).Minute().Do(func() {
		utils.WatchMethod(m.fetchCudosPrice)
	}); err != nil {
		return err
	}

	return nil
}

func (m *Module) fetchCudosPrice() error {
	usdPrice, err := cryptoCompare.GetCUDOSPrice("usd", m.cfg.CryptoCompareAPIKey)
	if err != nil {
		return err
	}

	btcPrice, err := cryptoCompare.GetCUDOSPrice("btc", m.cfg.CryptoCompareAPIKey)
	if err != nil {
		return err
	}

	m.cudosPrice = types.CudosPrice{USD: usdPrice, BTC: btcPrice}
	return nil
}
