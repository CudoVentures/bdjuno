package database

import (
	"fmt"

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
