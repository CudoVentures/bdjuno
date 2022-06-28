package cudomint

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/juno/v2/modules"
	"gopkg.in/yaml.v2"

	"github.com/forbole/bdjuno/v2/database"
)

var (
	_ modules.Module                   = &Module{}
	_ modules.GenesisModule            = &Module{}
	_ modules.PeriodicOperationsModule = &Module{}
)

// Module represent database/mint module
type Module struct {
	cdc    codec.Codec
	db     *database.Db
	config cudomintConfig
}

type cudomintConfig struct {
	EthNode      string   `yaml:"eth_node"`
	TokenAddress string   `yaml:"token_address"`
	EthAccounts  []string `yaml:"eth_accounts"`
}

type config struct {
	Cudomint cudomintConfig `yaml:"cudomint"`
}

// NewModule returns a new Module instance
func NewModule(cdc codec.Codec, db *database.Db, configBytes []byte) *Module {
	var config config
	if err := yaml.Unmarshal(configBytes, &config); err != nil {
		panic(fmt.Errorf("failed to parse cudomint config: %s", err))
	}

	return &Module{
		cdc:    cdc,
		db:     db,
		config: config.Cudomint,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "cudomint"
}
