package types

type VerifiedContractPublishMessage struct {
	ContractName      string
	CodeID            int
	InstantiateSchema string
	ExecuteSchema     string
	QuerySchema       string
}

type TokenInstance struct {
	Address     string
	Owner       string
	Name        string
	Denom       string
	TotalSupply string
	MaxSupply   string
}
