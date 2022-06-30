package database

import (
	"fmt"
	"time"

	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/types"
)

func (db *Db) SaveGroupWithPolicy(data types.GroupWithPolicy) error {
	_, err := db.Sql.Exec(
		`INSERT INTO group_with_policy VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT DO NOTHING`,
		data.ID, data.Address, data.GroupMetadata, data.PolicyMegadata,
		data.Threshold, data.VotingPeriod, data.MinExecutionPeriod,
	)
	if err != nil {
		return err
	}

	return db.saveGroupMembers(data.Members, data.ID)
}

func (db *Db) saveGroupMembers(data []*types.GroupMember, groupID uint64) error {
	stmt := "INSERT INTO group_member VALUES "
	var params []interface{}

	for i, member := range data {
		ai := i * 4
		stmt += fmt.Sprintf("($%d, $%d, $%d, $%d),", ai+1, ai+2, ai+3, ai+4)
		params = append(params, groupID, member.Address, member.Weight, member.MemberMetadata)
	}
	stmt = stmt[:len(stmt)-1]
	stmt += " ON CONFLICT DO NOTHING"

	_, err := db.Sql.Exec(stmt, params...)

	return err
}

func (db *Db) GetGroupId(groupAddress string) uint64 {
	stmt := `SELECT id FROM group_with_policy WHERE address = $1 LIMIT 1`
	var groupRows []dbtypes.GroupWithPolicyRow

	err := db.Sqlx.Select(&groupRows, stmt, groupAddress)
	if err != nil {
		return 0
	}
	return groupRows[0].ID
}

func (db *Db) SaveGroupProposal(data types.GroupProposal) error {
	_, err := db.Sql.Exec(
		`INSERT INTO group_proposal VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT DO NOTHING`,
		data.ID, data.GroupID, data.ProposalMetadata, data.Proposer,
		data.SubmitTime, data.Status, data.ExecutorResult, data.Messages,
	)

	return err
}

func (db *Db) SaveGroupProposalVote(data types.GroupProposalVote) error {
	_, err := db.Sql.Exec(
		`INSERT INTO group_proposal_vote VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING`,
		data.ProposalID, data.Voter, data.VoteOption,
		data.VoteMetadata, data.SubmitTime,
	)

	return err
}

func (db *Db) UpdateGroupProposalStatuses(blockTime time.Time) error {
	_, err := db.Sql.Exec(
		`UPDATE group_proposal
		SET status = 'PROPOSAL_STATUS_REJECTED'
		FROM group_proposal p
		JOIN group_with_policy g ON g.id = p.group_id
		WHERE g.voting_period < EXTRACT(EPOCH FROM ($1 - p.submit_time))`,
		blockTime,
	)

	return err
}
