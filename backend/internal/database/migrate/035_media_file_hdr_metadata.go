package migrate

import (
	"database/sql"
)

// 035: Persist HDR/DV metadata on media_files to avoid re-probing at stream/transcode time.
// Scanner extracts these from ffprobe side_data and color tags; runtime reads from DB.
func up035(tx *sql.Tx) error {
	checks := []struct {
		table  string
		column string
		ddl    string
	}{
		{"media_files", "is_hdr", `ALTER TABLE media_files ADD COLUMN is_hdr INTEGER NOT NULL DEFAULT 0`},
		{"media_files", "dv_profile", `ALTER TABLE media_files ADD COLUMN dv_profile INTEGER NOT NULL DEFAULT 0`},
		{"media_files", "color_transfer", `ALTER TABLE media_files ADD COLUMN color_transfer TEXT NOT NULL DEFAULT ''`},
		{"media_files", "color_primaries", `ALTER TABLE media_files ADD COLUMN color_primaries TEXT NOT NULL DEFAULT ''`},
	}
	for _, c := range checks {
		exists, err := columnExists(tx, c.table, c.column)
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

func down035(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE media_files DROP COLUMN is_hdr;
		ALTER TABLE media_files DROP COLUMN dv_profile;
		ALTER TABLE media_files DROP COLUMN color_transfer;
		ALTER TABLE media_files DROP COLUMN color_primaries;
	`)
	return err
}
