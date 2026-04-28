package migrate

import "database/sql"

func up040(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE live_playlists (
			id                  INTEGER PRIMARY KEY AUTOINCREMENT,
			name                TEXT NOT NULL,
			url                 TEXT NOT NULL,
			last_synced_at      DATETIME,
			sync_interval_hours INTEGER NOT NULL DEFAULT 24,
			is_active           INTEGER NOT NULL DEFAULT 1,
			created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE live_channels (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			playlist_id INTEGER NOT NULL REFERENCES live_playlists(id) ON DELETE CASCADE,
			channel_id  TEXT DEFAULT '',
			name        TEXT NOT NULL,
			logo        TEXT DEFAULT '',
			group_title TEXT DEFAULT '',
			stream_url  TEXT NOT NULL,
			country     TEXT DEFAULT '',
			is_active   INTEGER NOT NULL DEFAULT 1,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX idx_live_channels_playlist ON live_channels(playlist_id);
		CREATE INDEX idx_live_channels_group ON live_channels(group_title);
		CREATE UNIQUE INDEX idx_live_channels_uniq ON live_channels(playlist_id, name);
	`)
	return err
}

func down040(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS live_channels;
		DROP TABLE IF EXISTS live_playlists;
	`)
	return err
}
