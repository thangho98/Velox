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
		{
			Version: 2,
			Name:    "refactor_media_item",
			Up:      up002,
			Down:    down002,
		},
		{
			Version: 3,
			Name:    "series_model",
			Up:      up003,
			Down:    down003,
		},
		{
			Version: 4,
			Name:    "genres_people",
			Up:      up004,
			Down:    down004,
		},
		{
			Version: 5,
			Name:    "scan_jobs",
			Up:      up005,
			Down:    down005,
		},
		{
			Version: 6,
			Name:    "subtitles_audio_tracks",
			Up:      up006,
			Down:    down006,
		},
		{
			Version: 7,
			Name:    "users",
			Up:      up007,
			Down:    down007,
		},
		{
			Version: 8,
			Name:    "refresh_tokens_sessions",
			Up:      up008,
			Down:    down008,
		},
		{
			Version: 9,
			Name:    "user_data",
			Up:      up009,
			Down:    down009,
		},
		{
			Version: 10,
			Name:    "library_paths",
			Up:      up010,
			Down:    down010,
		},
		{
			Version: 11,
			Name:    "app_settings",
			Up:      up011,
			Down:    down011,
		},
		{
			Version: 12,
			Name:    "omdb_ratings",
			Up:      up012,
			Down:    down012,
		},
		{
			Version: 13,
			Name:    "tvdb_id",
			Up:      up013,
			Down:    down013,
		},
		{
			Version: 14,
			Name:    "fanart_logo",
			Up:      up014,
			Down:    down014,
		},
		{
			Version: 15,
			Name:    "series_network",
			Up:      up015,
			Down:    down015,
		},
		{
			Version: 16,
			Name:    "activity_log",
			Up:      up016,
			Down:    down016,
		},
		{
			Version: 17,
			Name:    "webhooks",
			Up:      up017,
			Down:    down017,
		},
		{
			Version: 18,
			Name:    "media_file_video_details",
			Up:      up018,
			Down:    down018,
		},
		{
			Version: 19,
			Name:    "media_markers",
			Up:      up019,
			Down:    down019,
		},
		{
			Version: 20,
			Name:    "audio_fingerprints",
			Up:      up020,
			Down:    down020,
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

// 002: Refactor media - separate logical media from physical files
func up002(tx *sql.Tx) error {
	_, err := tx.Exec(`
		-- Add type column to libraries
		ALTER TABLE libraries ADD COLUMN type TEXT DEFAULT 'mixed';

		-- Save old media data temporarily
		CREATE TABLE _media_old AS SELECT * FROM media;
		DROP TABLE progress;
		DROP TABLE media;

		-- Create new media table (logical item)
		CREATE TABLE media (
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

		-- Create media_files table (physical file)
		CREATE TABLE media_files (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			media_id         INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
			file_path        TEXT NOT NULL UNIQUE,
			file_size        INTEGER DEFAULT 0,
			duration         REAL DEFAULT 0,
			width            INTEGER DEFAULT 0,
			height           INTEGER DEFAULT 0,
			video_codec      TEXT DEFAULT '',
			audio_codec      TEXT DEFAULT '',
			container        TEXT DEFAULT '',
			bitrate          INTEGER DEFAULT 0,
			fingerprint      TEXT DEFAULT '',
			is_primary       INTEGER DEFAULT 1,
			added_at         DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_verified_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		-- Indexes for media
		CREATE INDEX idx_media_library ON media(library_id);
		CREATE INDEX idx_media_tmdb ON media(tmdb_id) WHERE tmdb_id IS NOT NULL;
		CREATE INDEX idx_media_imdb ON media(imdb_id) WHERE imdb_id IS NOT NULL;
		CREATE INDEX idx_media_type ON media(media_type);
		CREATE INDEX idx_media_title ON media(sort_title);

		-- Indexes for media_files
		CREATE INDEX idx_mf_media ON media_files(media_id);
		CREATE INDEX idx_mf_fingerprint ON media_files(fingerprint);

		-- Drop the backup table
		DROP TABLE IF EXISTS _media_old;
	`)
	return err
}

func down002(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS media_files;
		DROP TABLE IF EXISTS media;

		-- Recreate old media table
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

		-- Remove type column from libraries (SQLite < 3.35.0 doesn't support DROP COLUMN)
		CREATE TABLE libraries_new (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT NOT NULL,
			path       TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO libraries_new SELECT id, name, path, created_at FROM libraries;
		DROP TABLE libraries;
		ALTER TABLE libraries_new RENAME TO libraries;
	`)
	return err
}

// 003: Series model (TV shows)
func up003(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE series (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			library_id     INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
			title          TEXT NOT NULL,
			sort_title     TEXT DEFAULT '',
			tmdb_id        INTEGER UNIQUE DEFAULT NULL,
			imdb_id        TEXT DEFAULT NULL,
			overview       TEXT DEFAULT '',
			status         TEXT DEFAULT '',
			first_air_date TEXT DEFAULT '',
			poster_path    TEXT DEFAULT '',
			backdrop_path  TEXT DEFAULT '',
			created_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE seasons (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			series_id      INTEGER NOT NULL REFERENCES series(id) ON DELETE CASCADE,
			season_number  INTEGER NOT NULL,
			title          TEXT DEFAULT '',
			overview       TEXT DEFAULT '',
			poster_path    TEXT DEFAULT '',
			episode_count  INTEGER DEFAULT 0,
			created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE episodes (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			series_id      INTEGER NOT NULL REFERENCES series(id) ON DELETE CASCADE,
			season_id      INTEGER NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
			media_id       INTEGER UNIQUE NOT NULL REFERENCES media(id) ON DELETE CASCADE,
			episode_number INTEGER NOT NULL,
			title          TEXT DEFAULT '',
			overview       TEXT DEFAULT '',
			still_path     TEXT DEFAULT '',
			air_date       TEXT DEFAULT '',
			created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		-- Indexes for series
		CREATE INDEX idx_series_library ON series(library_id);
		CREATE INDEX idx_series_tmdb ON series(tmdb_id) WHERE tmdb_id IS NOT NULL;

		-- Indexes for seasons
		CREATE INDEX idx_seasons_series ON seasons(series_id);
		CREATE UNIQUE INDEX idx_seasons_number ON seasons(series_id, season_number);

		-- Indexes for episodes
		CREATE INDEX idx_ep_series ON episodes(series_id);
		CREATE INDEX idx_ep_season ON episodes(season_id);
		CREATE UNIQUE INDEX idx_ep_number ON episodes(season_id, episode_number);
	`)
	return err
}

func down003(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS episodes;
		DROP TABLE IF EXISTS seasons;
		DROP TABLE IF EXISTS series;
	`)
	return err
}

// 004: Genres and People (credits)
func up004(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE genres (
			id      INTEGER PRIMARY KEY AUTOINCREMENT,
			name    TEXT NOT NULL UNIQUE,
			tmdb_id INTEGER UNIQUE DEFAULT NULL
		);

		CREATE TABLE media_genres (
			media_id  INTEGER DEFAULT NULL REFERENCES media(id) ON DELETE CASCADE,
			series_id INTEGER DEFAULT NULL REFERENCES series(id) ON DELETE CASCADE,
			genre_id  INTEGER NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
			CONSTRAINT xor_owner CHECK (
				(media_id IS NOT NULL AND series_id IS NULL) OR
				(media_id IS NULL AND series_id IS NOT NULL)
			),
			CONSTRAINT unique_media_genre UNIQUE (media_id, genre_id) ON CONFLICT REPLACE
		);
		CREATE UNIQUE INDEX idx_series_genre ON media_genres(series_id, genre_id) WHERE series_id IS NOT NULL;

		CREATE TABLE people (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			name         TEXT NOT NULL,
			tmdb_id      INTEGER UNIQUE DEFAULT NULL,
			profile_path TEXT DEFAULT ''
		);

		CREATE TABLE credits (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			media_id       INTEGER DEFAULT NULL REFERENCES media(id) ON DELETE CASCADE,
			series_id      INTEGER DEFAULT NULL REFERENCES series(id) ON DELETE CASCADE,
			person_id      INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
			character      TEXT DEFAULT '',
			role           TEXT NOT NULL CHECK (role IN ('cast', 'director', 'writer')),
			display_order  INTEGER DEFAULT 0,
			CONSTRAINT xor_credit_owner CHECK (
				(media_id IS NOT NULL AND series_id IS NULL) OR
				(media_id IS NULL AND series_id IS NOT NULL)
			)
		);

		-- Indexes
		CREATE INDEX idx_credits_media ON credits(media_id) WHERE media_id IS NOT NULL;
		CREATE INDEX idx_credits_series ON credits(series_id) WHERE series_id IS NOT NULL;
		CREATE INDEX idx_credits_person ON credits(person_id);
	`)
	return err
}

func down004(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS credits;
		DROP TABLE IF EXISTS people;
		DROP TABLE IF EXISTS media_genres;
		DROP TABLE IF EXISTS genres;
	`)
	return err
}

// 005: Scan Jobs tracking
func up005(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE scan_jobs (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			library_id     INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
			status         TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'scanning', 'completed', 'failed')),
			total_files    INTEGER DEFAULT 0,
			scanned_files  INTEGER DEFAULT 0,
			new_files      INTEGER DEFAULT 0,
			errors         INTEGER DEFAULT 0,
			error_log      TEXT DEFAULT '',
			started_at     DATETIME DEFAULT NULL,
			finished_at    DATETIME DEFAULT NULL,
			created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX idx_scanjob_library ON scan_jobs(library_id);
		CREATE INDEX idx_scanjob_status ON scan_jobs(status);
	`)
	return err
}

