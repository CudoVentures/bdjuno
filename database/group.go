package database

import (
	"fmt"
	"strings"

	"github.com/forbole/bdjuno/v2/types"
)

func (db *Db) SaveGroupWithPolicy(group types.GroupWithPolicy) error {
	members := ""
	for _, m := range group.Members {
		members += m.ToString() + ","
	}
	members = strings.TrimSuffix(members, ",")
	stmt := fmt.Sprintf(
		"INSERT INTO group_with_policy VALUES (%d, '%s', '{%s}', '%s', '%s', %d, %d, %d) ON CONFLICT DO NOTHING",
		group.ID,
		group.Address,
		members,
		group.GroupMetadata,
		group.PolicyMegadata,
		group.Threshold,
		group.VotingPeriod,
		group.MinExecutionPeriod,
	)

	_, err := db.Sql.Exec(stmt)
	return err
}
