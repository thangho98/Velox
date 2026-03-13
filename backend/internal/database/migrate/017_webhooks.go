package migrate

import "database/sql"

// 017: Webhooks for event notifications
func up017(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE webhooks (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			url        TEXT NOT NULL,
			events     TEXT NOT NULL DEFAULT '[]',
			secret     TEXT DEFAULT '',
			active     INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

func down017(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS webhooks;`)
	return err
}
