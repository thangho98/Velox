package migrate

import "database/sql"

func up032(tx *sql.Tx) error {
	// Seed initial release version for Android App. Marked mandatory because
	// 0.1.6 is the first client that consumes the new ImageResource JSON
	// shape — older APKs render blank artwork after the server image cleanup.
	_, err := tx.Exec(`
		INSERT INTO app_versions (platform, version_name, version_code, is_mandatory, release_notes)
		VALUES ('android', '0.1.6', 106, 1, 'Image system upgrade: ResponsiveImage + blurhash placeholders. Older clients cannot load artwork after the server update and must upgrade.')
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
