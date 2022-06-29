package types

type GroupMember struct {
	Address        string
	Weight         uint64
	MemberMetadata string
}

type GroupWithPolicy struct {
	ID                 uint64
	Address            string
	Members            []*GroupMember
	GroupMetadata      string
	PolicyMegadata     string
	Threshold          uint64
	VotingPeriod       uint64
	MinExecutionPeriod uint64
}
