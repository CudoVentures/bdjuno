package types

import (
	"fmt"
)

type GroupMember struct {
	Address        string
	Weight         int64
	MemberMetadata string
}

func (m *GroupMember) ToString() string {
	return fmt.Sprintf("\"(%s, %d, %s)\"", m.Address, m.Weight, m.MemberMetadata)
}

type GroupWithPolicy struct {
	ID                 uint64
	Address            string
	Members            []*GroupMember
	GroupMetadata      string
	PolicyMegadata     string
	Threshold          int64
	VotingPeriod       int64
	MinExecutionPeriod int64
}
