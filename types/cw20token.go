package types

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
	Admin             string `json:"marketing"`
}
