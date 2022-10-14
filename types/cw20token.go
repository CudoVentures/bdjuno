package types

type MsgVerifiedContract struct {
	CodeID        uint64
	ExecuteSchema string
	QuerySchema   string
}

type MintInfo struct {
	Minter    string `json:"minter"`
	MaxSupply uint64 `json:"cap,string"`
}

type MarketingInfo struct {
	Project     string `json:"project"`
	Description string `json:"description"`
	Admin       string `json:"marketing"`
}

func NewMarketingInfo(project, description, admin string) *MarketingInfo {
	return &MarketingInfo{project, description, admin}
}

type TokenBalance struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount,string"`
}
type TokenInfo struct {
	Address           string         `json:"address,omitempty"`
	Name              string         `json:"name"`
	Symbol            string         `json:"symbol"`
	Decimals          int8           `json:"decimals"`
	CirculatingSupply uint64         `json:"total_supply,string"`
	MintInfo          MintInfo       `json:"mint,omitempty"`
	MarketingInfo     MarketingInfo  `json:"marketing_info,omitempty"`
	Logo              string         `json:"logo"`
	CodeID            uint64         `json:"code_id"`
	Balances          []TokenBalance `json:"initial_balances,omitempty"`
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
