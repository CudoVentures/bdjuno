package remote

import (
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/forbole/bdjuno/v2/modules/group/source"
	"github.com/forbole/juno/v2/node/remote"
)

var (
	_ source.Source = &Source{}
)

type Source struct {
	*remote.Source
	q group.QueryClient
}

func NewSource(source *remote.Source, gk group.QueryClient) *Source {
	return &Source{
		Source: source,
		q:      gk,
	}
}

func (s Source) GetGroupInfo(groupID uint64, height int64) (*group.GroupInfo, error) {
	res, err := s.q.GroupInfo(s.Ctx, &group.QueryGroupInfoRequest{GroupId: groupID})

	return res.Info, err
}
