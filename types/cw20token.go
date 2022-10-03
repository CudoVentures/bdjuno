package types

// todo move as anonymous struct
type VerifiedContractPublishMessage struct {
	ContractName  string
	CodeID        uint64
	ExecuteSchema string
	QuerySchema   string
}

type MintInfo struct {
	Minter    string
	MaxSupply uint64 `json:"cap,string"`
}

type MarketingInfo struct {
	Project     string
	Description string
	Admin       string `json:"marketing"`
}

type TokenInfo struct {
	Address           string
	Name              string
	Symbol            string
	Decimals          int8
	CirculatingSupply uint64   `json:"total_supply,string"`
	MintInfo          MintInfo `json:"mint"`
	MarketingInfo     MarketingInfo
	Logo              string
	Balances          []*TokenBalance
	CodeID            uint64
}

type TokenBalance struct {
	Address string
	Amount  uint64
}

type MsgTokenExecute struct {
	Recipient   string
	Amount      uint64 `json:"amount,string"`
	Owner       string
	Contract    string
	NewMinter   string `json:"new_minter"`
	Project     string
	Description string
	Admin       string `json:"marketing"`
}
