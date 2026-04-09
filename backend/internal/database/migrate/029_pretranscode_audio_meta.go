package migrate

import "database/sql"

// 029: Add real audio metadata to pretranscode_files so the playback overlay
// can show what is actually being streamed (not the source MKV's AC3 5.1).
// Worker probes the encoded MP4 with ffprobe and stores codec/channels/bitrate/sample_rate.
func up029(tx *sql.Tx) error {
	checks := []struct {
		column string
		ddl    string
	}{
		{"audio_codec", `ALTER TABLE pretranscode_files ADD COLUMN audio_codec TEXT NOT NULL DEFAULT ''`},
		{"audio_channels", `ALTER TABLE pretranscode_files ADD COLUMN audio_channels INTEGER NOT NULL DEFAULT 0`},
		{"audio_bitrate", `ALTER TABLE pretranscode_files ADD COLUMN audio_bitrate INTEGER NOT NULL DEFAULT 0`},
		{"audio_sample_rate", `ALTER TABLE pretranscode_files ADD COLUMN audio_sample_rate INTEGER NOT NULL DEFAULT 0`},
	}
	for _, c := range checks {
		exists, err := columnExists(tx, "pretranscode_files", c.column)
		if err != nil {
			return err
		}
		if !exists {
			if _, err := tx.Exec(c.ddl); err != nil {
				return err
			}
		}
	}
	return nil
}

func down029(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE pretranscode_files DROP COLUMN audio_codec;
		ALTER TABLE pretranscode_files DROP COLUMN audio_channels;
		ALTER TABLE pretranscode_files DROP COLUMN audio_bitrate;
		ALTER TABLE pretranscode_files DROP COLUMN audio_sample_rate;
	`)
	return err
}
