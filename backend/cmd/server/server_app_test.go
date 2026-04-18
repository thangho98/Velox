package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/thawng/velox/internal/config"
	"github.com/thawng/velox/internal/database"
	"github.com/thawng/velox/internal/model"
)

func openServerTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "server-app-test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	return db
}

func TestConfiguredAPIKeyPrefersPersistedAniListSettingAfterRestart(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := openServerTestDB(t)
	repos := newServerRepos(db)

	if err := repos.appSettings.Set(ctx, model.SettingAniListToken, "stored-anilist-token"); err != nil {
		t.Fatalf("persist AniList token: %v", err)
	}

	app := &serverApp{
		cfg: &config.Config{
			AniListToken: "builtin-anilist-token",
		},
		repos: repos,
	}

	got, builtin := app.configuredAPIKey(ctx, model.SettingAniListToken, app.cfg.AniListToken)
	if got != "stored-anilist-token" {
		t.Fatalf("expected persisted AniList token after restart, got %q", got)
	}
	if builtin {
		t.Fatalf("expected persisted AniList token to override builtin source")
	}
}