func down005(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS scan_jobs;`)
	return err
}

// 006: Subtitles and Audio Tracks
func up006(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE subtitles (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			media_file_id   INTEGER NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
			language        TEXT DEFAULT '',
			codec           TEXT DEFAULT '',
			title           TEXT DEFAULT '',
			is_embedded     INTEGER DEFAULT 1,
			stream_index    INTEGER DEFAULT -1,
			file_path       TEXT DEFAULT '',
			is_forced       INTEGER DEFAULT 0,
			is_default      INTEGER DEFAULT 0,
			is_sdh          INTEGER DEFAULT 0
		);

		CREATE INDEX idx_sub_mediafile ON subtitles(media_file_id);
		CREATE INDEX idx_sub_lang ON subtitles(language);
		CREATE UNIQUE INDEX idx_sub_embedded ON subtitles(media_file_id, stream_index) WHERE is_embedded = 1;
		CREATE UNIQUE INDEX idx_sub_external ON subtitles(media_file_id, file_path) WHERE is_embedded = 0;

		CREATE TABLE audio_tracks (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			media_file_id   INTEGER NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
			stream_index    INTEGER NOT NULL,
			codec           TEXT DEFAULT '',
			language        TEXT DEFAULT '',
			channels        INTEGER DEFAULT 2,
			channel_layout  TEXT DEFAULT '',
			bitrate         INTEGER DEFAULT 0,
			title           TEXT DEFAULT '',
			is_default      INTEGER DEFAULT 0,
			UNIQUE(media_file_id, stream_index)
		);

		CREATE INDEX idx_audio_mediafile ON audio_tracks(media_file_id);
	`)
	return err
}

