package migrate

import "database/sql"

func up041(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE live_playlists ADD COLUMN epg_url TEXT DEFAULT '';
	`)
	return err
}

func down041(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE live_playlists DROP COLUMN epg_url;
	`)
	return err
}
