package migrate

import "database/sql"

func up030(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE app_versions (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			platform       TEXT NOT NULL,
			version_name   TEXT NOT NULL,
			version_code   INTEGER NOT NULL DEFAULT 0,
			is_mandatory   INTEGER NOT NULL DEFAULT 0,
			release_notes  TEXT DEFAULT '',
			created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_app_versions_platform ON app_versions(platform);
	`)
	return err
}

func down030(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS app_versions;`)
	return err
}
