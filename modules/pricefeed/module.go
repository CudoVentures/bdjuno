package pricefeed

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/juno/v2/types/config"

	"github.com/forbole/bdjuno/v2/database"

	"github.com/forbole/bdjuno/v2/client/cryptoCompare"
	"github.com/forbole/juno/v2/modules"
)

var (
	_ modules.Module                     = &Module{}
	_ modules.AdditionalOperationsModule = &Module{}
	_ modules.PeriodicOperationsModule   = &Module{}
)

// Module represents the module that allows to get the token prices
type Module struct {
	cfg           *Config
	ccc           *cryptoCompare.CryptoCompareClient
	cdc           codec.Codec
	db            *database.Db
	historyModule HistoryModule
}

// NewModule returns a new Module instance
func NewModule(cfg config.Config, cryptoCompareClient *cryptoCompare.CryptoCompareClient, historyModule HistoryModule, cdc codec.Codec, db *database.Db) *Module {
	pricefeedCfg, err := ParseConfig(cfg.GetBytes())
	if err != nil {
		panic(err)
	}

	return &Module{
		cfg:           pricefeedCfg,
		ccc:           cryptoCompareClient,
		cdc:           cdc,
		db:            db,
		historyModule: historyModule,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "pricefeed"
}
