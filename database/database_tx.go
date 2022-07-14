package database

import "database/sql"

type DbTx struct {
	*sql.Tx
}

func (db *Db) ExecuteTx(callback func(*DbTx) error) error {
	tx, err := db.Sqlx.Begin()
	if err != nil {
		return err
	}

	if err = callback(&DbTx{tx}); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}

	return tx.Commit()
}
