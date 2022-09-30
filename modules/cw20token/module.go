package cw20token

import (
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/juno/v2/modules"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/cw20token/source"
	"github.com/forbole/bdjuno/v2/utils"
)

var (
	_ modules.Module                     = &Module{}
	_ modules.MessageModule              = &Module{}
	_ modules.AdditionalOperationsModule = &Module{}
)

type Module struct {
	cdc    codec.Codec
	db     *database.Db
	source source.Source
	pubsub utils.PubSub
	mu     sync.Mutex
}

func NewModule(cdc codec.Codec, db *database.Db, source source.Source, pubsub utils.PubSub) *Module {
	return &Module{
		cdc:    cdc,
		db:     db,
		source: source,
		pubsub: pubsub,
		mu:     sync.Mutex{},
	}
}

func (m *Module) Name() string {
	return "cw20token"
}
