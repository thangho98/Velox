package migrate

import "database/sql"

// 028: Add "original" audio-remux profile — video copy + AAC audio.
// This profile always runs (even when pretranscode is disabled) to ensure
// files with incompatible audio codecs can direct-play without realtime transcode.
func up028(tx *sql.Tx) error {
	_, err := tx.Exec(`
		INSERT INTO pretranscode_profiles (name, height, video_bitrate, audio_bitrate, video_codec, audio_codec, enabled)
		VALUES ('original', 0, 0, 192, 'copy', 'aac', 1);
	`)
	return err
}

func down028(tx *sql.Tx) error {
	_, err := tx.Exec(`DELETE FROM pretranscode_profiles WHERE name = 'original' AND video_codec = 'copy';`)
	return err
}
