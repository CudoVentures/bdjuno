package pricefeed

import (
	"fmt"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"

	"github.com/forbole/bdjuno/v2/client/cryptoCompare"
	"github.com/forbole/bdjuno/v2/modules/utils"
	"github.com/forbole/bdjuno/v2/types"
)

// RegisterPeriodicOperations implements modules.PeriodicOperationsModule
func (m *Module) RegisterPeriodicOperations(scheduler *gocron.Scheduler) error {
	log.Debug().Str("module", "pricefeed").Msg("setting up periodic tasks")

	// Fetch total supply of token in 30 seconds each
	if _, err := scheduler.Every(30).Second().Do(func() {
		utils.WatchMethod(m.updatePrice)
	}); err != nil {
		return fmt.Errorf("error while setting up pricefeed period operations: %s", err)
	}

	if _, err := scheduler.Every(1).Hour().Do(func() {
		utils.WatchMethod(m.updatePricesHistory)
	}); err != nil {
		return fmt.Errorf("error while setting up pricefeed price history periodic operation: %s", err)
	}

	return nil
}

// updatePrice fetch total amount of coins in the system from RPC and store it into database
func (m *Module) updatePrice() error {
	log.Debug().
		Str("module", "pricefeed").
		Str("operation", "pricefeed").
		Msg("getting token price and market cap")

	prices, err := m.getPrices()
	if err != nil {
		return err
	}

	// Save the token prices
	err = m.db.SaveTokensPrices(prices)
	if err != nil {
		return fmt.Errorf("error while saving token prices: %s", err)
	}

	return nil
}

// updatePricesHistory fetches total amount of coins in the system from RPC
// and stores historical perice data inside the database
func (m *Module) updatePricesHistory() error {
	log.Debug().
		Str("module", "pricefeed").
		Str("operation", "pricefeed").
		Msg("getting token price and market cap history")

	prices, err := m.getPrices()
	if err != nil {
		return err
	}

	return m.historyModule.UpdatePricesHistory(prices)
}

func (m *Module) getPrices() ([]types.TokenPrice, error) {
	// Get the list of tokens price id
	ids, err := m.db.GetTokensPriceID()
	if err != nil {
		return []types.TokenPrice{}, fmt.Errorf("error while getting tokens price id: %s", err)
	}

	if len(ids) == 0 {
		log.Debug().Str("module", "pricefeed").Msg("no traded tokens price id found")
		return []types.TokenPrice{}, nil
	}

	// Get the tokens prices
	prices, err := cryptoCompare.GetTokensPrices("usd", ids, m.cryptoCompareCfg.Config.CryptoCompareApiKey)
	if err != nil {
		return []types.TokenPrice{}, fmt.Errorf("error while getting tokens prices: %s", err)
	}

	return prices, nil
}
