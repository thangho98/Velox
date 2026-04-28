package migrate

import "database/sql"

// 044: Persist #EXTVLCOPT headers (http-user-agent, http-referrer, http-origin)
// from M3U playlists so the Live TV proxy can forward them to upstream CDNs
// that require per-channel UA/Referer (e.g. MyTV, TV360).
func up044(tx *sql.Tx) error {
	checks := []struct {
		column string
		ddl    string
	}{
		{"user_agent", `ALTER TABLE live_channels ADD COLUMN user_agent TEXT NOT NULL DEFAULT ''`},
		{"referer", `ALTER TABLE live_channels ADD COLUMN referer TEXT NOT NULL DEFAULT ''`},
		{"origin", `ALTER TABLE live_channels ADD COLUMN origin TEXT NOT NULL DEFAULT ''`},
	}
	for _, c := range checks {
		exists, err := columnExists(tx, "live_channels", c.column)
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

func down044(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE live_channels DROP COLUMN user_agent;
		ALTER TABLE live_channels DROP COLUMN referer;
		ALTER TABLE live_channels DROP COLUMN origin;
	`)
	return err
}
