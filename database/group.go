package database

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/x/group"
	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/types"
	"github.com/lib/pq"
)

func (db *Db) SaveGroupWithPolicy(group *types.GroupWithPolicy) error {
	if _, err := db.Sql.Exec(
		`INSERT INTO group_with_policy
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT DO NOTHING`,
		group.ID, group.Address, group.GroupMetadata, group.PolicyMetadata,
		group.Threshold, group.VotingPeriod, group.MinExecutionPeriod,
	); err != nil {
		return err
	}

	return db.SaveGroupMembers(group.Members, group.ID)
}

func (db *Db) SaveGroupMembers(
	members []group.MemberRequest,
	groupID uint64,
) error {
	stmt := "INSERT INTO group_member VALUES "

	var params []interface{}
	var removedMembers = make([]string, 0)

	count := -1
	for _, m := range members {
		if m.Weight == "0" {
			removedMembers = append(removedMembers, m.Address)
			continue
		}

		count++
		n := count * 4
		stmt += fmt.Sprintf("($%d, $%d, $%d, $%d),", n+1, n+2, n+3, n+4)
		weight, _ := strconv.ParseUint(m.Weight, 10, 64)
		params = append(params, groupID, m.Address, weight, m.Metadata)
	}

	stmt = stmt[:len(stmt)-1]
	stmt += `
	ON CONFLICT (group_id, address) DO UPDATE 
    SET weight = excluded.weight,
    member_metadata = excluded.member_metadata`

	if _, err := db.Sql.Exec(stmt, params...); err != nil {
		return err
	}

	if len(removedMembers) > 0 {
		if _, err := db.Sql.Exec(
			`DELETE FROM group_member
			WHERE group_id = $1
			AND address = ANY($2)`,
			groupID,
			pq.Array(&removedMembers),
		); err != nil {
			return err
		}
	}

	return nil
}

func (db *Db) GetGroupIdByGroupAddress(groupAddress string) uint64 {
	var groupID uint64

	_ = db.Sqlx.QueryRow(
		`SELECT id
		FROM group_with_policy 
		WHERE address = $1`,
		groupAddress,
	).Scan(&groupID)

	return groupID
}

func (db *Db) SaveGroupProposal(proposal *types.GroupProposal) error {
	_, err := db.Sql.Exec(
		`INSERT INTO group_proposal
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, null)
		ON CONFLICT DO NOTHING`,
		proposal.ID, proposal.GroupID, proposal.ProposalMetadata,
		proposal.Proposer, proposal.SubmitTime, proposal.Status,
		proposal.ExecutorResult, proposal.Messages,
	)

	return err
}

func (db *Db) SaveGroupProposalVote(vote *types.GroupProposalVote) error {
	groupID := db.getGroupIdByProposal(vote.ProposalID)

	_, err := db.Sql.Exec(
		`INSERT INTO group_proposal_vote
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING`,
		vote.ProposalID, groupID, vote.Voter,
		vote.VoteOption, vote.VoteMetadata, vote.SubmitTime,
	)

	return err
}

func (db *Db) getGroupIdByProposal(proposalID uint64) uint64 {
	var groupID uint64

	_ = db.Sqlx.QueryRow(
		`SELECT group_id
		FROM group_proposal
		WHERE id = $1`,
		proposalID,
	).Scan(&groupID)

	return groupID
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

func (db *Db) UpdateGroupProposalExecResult(
	proposalID uint64,
	executorResult string,
	txHash string,
) error {
	_, err := db.Sql.Exec(
		`UPDATE group_proposal
		SET executor_result = $1,
		transaction_hash = $2,
		status = 'PROPOSAL_STATUS_ACCEPTED'
		WHERE id = $3`,
		executorResult, txHash, proposalID,
	)

	return err
}

func (db *Db) UpdateGroupProposalsExpiration(blockTime time.Time) error {
	_, err := db.Sql.Exec(
		`UPDATE group_proposal
		SET status = 'PROPOSAL_STATUS_REJECTED'
		FROM group_proposal p
		JOIN group_with_policy g ON g.id = p.group_id
		WHERE p.status = 'PROPOSAL_STATUS_SUBMITTED'
		AND g.voting_period < EXTRACT(EPOCH FROM ($1 - p.submit_time))`,
		blockTime,
	)

	return err
}

func (db *Db) UpdateGroupProposalTallyResult(proposalID uint64) error {
	groupID := db.getGroupIdByProposal(proposalID)

	_, err := db.Sql.Exec(
		`UPDATE group_proposal
		SET status = 
		CASE
			WHEN yes = threshold 
			THEN 'PROPOSAL_STATUS_ACCEPTED'
			WHEN (total - yes) > (members - threshold) 
			THEN 'PROPOSAL_STATUS_REJECTED'
			ELSE status
		END
		FROM (
			SELECT
				COUNT(CASE WHEN vote_option = 'VOTE_OPTION_YES' THEN 1 END) AS yes,
				COUNT(*) as total,
				(SELECT COUNT(*) AS members FROM group_member WHERE group_id = $1),
				(SELECT threshold FROM group_with_policy WHERE id = $1)
			FROM group_proposal_vote
			WHERE proposal_id = $2
		) _
		WHERE id = $2 AND status = 'PROPOSAL_STATUS_SUBMITTED'`,
		groupID,
		proposalID,
	)

	return err
}

func (db *Db) GetGroupProposal(proposalID uint64) (*dbtypes.GroupProposalRow, error) {
	var proposal *dbtypes.GroupProposalRow

	err := db.Sqlx.QueryRow(
		`SELECT *
		FROM group_proposal 
		WHERE id = $1`,
		proposalID,
	).Scan(proposal)

	return proposal, err
}

func (db *Db) UpdateGroupMetadata(groupID uint64, metadata string) error {
	_, err := db.Sql.Exec(
		`UPDATE group_with_policy
		SET group_metadata = $1
		WHERE id = $2`,
		metadata, groupID,
	)

	return err
}

func (db *Db) UpdateGroupPolicyMetadata(groupID uint64, metadata string) error {
	_, err := db.Sql.Exec(
		`UPDATE group_with_policy
		SET policy_metadata = $1
		WHERE id = $2`,
		metadata, groupID,
	)

	return err
}

func (db *Db) UpdateGroupPolicy(
	groupID uint64,
	policy *group.ThresholdDecisionPolicy,
) error {
	_, err := db.Sql.Exec(
		`UPDATE group_with_policy
		SET threshold = $1,
		voting_period = $2,
		min_execution_period = $3
		WHERE id = $4`,
		policy.Threshold,
		policy.Windows.VotingPeriod,
		policy.Windows.MinExecutionPeriod,
		groupID,
	)

	return err
}
