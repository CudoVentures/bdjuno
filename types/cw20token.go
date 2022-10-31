package types

import "encoding/json"

type Mint struct {
	Minter    string `json:"minter"`
	MaxSupply uint64 `json:"cap,string"`
}

type Marketing struct {
	Project     string          `json:"project"`
	Description string          `json:"description"`
	Admin       string          `json:"marketing"`
	Logo        json.RawMessage `json:"logo"`
}

func NewMarketing(project, description, admin string, logo json.RawMessage) Marketing {
	return Marketing{project, description, admin, logo}
}

type TokenBalance struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount,string"`
}

type TokenInfo struct {
	Address     string         `json:"address,omitempty"`
	Name        string         `json:"name"`
	Symbol      string         `json:"symbol"`
	Decimals    int8           `json:"decimals"`
	TotalSupply uint64         `json:"total_supply,string"`
	Mint        Mint           `json:"mint,omitempty"`
	Marketing   Marketing      `json:"marketing_info,omitempty"`
	CodeID      uint64         `json:"code_id"`
	Balances    []TokenBalance `json:"initial_balances,omitempty"`
}

type MsgExecuteToken struct {
	Type              string
	MsgRaw            []byte
	Contract          string
	Sender            string
	Recipient         string
	Amount            uint64 `json:"amount,string"`
	Owner             string
	RecipientContract string `json:"contract"`
	NewMinter         string `json:"new_minter"`
	Project           string
	Description       string
	MarketingAdmin    string `json:"marketing"`
}
