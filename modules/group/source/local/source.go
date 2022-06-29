package local

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/forbole/bdjuno/v2/modules/group/source"
	"github.com/forbole/juno/v2/node/local"
)

var (
	_ source.Source = &Source{}
)

type Source struct {
	*local.Source
	q group.QueryServer
}

func NewSource(source *local.Source, gk group.QueryServer) *Source {
	return &Source{
		Source: source,
		q:      gk,
	}
}

func (s Source) GetGroupInfo(groupID uint64, height int64) (*group.GroupInfo, error) {
	ctx, err := s.LoadHeight(height)
	if err != nil {
		return nil, fmt.Errorf("error while loading height: %s", err)
	}

	res, err := s.q.GroupInfo(sdk.WrapSDKContext(ctx), &group.QueryGroupInfoRequest{GroupId: groupID})

	return res.Info, err
}
