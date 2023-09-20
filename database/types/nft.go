package types

import "database/sql"

type NftFromDB struct {
	TransactionHash       string        `json:"transaction_hash"`
	ID                    uint64        `json:"id"`
	DenomID               string        `json:"denom_id"`
	Name                  string        `json:"name"`
	URI                   string        `json:"uri"`
	DataJSON              string        `json:"data_json"`
	DataText              string        `json:"data_text"`
	Owner                 string        `json:"owner"`
	Sender                string        `json:"sender"`
	ContractAddressSigner string        `json:"contract_address_signer"`
	Burned                bool          `json:"burned"`
	UniqID                sql.NullInt64 `json:"uniq_id"`
}
