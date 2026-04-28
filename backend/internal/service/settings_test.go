package service

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thawng/velox/internal/repository"
)

func setupSettingsTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", "file::memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	schema := `
	CREATE TABLE app_settings (
		key        TEXT PRIMARY KEY,
		value      TEXT NOT NULL DEFAULT '',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestSettingsService_New(t *testing.T) {
	t.Parallel()
	svc := NewSettingsService(nil, nil)
	assert.NotNil(t, svc)
}

func TestSettingsService_GetOpenSubtitles(t *testing.T) {
	t.Parallel()
	db := setupSettingsTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewAppSettingsRepo(db)
	svc := NewSettingsService(repo, nil)

	t.Run("returns_empty_when_not_set", func(t *testing.T) {
		t.Parallel()
		settings, err := svc.GetOpenSubtitles(context.Background())
		require.NoError(t, err)
		assert.Empty(t, settings.APIKey)
		assert.Empty(t, settings.Username)
		assert.False(t, settings.PasswordSet)
	})
}

func TestSettingsService_UpdateOpenSubtitles(t *testing.T) {
	t.Parallel()
	db := setupSettingsTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewAppSettingsRepo(db)
	svc := NewSettingsService(repo, nil)

	settings, err := svc.UpdateOpenSubtitles(context.Background(), "new-key", "user", "pass")
	require.NoError(t, err)
	assert.Equal(t, "new-key", settings.APIKey)
	assert.Equal(t, "user", settings.Username)
	assert.True(t, settings.PasswordSet)
}

func TestSettingsService_GetTMDb(t *testing.T) {
	t.Parallel()
	db := setupSettingsTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewAppSettingsRepo(db)
	svc := NewSettingsService(repo, nil)

	t.Run("returns_empty_api_key_when_not_set", func(t *testing.T) {
		t.Parallel()
		settings, err := svc.GetTMDb(context.Background())
		require.NoError(t, err)
		assert.Empty(t, settings.APIKey)
		assert.False(t, settings.HasBuiltin)
	})
}

func TestSettingsService_UpdateTMDb(t *testing.T) {
	t.Parallel()
	db := setupSettingsTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewAppSettingsRepo(db)
	svc := NewSettingsService(repo, nil)

	settings, err := svc.UpdateTMDb(context.Background(), "tmdb-api-key")
	require.NoError(t, err)
	assert.Equal(t, "tmdb-api-key", settings.APIKey)
	assert.False(t, settings.HasBuiltin)
}

func TestSettingsService_GetPlayback(t *testing.T) {
	t.Parallel()
	db := setupSettingsTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewAppSettingsRepo(db)
	svc := NewSettingsService(repo, nil)

	t.Run("returns_default_mode_when_not_set", func(t *testing.T) {
		t.Parallel()
		settings, err := svc.GetPlayback(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "auto", settings.PlaybackMode)
	})
}

func TestSettingsService_UpdatePlayback(t *testing.T) {
	t.Parallel()
	db := setupSettingsTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewAppSettingsRepo(db)
	svc := NewSettingsService(repo, nil)

	settings, err := svc.UpdatePlayback(context.Background(), "auto")
	require.NoError(t, err)
	assert.Equal(t, "auto", settings.PlaybackMode)
}

func TestSettingsService_GetAutoSubtitles(t *testing.T) {
	t.Parallel()
	db := setupSettingsTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewAppSettingsRepo(db)
	svc := NewSettingsService(repo, nil)

	t.Run("returns_empty_when_not_set", func(t *testing.T) {
		t.Parallel()
		settings, err := svc.GetAutoSubtitles(context.Background())
		require.NoError(t, err)
		assert.Empty(t, settings.Languages)
	})
}

func TestSettingsService_UpdateAutoSubtitles(t *testing.T) {
	t.Parallel()
	db := setupSettingsTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewAppSettingsRepo(db)
	svc := NewSettingsService(repo, nil)

	settings, err := svc.UpdateAutoSubtitles(context.Background(), "eng,vie")
	require.NoError(t, err)
	assert.Equal(t, "eng,vie", settings.Languages)
}

func TestSettingsService_GetSubdl(t *testing.T) {
	t.Parallel()
	db := setupSettingsTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewAppSettingsRepo(db)
	svc := NewSettingsService(repo, nil)

	t.Run("returns_empty_when_not_set", func(t *testing.T) {
		t.Parallel()
		settings, err := svc.GetSubdl(context.Background())
		require.NoError(t, err)
		assert.Empty(t, settings.APIKey)
	})
}

func TestSettingsService_UpdateSubdl(t *testing.T) {
	t.Parallel()
	db := setupSettingsTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewAppSettingsRepo(db)
	svc := NewSettingsService(repo, nil)

	settings, err := svc.UpdateSubdl(context.Background(), "subdl-api-key")
	require.NoError(t, err)
	assert.Equal(t, "subdl-api-key", settings.APIKey)
}
