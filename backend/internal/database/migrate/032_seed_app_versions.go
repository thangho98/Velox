package migrate

import "database/sql"

func up032(tx *sql.Tx) error {
	// Seed initial release version for Android App into the versions table
	_, err := tx.Exec(`
		INSERT INTO app_versions (platform, version_name, version_code, is_mandatory, release_notes)
		VALUES ('android', '0.1.6', 106, 0, 'Initial version tracking seed.')
		ON CONFLICT DO NOTHING;
	`)
	return err
}

func down032(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DELETE FROM app_versions WHERE platform = 'android' AND version_code = 106;
	`)
	return err
}
