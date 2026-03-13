package migrate

import "database/sql"

func up011(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE app_settings (
			key        TEXT PRIMARY KEY,
			value      TEXT NOT NULL DEFAULT '',
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

func down011(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS app_settings;`)
	return err
}