func down006(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS audio_tracks;
		DROP TABLE IF EXISTS subtitles;
	`)
	return err
}

// 007: Users and permissions
func up007(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE users (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			username      TEXT NOT NULL UNIQUE,
			display_name  TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			is_admin      INTEGER DEFAULT 0,
			avatar_path   TEXT DEFAULT '',
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE user_preferences (
			user_id                INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			subtitle_language      TEXT DEFAULT '',
			audio_language         TEXT DEFAULT '',
			max_streaming_quality  TEXT DEFAULT 'auto',
			theme                  TEXT DEFAULT 'dark'
		);

		CREATE TABLE user_library_access (
			user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			library_id INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
			PRIMARY KEY (user_id, library_id)
		);
	`)
	return err
}

func down007(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS user_library_access;
		DROP TABLE IF EXISTS user_preferences;
		DROP TABLE IF EXISTS users;
	`)
	return err
}

// 008: Refresh tokens and sessions
func up008(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE refresh_tokens (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash  TEXT NOT NULL UNIQUE,
			device_name TEXT DEFAULT '',
			ip_address  TEXT DEFAULT '',
			expires_at  DATETIME NOT NULL,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_rt_user ON refresh_tokens(user_id);
		CREATE INDEX idx_rt_expires ON refresh_tokens(expires_at);

		CREATE TABLE sessions (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id          INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			refresh_token_id INTEGER REFERENCES refresh_tokens(id) ON DELETE SET NULL,
			device_name      TEXT DEFAULT '',
			ip_address       TEXT DEFAULT '',
			user_agent       TEXT DEFAULT '',
			expires_at       DATETIME NOT NULL,
			last_active_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_at       DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_sessions_user ON sessions(user_id);
		CREATE INDEX idx_sessions_expires ON sessions(expires_at);
	`)
	return err
}

