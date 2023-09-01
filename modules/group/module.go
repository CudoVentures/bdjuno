package group

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/forbole/bdjuno/v4/database"

	"github.com/forbole/juno/v5/modules"
)

var (
	_ modules.Module                   = &Module{}
	_ modules.PeriodicOperationsModule = &Module{}
	_ modules.MessageModule            = &Module{}
)

type Module struct {
	cdc codec.Codec
	db  *database.Db
}

func NewModule(cdc codec.Codec, db *database.Db,
) *Module {
	return &Module{
		cdc: cdc,
		db:  db,
	}
}

func (m *Module) Name() string {
	return "group"
}
