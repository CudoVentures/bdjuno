package types

type VerifiedContractPublishMessage struct {
	ContractName string
	// todo test : CodeID was int, now is uint64
	CodeID        uint64
	ExecuteSchema string
	QuerySchema   string
}

type MintInfo struct {
	Minter string
	Cap    uint64 `json:"cap,string"`
}
type TokenInfo struct {
	Address     string
	Name        string
	Symbol      string
	Decimals    int8
	TotalSupply string   `json:"total_supply"`
	MintInfo    MintInfo `json:"mint"`
	Balances    []TokenBalance
}

type TokenBalance struct {
	Address string
	Amount  uint64
}