func down008(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS sessions;
		DROP TABLE IF EXISTS refresh_tokens;
	`)
	return err
}

// 009: User data (progress, favorites, ratings) - unified per-user-per-media pattern
func up009(tx *sql.Tx) error {
	_, err := tx.Exec(`
		-- Drop old progress table (from migration 001)
		DROP TABLE IF EXISTS progress;

		-- Unified per-user-per-media state (Emby pattern: 1 row = 1 user-media pair)
		CREATE TABLE user_data (
			user_id        INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			media_id       INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
			position       REAL DEFAULT 0,
			completed      INTEGER DEFAULT 0,
			is_favorite    INTEGER DEFAULT 0,
			rating         REAL DEFAULT NULL CHECK (rating IS NULL OR (rating >= 1.0 AND rating <= 10.0)),
			play_count     INTEGER DEFAULT 0,
			last_played_at DATETIME DEFAULT NULL,
			updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, media_id)
		);
		CREATE INDEX idx_ud_user ON user_data(user_id);
		CREATE INDEX idx_ud_media ON user_data(media_id);
		CREATE INDEX idx_ud_favorite ON user_data(user_id) WHERE is_favorite = 1;
		CREATE INDEX idx_ud_recent ON user_data(user_id, last_played_at DESC) WHERE last_played_at IS NOT NULL;

		-- Series-level favorite/rating
		CREATE TABLE user_series_data (
			user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			series_id   INTEGER NOT NULL REFERENCES series(id) ON DELETE CASCADE,
			is_favorite INTEGER DEFAULT 0,
			rating      REAL DEFAULT NULL CHECK (rating IS NULL OR (rating >= 1.0 AND rating <= 10.0)),
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, series_id)
		);
	`)
	return err
}

func down009(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS user_series_data;
		DROP TABLE IF EXISTS user_data;
	`)
	return err
}

// 010: Library paths — support multiple folders per library
func up010(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE library_paths (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			library_id INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
			path       TEXT NOT NULL UNIQUE
		);

		CREATE INDEX idx_library_paths ON library_paths(library_id);

		-- Migrate existing single-path libraries
		INSERT INTO library_paths (library_id, path)
		SELECT id, path FROM libraries WHERE path != '';
	`)
	return err
}

func down010(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS library_paths;`)
	return err
}

// 018: Add video details (fps, profile, level) to media_files and sample_rate to audio_tracks.
func up018(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE media_files ADD COLUMN video_profile TEXT DEFAULT '';
		ALTER TABLE media_files ADD COLUMN video_level INTEGER DEFAULT 0;
		ALTER TABLE media_files ADD COLUMN video_fps REAL DEFAULT 0;
		ALTER TABLE audio_tracks ADD COLUMN sample_rate INTEGER DEFAULT 0;
	`)
	return err
}

func down018(tx *sql.Tx) error {
	// SQLite doesn't support DROP COLUMN before 3.35.0; recreate would be needed.
	// For simplicity, these columns are harmless if left.
	return nil
}

// 019: Media markers for intro/credits skip (chapter-based, fingerprint-based, or manual)
func up019(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE media_markers (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			media_file_id   INTEGER NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
			marker_type     TEXT NOT NULL CHECK (marker_type IN ('intro','credits')),
			start_sec       REAL NOT NULL,
			end_sec         REAL NOT NULL,
			source          TEXT NOT NULL CHECK (source IN ('chapter','fingerprint','manual')),
			confidence      REAL NOT NULL DEFAULT 1.0,
			label           TEXT NOT NULL DEFAULT '',
			created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,

			-- Validation: start must be >= 0, end must be > start
			CONSTRAINT valid_start CHECK (start_sec >= 0),
			CONSTRAINT valid_end CHECK (end_sec > start_sec)
		);

		-- Unique constraint: same file + type + source + exact same segment
		CREATE UNIQUE INDEX idx_marker_unique
			ON media_markers(media_file_id, marker_type, source, start_sec, end_sec);

		-- Lookup index for querying markers by file and type
		CREATE INDEX idx_marker_lookup
			ON media_markers(media_file_id, marker_type);

		-- Index for source-based queries (e.g., delete all chapter markers for rebuild)
		CREATE INDEX idx_marker_source
			ON media_markers(media_file_id, source);
	`)
	return err
}

func down019(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS media_markers;`)
	return err
}

// 020: Audio fingerprints cache for chromaprint-based intro/credits detection.
func up020(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE audio_fingerprints (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			media_file_id   INTEGER NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
			region          TEXT NOT NULL CHECK (region IN ('intro_region','credits_region')),
			fingerprint     BLOB NOT NULL,
			duration_sec    REAL NOT NULL,
			sample_count    INTEGER NOT NULL DEFAULT 0,
			created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(media_file_id, region)
		);
		CREATE INDEX idx_afp_mediafile ON audio_fingerprints(media_file_id);
	`)
	return err
}

func down020(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS audio_fingerprints;`)
	return err
}
