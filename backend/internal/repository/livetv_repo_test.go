package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/thawng/velox/internal/model"
)

func setupLiveTVTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE live_playlists (
			id                  INTEGER PRIMARY KEY AUTOINCREMENT,
			name                TEXT NOT NULL,
			url                 TEXT NOT NULL,
			last_synced_at      DATETIME,
			sync_interval_hours INTEGER NOT NULL DEFAULT 24,
			is_active           INTEGER NOT NULL DEFAULT 1,
			epg_url             TEXT DEFAULT '',
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
			user_agent  TEXT NOT NULL DEFAULT '',
			referer     TEXT NOT NULL DEFAULT '',
			origin      TEXT NOT NULL DEFAULT '',
			is_active   INTEGER NOT NULL DEFAULT 1,
			is_hidden   INTEGER NOT NULL DEFAULT 0,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX idx_live_channels_playlist ON live_channels(playlist_id);
		CREATE INDEX idx_live_channels_group ON live_channels(group_title);
		CREATE UNIQUE INDEX idx_live_channels_uniq ON live_channels(playlist_id, name);
	`)
	if err != nil {
		t.Fatalf("failed to create test schema: %v", err)
	}

	return db
}

func TestLiveTVRepo_UpdateDeleteReturnsErrNotFound(t *testing.T) {
	db := setupLiveTVTestDB(t)
	defer db.Close()
	ctx := context.Background()
	repo := NewLiveTVRepo(db)

	p := &model.LivePlaylist{
		ID:                999, // Non-existent
		Name:              "Missing",
		URL:               "http://missing",
		SyncIntervalHours: 24,
		IsActive:          true,
	}

	err := repo.UpdatePlaylist(ctx, p)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("UpdatePlaylist() err = %v, want ErrNotFound", err)
	}

	err = repo.DeletePlaylist(ctx, 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("DeletePlaylist() err = %v, want ErrNotFound", err)
	}
}

func TestLiveTVRepo_GetChannelsFiltersByPlaylistActive(t *testing.T) {
	db := setupLiveTVTestDB(t)
	defer db.Close()
	ctx := context.Background()
	repo := NewLiveTVRepo(db)

	_, err := db.Exec(`
		INSERT INTO live_playlists (id, name, url, is_active) VALUES 
		(1, 'Active Playlist', 'http://active', 1),
		(2, 'Inactive Playlist', 'http://inactive', 0);

		INSERT INTO live_channels (playlist_id, name, stream_url, is_active) VALUES 
		(1, 'Active Channel 1', 'http://stream1', 1),
		(1, 'Inactive Channel 1', 'http://stream2', 0),
		(2, 'Active Channel 2', 'http://stream3', 1);
	`)
	if err != nil {
		t.Fatalf("seed data failed: %v", err)
	}

	channels, err := repo.GetChannels(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("GetChannels() error = %v", err)
	}

	if len(channels) != 1 {
		t.Fatalf("GetChannels() len = %d, want 1", len(channels))
	}

	if channels[0].Name != "Active Channel 1" {
		t.Errorf("GetChannels() channel[0].Name = %q, want 'Active Channel 1'", channels[0].Name)
	}
}

func TestLiveTVRepo_SyncChannelsTransaction(t *testing.T) {
	db := setupLiveTVTestDB(t)
	defer db.Close()
	ctx := context.Background()
	repo := NewLiveTVRepo(db)

	_, err := db.Exec(`
		INSERT INTO live_playlists (id, name, url, is_active) VALUES 
		(1, 'Test Playlist', 'http://test', 1);

		INSERT INTO live_channels (playlist_id, name, stream_url, is_active, updated_at) VALUES 
		(1, 'Old Channel', 'http://old', 1, '2020-01-01 00:00:00'),
		(1, 'Keep Channel', 'http://keep-old', 1, '2020-01-01 00:00:00');
	`)
	if err != nil {
		t.Fatalf("seed data failed: %v", err)
	}

	// Sync with a new channel list
	// "Keep Channel" should be updated and remain active
	// "Old Channel" should become inactive
	// "New Channel" should be inserted and active
	newChannels := []*model.LiveChannel{
		{
			PlaylistID: 1,
			Name:       "Keep Channel",
			StreamURL:  "http://keep-new",
			IsActive:   true,
		},
		{
			PlaylistID: 1,
			Name:       "New Channel",
			StreamURL:  "http://new",
			IsActive:   true,
		},
	}

	err = repo.SyncChannelsTransaction(ctx, 1, newChannels)
	if err != nil {
		t.Fatalf("SyncChannelsTransaction() error = %v", err)
	}

	// Verify the database state including updated_at
	rows, err := db.Query(`SELECT name, stream_url, is_active, updated_at FROM live_channels ORDER BY name`)
	if err != nil {
		t.Fatalf("db query error: %v", err)
	}
	defer rows.Close()

	type ch struct {
		Name      string
		StreamURL string
		IsActive  int
		UpdatedAt string
	}
	var results []ch
	for rows.Next() {
		var c ch
		if err := rows.Scan(&c.Name, &c.StreamURL, &c.IsActive, &c.UpdatedAt); err != nil {
			t.Fatalf("scan error: %v", err)
		}
		results = append(results, c)
	}

	if len(results) != 3 {
		t.Fatalf("db channels count = %d, want 3", len(results))
	}

	// Check "Keep Channel"
	if results[0].Name != "Keep Channel" || results[0].StreamURL != "http://keep-new" || results[0].IsActive != 1 {
		t.Errorf("Keep Channel wrong state: %+v", results[0])
	}
	if results[0].UpdatedAt == "2020-01-01 00:00:00" {
		t.Errorf("Keep Channel updated_at did not change, still %q", results[0].UpdatedAt)
	}

	// Check "New Channel"
	if results[1].Name != "New Channel" || results[1].StreamURL != "http://new" || results[1].IsActive != 1 {
		t.Errorf("New Channel wrong state: %+v", results[1])
	}
	// Check "Old Channel"
	if results[2].Name != "Old Channel" || results[2].IsActive != 0 {
		t.Errorf("Old Channel wrong state: %+v", results[2])
	}
}

func TestLiveTVRepo_SyncChannelsTransaction_Rollback(t *testing.T) {
	db := setupLiveTVTestDB(t)
	defer db.Close()
	ctx := context.Background()
	repo := NewLiveTVRepo(db)

	_, err := db.Exec(`
		INSERT INTO live_playlists (id, name, url, is_active) VALUES 
		(1, 'Test Playlist', 'http://test', 1);

		INSERT INTO live_channels (playlist_id, name, stream_url, is_active) VALUES 
		(1, 'Channel A', 'http://a', 1);

		-- Add a trigger to force an error during insert
		CREATE TRIGGER force_insert_fail 
		BEFORE INSERT ON live_channels 
		WHEN NEW.name = 'Fail Channel' 
		BEGIN 
			SELECT RAISE(ABORT, 'forced failure'); 
		END;
	`)
	if err != nil {
		t.Fatalf("seed data failed: %v", err)
	}

	newChannels := []*model.LiveChannel{
		{
			PlaylistID: 1,
			Name:       "Channel A",
			StreamURL:  "http://a-new",
			IsActive:   true,
		},
		{
			PlaylistID: 1,
			Name:       "Fail Channel",
			StreamURL:  "http://fail",
			IsActive:   true,
		},
	}

	err = repo.SyncChannelsTransaction(ctx, 1, newChannels)
	if err == nil {
		t.Fatalf("expected error due to forced failure, got nil")
	}

	// Because of rollback, "Channel A" should remain active and unchanged
	var isActive int
	var streamURL string
	err = db.QueryRow(`SELECT is_active, stream_url FROM live_channels WHERE name = 'Channel A'`).Scan(&isActive, &streamURL)
	if err != nil {
		t.Fatalf("failed to query Channel A: %v", err)
	}

	if isActive != 1 {
		t.Errorf("Channel A should still be active due to rollback, got %d", isActive)
	}
	if streamURL != "http://a" {
		t.Errorf("Channel A URL should not have changed, got %s", streamURL)
	}
}

func TestLiveTVRepo_GetChannelGroupsAndSearch_FiltersByPlaylistActive(t *testing.T) {
	db := setupLiveTVTestDB(t)
	defer db.Close()
	ctx := context.Background()
	repo := NewLiveTVRepo(db)

	_, err := db.Exec(`
		INSERT INTO live_playlists (id, name, url, is_active) VALUES 
		(1, 'Active Playlist', 'http://active', 1),
		(2, 'Inactive Playlist', 'http://inactive', 0);

		INSERT INTO live_channels (playlist_id, name, stream_url, group_title, is_active) VALUES 
		(1, 'Target One', 'http://stream1', 'Sports', 1),
		(2, 'Target Two', 'http://stream2', 'News', 1);
	`)
	if err != nil {
		t.Fatalf("seed data failed: %v", err)
	}

	// Test GetChannelGroups
	groups, err := repo.GetChannelGroups(ctx)
	if err != nil {
		t.Fatalf("GetChannelGroups() error = %v", err)
	}

	if len(groups) != 1 || groups[0] != "Sports" {
		t.Errorf("GetChannelGroups() expected ['Sports'], got %v", groups)
	}

	// Test SearchChannels
	channels, err := repo.SearchChannels(ctx, "Target", 10)
	if err != nil {
		t.Fatalf("SearchChannels() error = %v", err)
	}

	if len(channels) != 1 || channels[0].Name != "Target One" {
		t.Errorf("SearchChannels() expected only 'Target One', got %d results", len(channels))
	}
}
