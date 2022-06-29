package group

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/forbole/bdjuno/v2/types"
	juno "github.com/forbole/juno/v2/types"
)

func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	if len(tx.Logs) == 0 {
		return nil
	}

	switch cosmosMsg := msg.(type) {
	case *group.MsgCreateGroupWithPolicy:
		return m.handleMsgCreateGroupWithPolicy(tx, index, cosmosMsg)
	}

	return nil
}

func (m *Module) handleMsgCreateGroupWithPolicy(tx *juno.Tx, index int, msg *group.MsgCreateGroupWithPolicy) error {
	eventCreateGroup, err := tx.FindEventByType(index, "cosmos.group.v1.EventCreateGroup")
	if err != nil {
		return fmt.Errorf("error while searching for EventCreateGroup: %s", err)
	}
	groupIdAttr, err := tx.FindAttributeByKey(eventCreateGroup, "group_id")
	if err != nil {
		return fmt.Errorf("error while searching for AttributeKeyGroupID: %s", err)
	}
	groupIdAttr, _ = strconv.Unquote(groupIdAttr)
	groupID, err := strconv.ParseUint(groupIdAttr, 10, 64)
	if err != nil {
		return fmt.Errorf("error while parsing groupID: %s", err)
	}

	eventCreateGroupPolicy, err := tx.FindEventByType(index, "cosmos.group.v1.EventCreateGroupPolicy")
	if err != nil {
		return fmt.Errorf("error while searching for EventCreateGroupPolicy: %s", err)
	}
	address, err := tx.FindAttributeByKey(eventCreateGroupPolicy, "address")
	if err != nil {
		return fmt.Errorf("error while searching for AttributeKeyAddress: %s", err)
	}
	address, _ = strconv.Unquote(address)

	members := make([]*types.GroupMember, 0)
	for _, m := range msg.Members {
		weight, _ := strconv.ParseUint(m.Weight, 10, 64)
		member := &types.GroupMember{Address: m.Address, Weight: weight, MemberMetadata: m.Metadata}
		members = append(members, member)
	}

	decisionPolicy, _ := msg.DecisionPolicy.GetCachedValue().(*group.ThresholdDecisionPolicy)
	threshold, _ := strconv.ParseUint(decisionPolicy.Threshold, 10, 64)

	return m.db.SaveGroupWithPolicy(types.GroupWithPolicy{
		ID:                 groupID,
		Address:            address,
		Members:            members,
		GroupMetadata:      msg.GroupMetadata,
		PolicyMegadata:     msg.GroupPolicyMetadata,
		Threshold:          threshold,
		VotingPeriod:       uint64(decisionPolicy.Windows.VotingPeriod.Seconds()),
		MinExecutionPeriod: uint64(decisionPolicy.Windows.MinExecutionPeriod.Seconds()),
	})
}
