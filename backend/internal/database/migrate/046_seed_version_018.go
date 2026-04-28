package migrate

import "database/sql"

func up046(tx *sql.Tx) error {
	_, err := tx.Exec(`
		INSERT INTO app_versions (platform, version_name, version_code, is_mandatory, release_notes)
		VALUES ('android', '0.1.8', 108, 0, 'Live TV / IPTV across web + Android (M3U playlists, EPG, recently-watched channels). Delete local NAS downloads from media + series detail. Android TV baseline + NAS download pipeline. HLS audio-track switching and A/V sync drift fixes.')
		ON CONFLICT DO NOTHING;
	`)
	return err
}

func down046(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DELETE FROM app_versions WHERE platform = 'android' AND version_code = 108;
	`)
	return err
}
