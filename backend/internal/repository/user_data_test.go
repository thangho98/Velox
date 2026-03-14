package repository

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupUserDataTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE media (
			id            INTEGER PRIMARY KEY,
			library_id    INTEGER NOT NULL DEFAULT 1,
			media_type    TEXT NOT NULL,
			title         TEXT NOT NULL,
			sort_title    TEXT NOT NULL DEFAULT '',
			overview      TEXT NOT NULL DEFAULT '',
			release_date  TEXT NOT NULL DEFAULT '',
			rating        REAL NOT NULL DEFAULT 0,
			imdb_rating   REAL NOT NULL DEFAULT 0,
			rt_score      INTEGER NOT NULL DEFAULT 0,
			metacritic_score INTEGER NOT NULL DEFAULT 0,
			poster_path   TEXT NOT NULL DEFAULT '',
			backdrop_path TEXT NOT NULL DEFAULT '',
			logo_path     TEXT NOT NULL DEFAULT '',
			thumb_path    TEXT NOT NULL DEFAULT '',
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE media_files (
			id             INTEGER PRIMARY KEY,
			media_id        INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
			file_path       TEXT NOT NULL DEFAULT '',
			file_size       INTEGER NOT NULL DEFAULT 0,
			duration        REAL NOT NULL DEFAULT 0,
			width           INTEGER NOT NULL DEFAULT 0,
			height          INTEGER NOT NULL DEFAULT 0,
			video_codec     TEXT NOT NULL DEFAULT '',
			audio_codec     TEXT NOT NULL DEFAULT '',
			container       TEXT NOT NULL DEFAULT '',
			bitrate         INTEGER NOT NULL DEFAULT 0,
			fingerprint     TEXT NOT NULL DEFAULT '',
			is_primary      INTEGER NOT NULL DEFAULT 1,
			added_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_verified_at DATETIME DEFAULT NULL
		);

		CREATE TABLE series (
			id             INTEGER PRIMARY KEY,
			library_id     INTEGER NOT NULL DEFAULT 1,
			title          TEXT NOT NULL,
			sort_title     TEXT NOT NULL DEFAULT '',
			overview       TEXT NOT NULL DEFAULT '',
			status         TEXT NOT NULL DEFAULT '',
			network        TEXT NOT NULL DEFAULT '',
			first_air_date TEXT NOT NULL DEFAULT '',
			poster_path    TEXT NOT NULL DEFAULT '',
			backdrop_path  TEXT NOT NULL DEFAULT '',
			logo_path      TEXT NOT NULL DEFAULT '',
			thumb_path     TEXT NOT NULL DEFAULT '',
			created_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE seasons (
			id            INTEGER PRIMARY KEY,
			series_id     INTEGER NOT NULL REFERENCES series(id) ON DELETE CASCADE,
			season_number INTEGER NOT NULL,
			title         TEXT NOT NULL DEFAULT '',
			overview      TEXT NOT NULL DEFAULT '',
			poster_path   TEXT NOT NULL DEFAULT '',
			episode_count INTEGER NOT NULL DEFAULT 0,
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE episodes (
			id             INTEGER PRIMARY KEY,
			series_id      INTEGER NOT NULL REFERENCES series(id) ON DELETE CASCADE,
			season_id      INTEGER NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
			media_id       INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
			episode_number INTEGER NOT NULL,
			title          TEXT NOT NULL DEFAULT '',
			overview       TEXT NOT NULL DEFAULT '',
			still_path     TEXT NOT NULL DEFAULT '',
			air_date       TEXT NOT NULL DEFAULT '',
			created_at     DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE user_data (
			user_id        INTEGER NOT NULL,
			media_id       INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
			position       REAL DEFAULT 0,
			completed      INTEGER DEFAULT 0,
			is_favorite    INTEGER DEFAULT 0,
			rating         REAL DEFAULT NULL,
			play_count     INTEGER DEFAULT 0,
			last_played_at DATETIME DEFAULT NULL,
			updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, media_id)
		);
	`)
	if err != nil {
		t.Fatalf("failed to create test schema: %v", err)
	}

	return db
}

func TestUserDataRepo_ListContinueWatchingIncludesSeriesContext(t *testing.T) {
	db := setupUserDataTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := NewUserDataRepo(db)

	_, err := db.ExecContext(ctx, `
		INSERT INTO media (id, media_type, title, poster_path, backdrop_path) VALUES
			(1, 'movie', 'Movie One', '/movie.jpg', '/movie-bg.jpg'),
			(2, 'episode', 'Pilot', '/pilot.jpg', '/pilot-bg.jpg'),
			(3, 'episode', 'Completed Episode', '/done.jpg', '/done-bg.jpg');

		INSERT INTO media_files (id, media_id, duration, is_primary) VALUES
			(11, 1, 7200, 1),
			(12, 2, 1800, 1),
			(13, 3, 1800, 1);

		INSERT INTO series (id, title, poster_path) VALUES (10, 'The Show', '/series.jpg');
		INSERT INTO seasons (id, series_id, season_number, title) VALUES (20, 10, 1, 'Season 1');
		INSERT INTO episodes (id, series_id, season_id, media_id, episode_number, title) VALUES
			(30, 10, 20, 2, 1, 'Pilot'),
			(31, 10, 20, 3, 2, 'Completed Episode');

		INSERT INTO user_data (user_id, media_id, position, completed, last_played_at) VALUES
			(7, 1, 600, 0, '2026-03-13 10:00:00'),
			(7, 2, 300, 0, '2026-03-13 11:00:00'),
			(7, 3, 1800, 1, '2026-03-13 12:00:00');
	`)
	if err != nil {
		t.Fatalf("failed to seed data: %v", err)
	}

	items, err := repo.ListContinueWatching(ctx, 7, 10)
	if err != nil {
		t.Fatalf("ListContinueWatching() error = %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("ListContinueWatching() len = %d, want 2", len(items))
	}

	if got := items[0].MediaID; got != 2 {
		t.Fatalf("first item media_id = %d, want 2", got)
	}
	if got := items[0].SeriesID; got != 10 {
		t.Fatalf("first item series_id = %d, want 10", got)
	}
	if got := items[0].SeriesTitle; got != "The Show" {
		t.Fatalf("first item series_title = %q, want %q", got, "The Show")
	}
	if got := items[0].SeasonNumber; got != 1 {
		t.Fatalf("first item season_number = %d, want 1", got)
	}
	if got := items[0].EpisodeNumber; got != 1 {
		t.Fatalf("first item episode_number = %d, want 1", got)
	}
	if got := items[1].SeriesID; got != 0 {
		t.Fatalf("movie item series_id = %d, want 0", got)
	}
}

func TestUserDataRepo_ListNextUpIncludesSeriesIDAndSkipsInProgressEpisode(t *testing.T) {
	db := setupUserDataTestDB(t)
	defer db.Close()

	ctx := context.Background()
	repo := NewUserDataRepo(db)

	_, err := db.ExecContext(ctx, `
		INSERT INTO media (id, media_type, title, poster_path, backdrop_path) VALUES
			(101, 'episode', 'Episode 1', '/ep1.jpg', '/ep1-bg.jpg'),
			(102, 'episode', 'Episode 2', '/ep2.jpg', '/ep2-bg.jpg'),
			(103, 'episode', 'Episode 3', '/ep3.jpg', '/ep3-bg.jpg');

		INSERT INTO media_files (id, media_id, duration, is_primary) VALUES
			(201, 101, 1800, 1),
			(202, 102, 1800, 1),
			(203, 103, 1800, 1);

		INSERT INTO series (id, title, poster_path) VALUES (1000, 'Followed Show', '/series.jpg');
		INSERT INTO seasons (id, series_id, season_number, title) VALUES (2000, 1000, 1, 'Season 1');
		INSERT INTO episodes (id, series_id, season_id, media_id, episode_number, title, still_path) VALUES
			(3001, 1000, 2000, 101, 1, 'Episode 1', '/still1.jpg'),
			(3002, 1000, 2000, 102, 2, 'Episode 2', '/still2.jpg'),
			(3003, 1000, 2000, 103, 3, 'Episode 3', '/still3.jpg');

		INSERT INTO user_data (user_id, media_id, position, completed, last_played_at) VALUES
			(9, 101, 1800, 1, '2026-03-13 09:00:00'),
			(9, 102, 900, 0, '2026-03-13 10:00:00');
	`)
	if err != nil {
		t.Fatalf("failed to seed data: %v", err)
	}

	items, err := repo.ListNextUp(ctx, 9, 10)
	if err != nil {
		t.Fatalf("ListNextUp() error = %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("ListNextUp() len = %d, want 1", len(items))
	}

	item := items[0]
	if got := item.MediaID; got != 103 {
		t.Fatalf("next up media_id = %d, want 103", got)
	}
	if got := item.SeriesID; got != 1000 {
		t.Fatalf("next up series_id = %d, want 1000", got)
	}
	if got := item.EpisodeNumber; got != 3 {
		t.Fatalf("next up episode_number = %d, want 3", got)
	}
}
