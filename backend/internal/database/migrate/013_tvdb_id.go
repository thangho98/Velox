package migrate

import "database/sql"

// 013: Add tvdb_id columns to media and series tables
func up013(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE media ADD COLUMN tvdb_id INTEGER DEFAULT NULL;
		ALTER TABLE series ADD COLUMN tvdb_id INTEGER DEFAULT NULL;

		CREATE INDEX idx_media_tvdb ON media(tvdb_id) WHERE tvdb_id IS NOT NULL;
		CREATE INDEX idx_series_tvdb ON series(tvdb_id) WHERE tvdb_id IS NOT NULL;
	`)
	return err
}

func down013(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP INDEX IF EXISTS idx_media_tvdb;
		DROP INDEX IF EXISTS idx_series_tvdb;
	`)
	return err
}
