package cw20token

import (
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/juno/v2/modules"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/forbole/bdjuno/v2/modules/cw20token/source"
	"github.com/forbole/bdjuno/v2/utils/pubsub"
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
	pubsub pubsub.PubSubClient
	// we use mutex.Lock() to keep Subscribe and InstantiateMsg in sync
	// for example a matching contract may get saved in the db
	// right after subscribeCallback starts updating matching contracts
	// in this case, that matching contract would never be recognized as token
	mu *sync.Mutex
}

func NewModule(cdc codec.Codec, db *database.Db, source source.Source, pubsub pubsub.PubSubClient) *Module {
	return &Module{
		cdc:    cdc,
		db:     db,
		source: source,
		pubsub: pubsub,
		mu:     &sync.Mutex{},
	}
}

func (m *Module) Name() string {
	return "cw20token"
}
