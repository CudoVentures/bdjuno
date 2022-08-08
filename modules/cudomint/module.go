package cudomint

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/juno/v2/modules"
	"gopkg.in/yaml.v2"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/cudomint/rest"
)

var (
	_ modules.Module                   = &Module{}
	_ modules.PeriodicOperationsModule = &Module{}
)

// Module represent database/mint module
type Module struct {
	cdc    codec.Codec
	db     *database.Db
	client *rest.Client
}

type config struct {
	Config struct {
		StatsServiceURL string `yaml:"stats_service_url"`
	} `yaml:"cudomint"`
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
		client: rest.NewClient(config.Config.StatsServiceURL),
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "cudomint"
}
