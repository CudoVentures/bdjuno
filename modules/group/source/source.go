package source

import (
	"github.com/cosmos/cosmos-sdk/x/group"
)

type Source interface {
	GroupInfo(groupID uint64, height int64) (*group.GroupInfo, error)
	Proposal(proposalID uint64, height int64) (*group.Proposal, error)
	Vote(proposalID uint64, voter string, height int64) (*group.Vote, error)
	GroupPolicyInfo(groupAddress string, height int64) (*group.GroupPolicyInfo, error)
}
