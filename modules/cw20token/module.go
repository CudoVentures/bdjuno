package cw20token

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/juno/v2/modules"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/cw20token/source"
)

var (
	_ modules.Module                   = &Module{}
	_ modules.MessageModule            = &Module{}
	_ modules.PeriodicOperationsModule = &Module{}
)

type Module struct {
	cdc    codec.Codec
	db     *database.Db
	source source.Source
}

func NewModule(cdc codec.Codec, db *database.Db, source source.Source) *Module {
	return &Module{
		cdc:    cdc,
		db:     db,
		source: source,
	}
}

func (m *Module) Name() string {
	return "cw20token"
}
