package types

type VerifiedContractPublishMessage struct {
	ContractName  string
	CodeID        uint64
	ExecuteSchema string
	QuerySchema   string
}

type MintInfo struct {
	Minter string
	Cap    uint64 `json:"cap,string"`
}

// todo test TotalSupply was string, now is uint64
type TokenInfo struct {
	Address     string
	Name        string
	Symbol      string
	Decimals    int8
	TotalSupply uint64   `json:"total_supply,string"`
	MintInfo    MintInfo `json:"mint"`
	Balances    []TokenBalance
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
