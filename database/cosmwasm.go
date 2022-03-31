package database

import "github.com/forbole/bdjuno/v2/types"

func (db *Db) SaveMsgStoreCodeData(data *types.MsgStoreCodeData) error {
	_, err := db.Sql.Exec(
		`INSERT INTO cosmwasm_store (
			transaction_hash, index, sender, instantiate_permission, result_code_id, success
		)  VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (transaction_hash, index) DO UPDATE SET 
		sender = excluded.sender, instantiate_permission = excluded.instantiate_permission, 
		result_code_id = excluded.result_code_id, success = excluded.success`,
		data.TxHash, data.Index, data.Sender, data.InstantiatePermission, data.ResultCodeId, data.Success,
	)
	return err
}

func (db *Db) SaveMsgInstantiateContractData(data *types.MsgInstantiateContractData) error {
	_, err := db.Sql.Exec(
		`INSERT INTO cosmwasm_instantiate (
			transaction_hash, index, admin, funds, label, sender, code_id,
			result_contract_address, success
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT (transaction_hash, index) DO UPDATE SET 
		admin = excluded.admin, funds = excluded.funds, label = excluded.label, sender = excluded.sender, 
		code_id = excluded.code_id, result_contract_address = excluded.result_contract_address, success = excluded.success`,
		data.TxHash, data.Index, data.Admin, data.Funds, data.Label,
		data.Sender, data.CodeId, data.ResultContractAddress, data.Success,
	)
	return err
}

func (db *Db) SaveMsgExecuteContractData(data *types.MsgExecuteContractData) error {
	_, err := db.Sql.Exec(
		`INSERT INTO cosmwasm_execute (
			transaction_hash, index, method, arguments,
			funds, sender, contract, success
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (transaction_hash, index) DO UPDATE SET 
		method = excluded.method, arguments = excluded.arguments, funds = excluded.funds, 
		sender = excluded.sender, contract = excluded.contract, success = excluded.success`,
		data.TxHash, data.Index, data.Method, data.Arguments,
		data.Funds, data.Sender, data.Contract, data.Success,
	)
	return err
}

func (db *Db) SaveMsgMigrateContactData(data *types.MsgMigrateContractData) error {
	_, err := db.Sql.Exec(
		`INSERT INTO cosmwasm_migrate (
			transaction_hash, index, contract, code_id,
			arguments, sender, success
		) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (transaction_hash, index) DO UPDATE SET 
		contract = excluded.contract, code_id = excluded.code_id, arguments = excluded.arguments, 
		sender = excluded.sender, success = excluded.success`,
		data.TxHash, data.Index, data.Contract, data.CodeId,
		data.Arguments, data.Sender, data.Success,
	)
	return err
}

func (db *Db) SaveMsgUpdateAdminData(data *types.MsgUpdateAdminData) error {
	_, err := db.Sql.Exec(
		`INSERT INTO cosmwasm_update_admin (
			transaction_hash, index, contract, new_admin,
			sender, success
		) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (transaction_hash, index) DO UPDATE SET 
		contract = excluded.contract, new_admin = excluded.new_admin, sender = excluded.sender, 
		success = excluded.success`,
		data.TxHash, data.Index, data.Contract, data.NewAdmin,
		data.Sender, data.Success,
	)
	return err
}

func (db *Db) SaveMsgClearAdminData(data *types.MsgClearAdminData) error {
	_, err := db.Sql.Exec(
		`INSERT INTO cosmwasm_clear_admin (
			transaction_hash, index, contract,
			sender, success
		) VALUES($1, $2, $3, $4, $5) ON CONFLICT (transaction_hash, index) DO UPDATE SET 
		contract = excluded.contract, sender = excluded.sender, success = excluded.success`,
		data.TxHash, data.Index, data.Contract, data.Sender, data.Success,
	)
	return err
}
