package marketplace

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/juno/v2/modules"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/types"
)

var (
	_ modules.Module                   = &Module{}
	_ modules.MessageModule            = &Module{}
	_ modules.PeriodicOperationsModule = &Module{}
)

// Module represents the nft module
type Module struct {
	cdc        codec.Codec
	db         *database.Db
	cudosPrice types.CudosPrice
}

// NewModule returns a new Module instance
func NewModule(cdc codec.Codec, db *database.Db) *Module {
	return &Module{
		cdc: cdc,
		db:  db,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "marketplace"
}
