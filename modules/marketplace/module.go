package marketplace

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/juno/v2/modules"
	"gopkg.in/yaml.v2"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/types"
)

var (
	_ modules.Module                   = &Module{}
	_ modules.MessageModule            = &Module{}
	_ modules.PeriodicOperationsModule = &Module{}
)

type config struct {
	Config struct {
		CryptoCompareApiKey string `yaml:"crypto_compare_api_key"`
	} `yaml:"marketplace"`
}

// Module represents the nft module
type Module struct {
	cdc        codec.Codec
	db         *database.Db
	cfg        Config
	cudosPrice types.CudosPrice
}

// NewModule returns a new Module instance
func NewModule(cdc codec.Codec, db *database.Db, configBytes []byte) *Module {
	var config config
	if err := yaml.Unmarshal(configBytes, &config); err != nil {
		panic(fmt.Errorf("failed to parse cudomint config: %s", err))
	}
	return &Module{
		cdc: cdc,
		db:  db,
		cfg: cfg,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "marketplace"
}
