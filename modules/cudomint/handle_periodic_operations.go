package cudomint

import (
	"context"
	"encoding/json"
	"time"

	"github.com/forbole/bdjuno/v2/modules/utils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
)

// RegisterPeriodicOperations implements modules.PeriodicOperationsModule
func (m *Module) RegisterPeriodicOperations(scheduler *gocron.Scheduler) error {
	log.Debug().Str("module", "cudomint").Msg("setting up periodic tasks")

	// 00:10 because stats service executes at 00:00
	if _, err := scheduler.Every(1).Day().At("00:10").Do(func() {
		utils.WatchMethod(m.fetchStats)
	}); err != nil {
		return err
	}

	return nil
}

func (m *Module) fetchStats() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()
	response, err := m.client.GET(ctx, "/stats")
	if err != nil {
		return err
	}

	var stats statsResponse
	if err := json.Unmarshal([]byte(response), &stats); err != nil {
		return err
	}

	apr, err := sdk.NewDecFromStr(stats.APR.Value)
	if err != nil {
		return err
	}

	if err := m.db.SaveAPR(apr, stats.APR.Height); err != nil {
		return err
	}

	if err := m.db.SaveAPRHistory(apr, stats.APR.Height, time.Now().UnixNano()); err != nil {
		return err
	}

	inflation, err := sdk.NewDecFromStr(stats.Inflation.Value)
	if err != nil {
		return err
	}

	if err := m.db.SaveInflation(inflation, stats.Inflation.Height); err != nil {
		return err
	}

	supply, err := sdk.NewDecFromStr(stats.Supply.Value)
	if err != nil {
		return err
	}

	supply = supply.MulInt64(1000000000000000000)

	if err := m.db.SaveAdjustedSupply(supply, stats.Supply.Height); err != nil {
		return err
	}

	return nil
}

type statsResponse struct {
	Inflation valueAtHeight `json:"inflation"`
	APR       valueAtHeight `json:"apr"`
	Supply    valueAtHeight `json:"supply"`
}

type valueAtHeight struct {
	Value  string `json:"value"`
	Height int64  `json:"height"`
}
