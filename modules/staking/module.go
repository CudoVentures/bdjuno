package staking

import (
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/juno/v3/modules"

	"github.com/forbole/bdjuno/v2/database"
	stakingsource "github.com/forbole/bdjuno/v2/modules/staking/source"
)

var (
	_ modules.Module        = &Module{}
	_ modules.GenesisModule = &Module{}
	_ modules.BlockModule   = &Module{}
	_ modules.MessageModule = &Module{}
)

// Module represents the x/staking module
type Module struct {
	cdc                    codec.Codec
	db                     *database.Db
	source                 stakingsource.Source
	slashingModule         SlashingModule
	authModule             AuthModule
	refreshedAccounts      map[string]bool
	refreshedAccountsMutex sync.Mutex
}

// NewModule returns a new Module instance
func NewModule(
	source stakingsource.Source, slashingModule SlashingModule, authModule AuthModule,
	cdc codec.Codec, db *database.Db,
) *Module {
	return &Module{
		cdc:               cdc,
		db:                db,
		source:            source,
		slashingModule:    slashingModule,
		authModule:        authModule,
		refreshedAccounts: make(map[string]bool),
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "staking"
}
