package database

import (
	"fmt"

	dbtypes "github.com/forbole/bdjuno/v2/database/types"
	"github.com/forbole/bdjuno/v2/types"
	"github.com/lib/pq"
)

func (dbTx *DbTx) SaveGroupWithPolicy(group *types.GroupWithPolicy) error {
	_, err := dbTx.Exec(
		`INSERT INTO group_with_policy
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT DO NOTHING`,
		group.ID,
		group.Address,
		group.GroupMetadata,
		group.PolicyMetadata,
		group.Threshold,
		group.VotingPeriod,
		group.MinExecutionPeriod,
	)
	return err
}

func (dbTx *DbTx) SaveGroupMembers(groupID uint64, members []*types.GroupMember) error {
	stmt := "INSERT INTO group_member VALUES "
	var params []interface{}
	for i, m := range members {
		n := i * 4
		stmt += fmt.Sprintf("($%d, $%d, $%d, $%d),", n+1, n+2, n+3, n+4)
		params = append(params, groupID, m.Address, m.Weight, m.Metadata)
	}

	stmt = stmt[:len(stmt)-1]
	stmt += `
	ON CONFLICT (group_id, address) DO UPDATE 
	SET weight = excluded.weight,
		member_metadata = excluded.member_metadata`

	_, err := dbTx.Exec(stmt, params...)
	return err
}

func (dbTx *DbTx) DeleteGroupMembers(members []string, groupID uint64) error {
	_, err := dbTx.Exec(
		`DELETE FROM group_member WHERE group_id = $1 AND address = ANY($2)`,
		groupID, pq.Array(&members),
	)
	return err
}

func (dbTx *DbTx) SaveGroupProposal(proposal *types.GroupProposal) error {
	_, err := dbTx.Exec(
		`INSERT INTO group_proposal
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, null)
		ON CONFLICT DO NOTHING`,
		proposal.ID,
		proposal.GroupID,
		proposal.ProposalMetadata,
		proposal.Proposer,
		proposal.Status,
		proposal.ExecutorResult,
		proposal.Messages,
		proposal.BlockHeight,
	)
	return err
}

func (dbTx *DbTx) SaveGroupProposalVote(vote *types.GroupProposalVote) error {
	_, err := dbTx.Exec(
		`INSERT INTO group_proposal_vote
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING`,
		vote.ProposalID,
		vote.GroupID,
		vote.Voter,
		vote.VoteOption,
		vote.VoteMetadata,
		vote.SubmitTime,
	)
	return err
}

func (dbTx *DbTx) UpdateGroupProposalStatus(proposalIDs []uint64, status string) error {
	_, err := dbTx.Exec(
		`UPDATE group_proposal SET status = $1 WHERE id = ANY($2)`,
		status, pq.Array(proposalIDs),
	)
	return err
}

func (dbTx *DbTx) UpdateGroupProposalExecutorResult(proposalID uint64, executorResult string, txHash string) error {
	_, err := dbTx.Exec(
		`UPDATE group_proposal
		SET executor_result = $1,
			transaction_hash = $2
		WHERE id = $3`,
		executorResult, txHash, proposalID,
	)
	return err
}

func (dbTx *DbTx) UpdateGroupMetadata(groupID uint64, metadata string) error {
	_, err := dbTx.Exec(
		`UPDATE group_with_policy SET group_metadata = $1 WHERE id = $2`,
		metadata, groupID,
	)
	return err
}

func (dbTx *DbTx) UpdateGroupPolicyMetadata(groupID uint64, metadata string) error {
	_, err := dbTx.Exec(
		`UPDATE group_with_policy SET policy_metadata = $1 WHERE id = $2`,
		metadata, groupID,
	)
	return err
}

func (dbTx *DbTx) UpdateGroupPolicy(groupID uint64, policy *types.GroupDecisionPolicy) error {
	_, err := dbTx.Exec(
		`UPDATE group_with_policy
		SET threshold = $1,
			voting_period = $2,
			min_execution_period = $3
		WHERE id = $4`,
		policy.Threshold,
		policy.VotingPeriod,
		policy.MinExecutionPeriod,
		groupID,
	)
	return err
}

func (dbTx *DbTx) GetGroupIDByGroupAddress(groupAddress string) (uint64, error) {
	var groupID uint64
	err := dbTx.QueryRow(
		`SELECT id FROM group_with_policy WHERE address = $1`,
		groupAddress,
	).Scan(&groupID)

	return groupID, err
}

func (dbTx *DbTx) GetGroupProposal(proposalID uint64) (*dbtypes.GroupProposalRow, error) {
	var p dbtypes.GroupProposalRow
	err := dbTx.QueryRow(
		`SELECT * FROM group_proposal WHERE id = $1`,
		proposalID,
	).Scan(
		&p.ID,
		&p.GroupID,
		&p.ProposalMetadata,
		&p.Proposer,
		&p.Status,
		&p.ExecutorResult,
		&p.Messages,
		&p.BlockHeight,
		&p.TxHash,
	)

	return &p, err
}

func (dbTx *DbTx) GetGroupThreshold(groupID uint64) (int, error) {
	var threshold int
	err := dbTx.QueryRow(
		`SELECT threshold FROM group_with_policy WHERE id = $1`,
		groupID,
	).Scan(&threshold)

	return threshold, err
}

func (dbTx *DbTx) GetGroupProposalVotes(proposalID uint64) ([]string, error) {
	rows, err := dbTx.Query(
		`SELECT vote_option FROM group_proposal_vote WHERE proposal_id = $1`,
		proposalID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var votes []string

	for rows.Next() {
		var vote string
		err = rows.Scan(&vote)
		if err != nil {
			return nil, err
		}

		votes = append(votes, vote)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return votes, nil
}

func (dbTx *DbTx) GetGroupTotalVotingPower(groupID uint64) (int, error) {
	var power int
	err := dbTx.QueryRow(
		`SELECT SUM(weight) FROM group_member WHERE group_id = $1`,
		groupID,
	).Scan(&power)

	return power, err
}

func (dbTx *DbTx) GetActiveGroupProposalsDecisionPolicies() ([]*types.GroupProposalDecisionPolicy, error) {
	rows, err := dbTx.Query(
		`SELECT p.id, g.voting_period, g.min_execution_period, b.timestamp
		FROM group_proposal p
		JOIN group_with_policy g ON g.id = p.group_id
		JOIN block b ON b.height = p.height
		WHERE p.status = 'PROPOSAL_STATUS_SUBMITTED'`,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	proposals := make([]*types.GroupProposalDecisionPolicy, 0)

	for rows.Next() {
		var p types.GroupProposalDecisionPolicy
		err = rows.Scan(&p.ID, &p.SubmitTime, &p.VotingPeriod)
		if err != nil {
			return nil, err
		}

		proposals = append(proposals, &p)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return proposals, nil
}

func (dbTx *DbTx) GetGroupProposalDecisionPolicy(proposalID uint64) (*types.GroupProposalDecisionPolicy, error) {
	var p types.GroupProposalDecisionPolicy
	err := dbTx.QueryRow(
		`SELECT p.id, p.status, g.voting_period, g.min_execution_period, b.timestamp
		FROM group_proposal p
		JOIN group_with_policy g ON g.id = p.group_id
		JOIN block b ON b.height = p.height
		WHERE p.id = $1`,
		proposalID,
	).Scan(&p.ID, &p.Status, &p.VotingPeriod, &p.MinExecutionPeriod, &p.SubmitTime)

	return &p, err
}
