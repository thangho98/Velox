package migrate

import "database/sql"

// All returns all registered migrations in order.
// New migrations should be appended here.
func All() []Migration {
	return []Migration{
		{
			Version: 1,
			Name:    "initial_schema",
			Up:      up001,
			Down:    down001,
		},
	}
}

// 001: Initial schema - baseline from existing database.
func up001(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE libraries (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT NOT NULL,
			path       TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE media (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			library_id   INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
			title        TEXT NOT NULL,
			file_path    TEXT NOT NULL UNIQUE,
			duration     REAL DEFAULT 0,
			size         INTEGER DEFAULT 0,
			width        INTEGER DEFAULT 0,
			height       INTEGER DEFAULT 0,
			video_codec  TEXT DEFAULT '',
			audio_codec  TEXT DEFAULT '',
			container    TEXT DEFAULT '',
			bitrate      INTEGER DEFAULT 0,
			has_subtitle INTEGER DEFAULT 0,
			poster_path  TEXT DEFAULT '',
			created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE progress (
			media_id   INTEGER PRIMARY KEY REFERENCES media(id) ON DELETE CASCADE,
			position   REAL DEFAULT 0,
			completed  INTEGER DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX idx_media_library ON media(library_id);
	`)
	return err
}

func down001(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS progress;
		DROP TABLE IF EXISTS media;
		DROP TABLE IF EXISTS libraries;
	`)
	return err
}
