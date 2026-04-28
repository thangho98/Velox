package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thawng/velox/internal/repository"
)

func setupLiveTVTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", "file::memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	schema := `
	CREATE TABLE live_playlists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		epg_url TEXT NOT NULL DEFAULT '',
		last_synced_at DATETIME,
		sync_interval_hours INTEGER NOT NULL DEFAULT 24,
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE live_channels (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		playlist_id INTEGER NOT NULL REFERENCES live_playlists(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		logo TEXT NOT NULL DEFAULT '',
		stream_url TEXT NOT NULL,
		epg_channel_id TEXT NOT NULL DEFAULT '',
		category TEXT NOT NULL DEFAULT '',
		is_hidden INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestLiveTVService_New(t *testing.T) {
	t.Parallel()
	db := setupLiveTVTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewLiveTVRepo(db)
	svc := NewLiveTVService(repo, nil)

	assert.NotNil(t, svc)
}

func TestLiveTVService_SyncAllPlaylists_Empty(t *testing.T) {
	t.Parallel()
	db := setupLiveTVTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewLiveTVRepo(db)
	svc := NewLiveTVService(repo, nil)

	err := svc.SyncAllPlaylists(context.Background())
	assert.NoError(t, err)
}

func TestLiveTVService_SyncPlaylist_NotFound(t *testing.T) {
	t.Parallel()
	db := setupLiveTVTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewLiveTVRepo(db)
	svc := NewLiveTVService(repo, nil)

	err := svc.SyncPlaylist(context.Background(), 9999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLiveTVService_ToggleChannelHidden(t *testing.T) {
	t.Parallel()
	db := setupLiveTVTestDB(t)
	t.Cleanup(func() { db.Close() })

	// Insert test playlist and channel
	_, err := db.Exec(`
		INSERT INTO live_playlists (id, name, url, is_active) VALUES (1, 'Test', 'http://test.m3u', 1);
		INSERT INTO live_channels (id, playlist_id, name, stream_url, is_hidden) VALUES (1, 1, 'Channel 1', 'http://stream', 0);
	`)
	require.NoError(t, err)

	repo := repository.NewLiveTVRepo(db)
	svc := NewLiveTVService(repo, nil)

	// Toggle from hidden=0 to hidden=1
	hidden, err := svc.ToggleChannelHidden(context.Background(), 1)
	require.NoError(t, err)
	assert.True(t, hidden)

	// Toggle back to hidden=0
	hidden, err = svc.ToggleChannelHidden(context.Background(), 1)
	require.NoError(t, err)
	assert.False(t, hidden)
}

func TestLiveTVService_ToggleChannelHidden_NotFound(t *testing.T) {
	t.Parallel()
	db := setupLiveTVTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewLiveTVRepo(db)
	svc := NewLiveTVService(repo, nil)

	_, err := svc.ToggleChannelHidden(context.Background(), 9999)
	assert.True(t, errors.Is(err, repository.ErrNotFound))
}
