package types

type TokenInfoRow struct {
	Address        string `db:"address"`
	CodeID         uint64 `db:"code_id"`
	Name           string `db:"name"`
	Symbol         string `db:"symbol"`
	Decimals       int8   `db:"decimals"`
	InitialSupply  string `db:"initial_supply"`
	TotalSupply    string `db:"circulating_supply"`
	MaxSupply      string `db:"max_supply"`
	Minter         string `db:"minter"`
	MarketingAdmin string `db:"marketing_admin"`
	ProjectURL     string `db:"project_url"`
	Description    string `db:"description"`
	Logo           string `db:"logo"`
	Type           string `db:"type"`
	Creator        string `db:"creator"`
}

type AllowanceRow struct {
	Token   string `db:"token"`
	Owner   string `db:"owner"`
	Spender string `db:"spender"`
	Amount  string `db:"amount"`
	Expires string `db:"expires"`
}
