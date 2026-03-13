package migrate

import "database/sql"

// 012: Add OMDb rating columns to media table
func up012(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE media ADD COLUMN imdb_rating REAL DEFAULT 0;
		ALTER TABLE media ADD COLUMN rt_score INTEGER DEFAULT 0;
		ALTER TABLE media ADD COLUMN metacritic_score INTEGER DEFAULT 0;
	`)
	return err
}

func down012(tx *sql.Tx) error {
	// SQLite doesn't support DROP COLUMN before 3.35.0;
	// recreate table without the columns for rollback.
	_, err := tx.Exec(`
		CREATE TABLE media_new (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			library_id    INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
			media_type    TEXT NOT NULL DEFAULT 'movie' CHECK (media_type IN ('movie', 'episode')),
			title         TEXT NOT NULL,
			sort_title    TEXT DEFAULT '',
			tmdb_id       INTEGER DEFAULT NULL,
			imdb_id       TEXT DEFAULT NULL,
			overview      TEXT DEFAULT '',
			release_date  TEXT DEFAULT '',
			rating        REAL DEFAULT 0,
			poster_path   TEXT DEFAULT '',
			backdrop_path TEXT DEFAULT '',
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO media_new SELECT id, library_id, media_type, title, sort_title,
			tmdb_id, imdb_id, overview, release_date, rating, poster_path, backdrop_path,
			created_at, updated_at FROM media;
		DROP TABLE media;
		ALTER TABLE media_new RENAME TO media;
		CREATE INDEX idx_media_library ON media(library_id);
		CREATE INDEX idx_media_tmdb ON media(tmdb_id) WHERE tmdb_id IS NOT NULL;
		CREATE INDEX idx_media_imdb ON media(imdb_id) WHERE imdb_id IS NOT NULL;
		CREATE INDEX idx_media_type ON media(media_type);
		CREATE INDEX idx_media_title ON media(sort_title);
	`)
	return err
}
