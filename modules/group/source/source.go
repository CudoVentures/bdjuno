package source

import (
	"github.com/cosmos/cosmos-sdk/x/group"
)

type Source interface {
	GetGroupInfo(groupID uint64, height int64) (*group.GroupInfo, error)
}
