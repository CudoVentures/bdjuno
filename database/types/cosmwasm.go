package types

type CosmwasmStoreRow struct {
	TransactionHash       string `db:"transaction_hash"`
	Index                 int    `db:"index"`
	Sender                string `db:"sender"`
	InstantiatePermission string `db:"instantiate_permission"`
	ResultCodeId          string `db:"result_code_id"`
	Success               bool   `db:"success"`
}

type CosmwasmInstantiateRow struct {
	TransactionHash       string `db:"transaction_hash"`
	Index                 int    `db:"index"`
	Admin                 string `db:"admin"`
	Funds                 string `db:"funds"`
	Label                 string `db:"label"`
	Sender                string `db:"sender"`
	CodeId                string `db:"code_id"`
	ResultContractAddress string `db:"result_contract_address"`
	Success               bool   `db:"success"`
}

type CosmwasmExecuteRow struct {
	TransactionHash string `db:"transaction_hash"`
	Index           int    `db:"index"`
	Method          string `db:"method"`
	Arguments       string `db:"arguments"`
	Funds           string `db:"funds"`
	Sender          string `db:"sender"`
	Contract        string `db:"contract"`
	Success         bool   `db:"success"`
}

type CosmwasmMigrateRow struct {
	TransactionHash string `db:"transaction_hash"`
	Index           int    `db:"index"`
	Sender          string `db:"sender"`
	Contract        string `db:"contract"`
	CodeId          string `db:"code_id"`
	Arguments       string `db:"arguments"`
	Success         bool   `db:"success"`
}

type CosmwasmUpdateAdminRow struct {
	TransactionHash string `db:"transaction_hash"`
	Index           int    `db:"index"`
	Sender          string `db:"sender"`
	Contract        string `db:"contract"`
	NewAdmin        string `db:"new_admin"`
	Success         bool   `db:"success"`
}

type CosmwasmClearAdminRow struct {
	TransactionHash string `db:"transaction_hash"`
	Index           int    `db:"index"`
	Sender          string `db:"sender"`
	Contract        string `db:"contract"`
	Success         bool   `db:"success"`
}
