package pricefeed

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/juno/v5/types/config"

	"github.com/forbole/bdjuno/v4/database"

	"github.com/forbole/bdjuno/v4/client/cryptocompare"
	"github.com/forbole/juno/v5/modules"
)

var (
	_ modules.Module                     = &Module{}
	_ modules.AdditionalOperationsModule = &Module{}
	_ modules.PeriodicOperationsModule   = &Module{}
)

// Module represents the module that allows to get the token prices
type Module struct {
	cfg *Config
	ccc *cryptocompare.Client
	cdc codec.Codec
	db  *database.Db
}

// NewModule returns a new Module instance
func NewModule(cfg config.Config, cryptoCompareClient *cryptocompare.Client, cdc codec.Codec, db *database.Db) *Module {
	bz, err := cfg.GetBytes()
	if err != nil {
		panic(err)
	}

	pricefeedCfg, err := ParseConfig(bz)
	if err != nil {
		panic(err)
	}

	return &Module{
		cfg: pricefeedCfg,
		ccc: cryptoCompareClient,
		cdc: cdc,
		db:  db,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "pricefeed"
}
