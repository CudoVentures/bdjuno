package database

import (
	"fmt"
	"time"

	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/types"
)

func (db *Db) SaveGroupWithPolicy(group types.GroupWithPolicy) error {
	_, err := db.Sql.Exec(
		`INSERT INTO group_with_policy
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT DO NOTHING`,
		group.ID, group.Address, group.GroupMetadata, group.PolicyMetadata,
		group.Threshold, group.VotingPeriod, group.MinExecutionPeriod,
	)
	if err != nil {
		return err
	}

	return db.saveGroupMembers(group.Members, group.ID)
}

func (db *Db) saveGroupMembers(
	members []*types.GroupMember,
	groupID uint64,
) error {
	stmt := "INSERT INTO group_member VALUES "
	var params []interface{}
	for i, member := range members {
		n := i * 4
		stmt += fmt.Sprintf("($%d, $%d, $%d, $%d),", n+1, n+2, n+3, n+4)
		params = append(
			params,
			groupID,
			member.Address,
			member.Weight,
			member.MemberMetadata,
		)
	}
	stmt = stmt[:len(stmt)-1]
	stmt += " ON CONFLICT DO NOTHING"

	_, err := db.Sql.Exec(stmt, params...)
	return err
}

func (db *Db) GetGroupID(groupAddress string) (uint64, error) {
	var groupRows []dbtypes.GroupWithPolicyRow
	err := db.Sqlx.Select(
		&groupRows,
		`SELECT id FROM group_with_policy 
		WHERE address = $1
		LIMIT 1`,
		groupAddress)
	if err != nil {
		return 0, err
	}
	return groupRows[0].ID, nil
}

func (db *Db) SaveGroupProposal(proposal types.GroupProposal) error {
	_, err := db.Sql.Exec(
		`INSERT INTO group_proposal
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT DO NOTHING`,
		proposal.ID, proposal.GroupID, proposal.ProposalMetadata,
		proposal.Proposer, proposal.SubmitTime, proposal.Status,
		proposal.ExecutorResult, proposal.Messages,
	)
	return err
}

func (db *Db) SaveGroupProposalVote(vote types.GroupProposalVote) error {
	_, err := db.Sql.Exec(
		`INSERT INTO group_proposal_vote
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING`,
		vote.ProposalID, vote.Voter, vote.VoteOption,
		vote.VoteMetadata, vote.SubmitTime,
	)
	return err
}

func (db *Db) UpdateGroupProposalStatus(proposalID uint64, status string) error {
	_, err := db.Sql.Exec(
		`UPDATE group_proposal
		SET status = $1
		WHERE id = $2`,
		status, proposalID,
	)
	return err
}

func (db *Db) UpdateGroupProposalExecutorResult(
	proposalID uint64,
	executorResult string,
) error {
	_, err := db.Sql.Exec(
		`UPDATE group_proposal
		SET executor_result = $1
		WHERE id = $2`,
		executorResult, proposalID,
	)
	return err
}

func (db *Db) UpdateGroupProposalsExpiration(blockTime time.Time) error {
	_, err := db.Sql.Exec(
		`UPDATE group_proposal
		SET status = 'PROPOSAL_STATUS_REJECTED'
		FROM group_proposal p
		JOIN group_with_policy g ON g.id = p.group_id
		WHERE g.voting_period < $1`,
		blockTime,
	)
	return err
}

func (db *Db) UpdateGroupProposalTallyResult(
	proposalID uint64,
	executorResult string,
) error {
	_, err := db.Sql.Exec(
		`
UPDATE group_proposal
SET status = 
CASE
	WHEN p.votes_yes = p.threshold THEN 'PROPOSAL_STATUS_ACCEPTED'
	WHEN p.votes_no > (p.members - p.threshold) THEN 'PROPOSAL_STATUS_REJECTED'
	ELSE status
END,
executor_result = $1
FROM (
	SELECT 
		(SELECT COUNT(*) FROM group_member WHERE group_id = p.group_id) AS members,
		(SELECT threshold FROM group_with_policy WHERE id = p.group_id) AS threshold,
		COUNT(CASE WHEN v.vote_option = 'VOTE_OPTION_YES' THEN 1 END) AS votes_yes,
		COUNT(CASE WHEN v.vote_option = 'VOTE_OPTION_NO' THEN 1 END) AS votes_no
	FROM group_proposal p
	LEFT JOIN group_proposal_vote v ON v.proposal_id = p.id
	WHERE p.id = $2
	GROUP BY p.group_id
	LIMIT 1
) p
WHERE id = $2`,
		executorResult,
		proposalID,
	)
	return err
}
