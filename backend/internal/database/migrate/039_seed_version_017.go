package migrate

import "database/sql"

func up039(tx *sql.Tx) error {
	_, err := tx.Exec(`
		INSERT INTO app_versions (platform, version_name, version_code, is_mandatory, release_notes)
		VALUES ('android', '0.1.7', 107, 1, 'Cloud storage: FShare integration with subtitle extraction, metadata fallback (movie → TV series), scheduled cloud probe backfill, AniList OAuth connect.')
		ON CONFLICT DO NOTHING;
	`)
	return err
}

func down039(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DELETE FROM app_versions WHERE platform = 'android' AND version_code = 107;
	`)
	return err
}
