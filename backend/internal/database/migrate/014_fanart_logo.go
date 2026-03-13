package migrate

import "database/sql"

// 014: Add logo_path and thumb_path columns for fanart.tv artwork
func up014(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE media ADD COLUMN logo_path TEXT DEFAULT '';
		ALTER TABLE media ADD COLUMN thumb_path TEXT DEFAULT '';

		ALTER TABLE series ADD COLUMN logo_path TEXT DEFAULT '';
		ALTER TABLE series ADD COLUMN thumb_path TEXT DEFAULT '';
	`)
	return err
}

func down014(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE media DROP COLUMN logo_path;
		ALTER TABLE media DROP COLUMN thumb_path;

		ALTER TABLE series DROP COLUMN logo_path;
		ALTER TABLE series DROP COLUMN thumb_path;
	`)
	return err
}
