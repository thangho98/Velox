package migrate

import "database/sql"

// 045: Per-user recently-watched tracking for live TV channels.
// Mirrors user_data (for media) but pointed at live_channels. Populated
// automatically by the live TV proxy when it can identify the requester.
func up045(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS user_live_channel_data (
			user_id        INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			channel_id     INTEGER NOT NULL REFERENCES live_channels(id) ON DELETE CASCADE,
			play_count     INTEGER NOT NULL DEFAULT 0,
			last_watched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, channel_id)
		);
		CREATE INDEX IF NOT EXISTS idx_ulcd_recent ON user_live_channel_data(user_id, last_watched_at DESC);
	`)
	return err
}

func down045(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS user_live_channel_data;`)
	return err
}
