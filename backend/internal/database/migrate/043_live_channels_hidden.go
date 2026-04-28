package migrate

import "database/sql"

func up043(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE live_channels ADD COLUMN is_hidden INTEGER NOT NULL DEFAULT 0;
	`)
	return err
}

func down043(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE live_channels DROP COLUMN is_hidden;
	`)
	return err
}
