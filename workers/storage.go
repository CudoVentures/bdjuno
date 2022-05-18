package workers

import (
	"errors"

	"github.com/forbole/bdjuno/v2/database"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db         *database.Db
	workerName string
}

type keyValueRow struct {
	value string `db:"value"`
}

var ErrKeyNotFound = errors.New("key not found")

func NewWorkersStorage(db *database.Db, workerName string) *Storage {
	return &Storage{
		db:         db,
		workerName: workerName,
	}
}

func (ws *Storage) SetValue(key, value string) error {
	workerKey := ws.workerName + "_" + key
	if _, err := ws.db.Sqlx.Exec(sqlx.Rebind(sqlx.DOLLAR, `INSERT INTO workers_storage (key, value) VALUES(?, ?) ON CONFLICT (key) 
		DO UPDATE SET value = EXCLUDED.value`), workerKey, value); err != nil {
		return err
	}
	return nil
}

func (ws *Storage) GetValue(key string) (string, error) {
	var rows []keyValueRow
	if err := ws.db.Sqlx.Select(&rows, sqlx.Rebind(sqlx.DOLLAR, `SELECT value FROM workers_storage WHERE key = ?`), key); err != nil {
		return "", err
	}

	if len(rows) == 0 {
		return "", ErrKeyNotFound
	}

	return rows[0].value, nil
}

func (ws *Storage) GetOrDefaultValue(key, defaultValue string) (string, error) {
	startHeightVal, err := ws.GetValue(startHeightKey)
	if err == ErrKeyNotFound {
		return defaultValue, nil
	}
	return startHeightVal, err
}
