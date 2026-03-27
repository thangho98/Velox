package migrate

import "database/sql"

// 027: Add 1440p and 4K pre-transcode profiles
func up027(tx *sql.Tx) error {
	_, err := tx.Exec(`
		INSERT INTO pretranscode_profiles (name, height, video_bitrate, audio_bitrate) VALUES
			('1440p', 1440, 16000, 192),
			('4K',    2160, 40000, 256);
	`)
	return err
}

func down027(tx *sql.Tx) error {
	_, err := tx.Exec(`DELETE FROM pretranscode_profiles WHERE name IN ('1440p', '4K');`)
	return err
}
