package pricefeed

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"

	"github.com/forbole/bdjuno/v4/modules/utils"
	"github.com/forbole/bdjuno/v4/types"
)

// RegisterPeriodicOperations implements modules.PeriodicOperationsModule
func (m *Module) RegisterPeriodicOperations(scheduler *gocron.Scheduler) error {
	log.Debug().Str("module", "pricefeed").Msg("setting up periodic tasks")

	// Fetch the token prices every 5 mins
	if _, err := scheduler.Every(5).Minutes().Do(func() {
		utils.WatchMethod(m.UpdatePrice)
	}); err != nil {
		return fmt.Errorf("error while setting up pricefeed period operations: %s", err)
	}

	// Update the historical token prices every 1 hour
	if _, err := scheduler.Every(1).Hour().Do(func() {
		utils.WatchMethod(m.UpdatePricesHistory)
	}); err != nil {
		return fmt.Errorf("error while setting up history period operations: %s", err)
	}

	return nil
}

// UpdatePrice fetches the total amount of coins in the system from RPC and stores it in database
func (m *Module) UpdatePrice() error {
	log.Debug().
		Str("module", "pricefeed").
		Str("operation", "pricefeed").
		Msg("updating token price and market cap")

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

// UpdatePricesHistory fetches total amount of coins in the system from RPC
// and stores historical perice data inside the database
func (m *Module) UpdatePricesHistory() error {
	log.Debug().
		Str("module", "pricefeed").
		Str("operation", "pricefeed").
		Msg("updating token price and market cap history")

	prices, err := m.getPrices()
	if err != nil {
		return err
	}

	// Normally, the last updated value reflects the time when the price was last updated.
	// If price hasn't changed, the returned timestamp will be the same as one hour ago, and it will not
	// be stored in db as it will be a duplicated value.
	// To fix this, we set each price timestamp to be the same as other ones.
	timestamp := time.Now()
	for _, price := range prices {
		price.Timestamp = timestamp
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
	prices, err := m.ccc.GetTokensPrices("usd", ids)
	if err != nil {
		return []types.TokenPrice{}, fmt.Errorf("error while getting tokens prices: %s", err)
	}

	return prices, nil
}
