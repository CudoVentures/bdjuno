package types

type TokenInfoRow struct {
	Address        string `db:"address"`
	CodeID         uint64 `db:"code_id"`
	Name           string `db:"name"`
	Symbol         string `db:"symbol"`
	Decimals       int8   `db:"decimals"`
	TotalSupply    uint64 `db:"circulating_supply"`
	MaxSupply      uint64 `db:"max_supply"`
	Minter         string `db:"minter"`
	MarketingAdmin string `db:"marketing_admin"`
	ProjectURL     string `db:"project_url"`
	Description    string `db:"description"`
	Logo           string `db:"logo"`
}
