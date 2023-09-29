package types

import (
	"github.com/forbole/bdjuno/v4/modules/actions/types"
)

type Member struct {
	Address  string `json:"address"`
	Weight   string `json:"weight"`
	Metadata string `json:"metadata"`
}

type GroupMembers struct {
	Members []Member `json:"members"`
}

type Window struct {
	VotingPeriod       string `json:"voting_period"`
	MinExecutionPeriod string `json:"min_execution_period"`
}

type DecisionPolicy struct {
	Type      string `json:"@type"`
	Threshold string `json:"threshold"`
	Windows   Window `json:"windows"`
}

type TxResult struct {
	Code   int    `json:"code"`
	Height uint64 `json:"height"`
	TxHash string `json:"txhash"`
	RawLog string `json:"raw_log"`
}

type MsgSend struct {
	Type        string       `json:"@type"`
	FromAddress string       `json:"from_address"`
	ToAddress   string       `json:"to_address"`
	Amount      []types.Coin `json:"amount"`
}

type GroupProposal struct {
	GroupPolicyAddress string    `json:"group_policy_address"`
	Messages           []MsgSend `json:"messages"`
	Metadata           string    `json:"metadata"` // This is a base64 encoded string
	Proposers          []string  `json:"proposers"`
	Title              string    `json:"title"`
	Summary            string    `json:"summary"`
}

type GovProposal struct {
	Messages []MsgSend `json:"messages"`
	Metadata string    `json:"metadata"`
	Deposit  string    `json:"deposit"`
	Title    string    `json:"title"`
	Summary  string    `json:"summary"`
}

type ProposalExecStatuses struct {
	ExecutorResult string `json:"executor_result"`
	Executor       string `json:"executor"`
	ExecutionTime  string `json:"execution_time"`
	ExecutionLog   string `json:"execution_log"`
}

type NftDenomQuery struct {
	DenomID     string `json:"id"`
	Name        string `json:"name"`
	Schema      string `json:"schema"`
	Symbol      string `json:"symbol"`
	Owner       string `json:"owner"`
	Traits      string `json:"traits"`
	Minter      string `json:"minter"`
	Description string `json:"description"`
	DataText    string `json:"data_text"`
}

type NftQuery struct {
	ID      int    `json:"id"`
	DenomID string `json:"denom_id"`
	Name    string `json:"name"`
	URI     string `json:"uri"`
	Owner   string `json:"owner"`
	Sender  string `json:"sender"`
	Burned  bool   `json:"burned"`
	UniqID  string `json:"uniq_id"`
}

type NftTransferQuery struct {
	ID       int    `json:"id"`
	DenomID  string `json:"denom_id"`
	OldOwner string `json:"old_owner"`
	NewOwner string `json:"new_owner"`
	UniqID   string `json:"uniq_id"`
}

type MarketplaceCollectionQuery struct {
	DenomID         string `json:"denom_id"`
	MintRoyalties   string `json:"mint_royalties"`
	ResaleRoyalties string `json:"resale_royalties"`
	Verified        bool   `json:"verified"`
	Creator         string `json:"creator"`
}

type Royalty struct {
	Address string `json:"address"`
	Percent string `json:"percent"`
}
