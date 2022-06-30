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

func (s Source) GroupInfo(groupID uint64, height int64) (*group.GroupInfo, error) {
	ctx := remote.GetHeightRequestContext(s.Ctx, height)

	res, err := s.q.GroupInfo(ctx, &group.QueryGroupInfoRequest{GroupId: groupID})

	return res.Info, err
}

func (s Source) Proposal(proposalID uint64, height int64) (*group.Proposal, error) {
	ctx := remote.GetHeightRequestContext(s.Ctx, height)

	res, err := s.q.Proposal(ctx, &group.QueryProposalRequest{ProposalId: proposalID})

	return res.Proposal, err
}

func (s Source) Vote(proposalID uint64, voter string, height int64) (*group.Vote, error) {
	ctx := remote.GetHeightRequestContext(s.Ctx, height)

	res, err := s.q.VoteByProposalVoter(ctx, &group.QueryVoteByProposalVoterRequest{
		ProposalId: proposalID,
		Voter:      voter,
	})

	return res.Vote, err
}

func (s Source) GroupPolicyInfo(groupAddress string, height int64) (*group.GroupPolicyInfo, error) {
	ctx := remote.GetHeightRequestContext(s.Ctx, height)

	res, err := s.q.GroupPolicyInfo(ctx, &group.QueryGroupPolicyInfoRequest{
		Address: groupAddress,
	})

	return res.Info, err
}
