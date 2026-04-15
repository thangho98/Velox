package migrate

import (
	"database/sql"
)

func up034(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE image_metadata (
			path        TEXT PRIMARY KEY,
			blurhash    TEXT NOT NULL,
			width       INTEGER NOT NULL,
			height      INTEGER NOT NULL,
			computed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

func down034(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS image_metadata;`)
	return err
}
