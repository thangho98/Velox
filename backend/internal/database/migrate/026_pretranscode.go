package migrate

import "database/sql"

// 026: Pre-transcode tables for offline encoding (Plan P)
func up026(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE pretranscode_profiles (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			name          TEXT NOT NULL,
			height        INTEGER NOT NULL,
			video_bitrate INTEGER NOT NULL,
			audio_bitrate INTEGER NOT NULL,
			video_codec   TEXT NOT NULL DEFAULT 'h264',
			audio_codec   TEXT NOT NULL DEFAULT 'aac',
			enabled       INTEGER NOT NULL DEFAULT 0,
			created_at    TEXT NOT NULL DEFAULT (datetime('now'))
		);

		INSERT INTO pretranscode_profiles (name, height, video_bitrate, audio_bitrate) VALUES
			('480p',  480,  1500, 128),
			('720p',  720,  4000, 128),
			('1080p', 1080, 8000, 192);

		CREATE TABLE pretranscode_files (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			media_file_id INTEGER NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
			profile_id    INTEGER NOT NULL REFERENCES pretranscode_profiles(id) ON DELETE CASCADE,
			file_path     TEXT NOT NULL,
			file_size     INTEGER NOT NULL DEFAULT 0,
			duration_secs REAL,
			status        TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','encoding','ready','failed')),
			error_message TEXT,
			started_at    TEXT,
			completed_at  TEXT,
			created_at    TEXT NOT NULL DEFAULT (datetime('now')),
			UNIQUE(media_file_id, profile_id)
		);

		CREATE INDEX idx_pretranscode_files_media ON pretranscode_files(media_file_id);
		CREATE INDEX idx_pretranscode_files_status ON pretranscode_files(status);

		CREATE TABLE pretranscode_queue (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			media_file_id INTEGER NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
			profile_id    INTEGER NOT NULL REFERENCES pretranscode_profiles(id) ON DELETE CASCADE,
			priority      INTEGER NOT NULL DEFAULT 0,
			status        TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','encoding','done','failed','cancelled')),
			created_at    TEXT NOT NULL DEFAULT (datetime('now')),
			UNIQUE(media_file_id, profile_id)
		);

		CREATE INDEX idx_pretranscode_queue_status ON pretranscode_queue(status);
	`)
	return err
}

func down026(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS pretranscode_queue;
		DROP TABLE IF EXISTS pretranscode_files;
		DROP TABLE IF EXISTS pretranscode_profiles;
	`)
	return err
}
