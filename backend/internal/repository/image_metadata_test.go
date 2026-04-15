package repository

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/thawng/velox/internal/model"
)

func setupImageMetadataDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
		CREATE TABLE image_metadata (
			path        TEXT PRIMARY KEY,
			blurhash    TEXT NOT NULL,
			width       INTEGER NOT NULL,
			height      INTEGER NOT NULL,
			computed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestImageMetadataRepo_UpsertAndGet(t *testing.T) {
	db := setupImageMetadataDB(t)
	defer db.Close()
	repo := NewImageMetadataRepo(db)
	ctx := context.Background()

	// Upsert a new record
	err := repo.Upsert(ctx, &model.ImageMetadata{
		Path:     "/test/path.jpg",
		Blurhash: "LEHV6nWB2yk8pyo0adR*.7kCMdnj",
		Width:    800,
		Height:   600,
	})
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	// Get it back
	meta, err := repo.Get(ctx, "/test/path.jpg")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if meta == nil {
		t.Fatal("expected meta to be returned")
	}
	if meta.Blurhash != "LEHV6nWB2yk8pyo0adR*.7kCMdnj" {
		t.Errorf("expected blurhash LEHV6nWB2yk8pyo0adR*.7kCMdnj, got %s", meta.Blurhash)
	}

	// Upsert to update
	err = repo.Upsert(ctx, &model.ImageMetadata{
		Path:     "/test/path.jpg",
		Blurhash: "NEW_BLURHASH",
		Width:    1024,
		Height:   768,
	})
	if err != nil {
		t.Fatalf("Upsert update failed: %v", err)
	}

	meta, err = repo.Get(ctx, "/test/path.jpg")
	if err != nil {
		t.Fatalf("Get updated failed: %v", err)
	}
	if meta.Blurhash != "NEW_BLURHASH" {
		t.Errorf("expected updated blurhash NEW_BLURHASH, got %s", meta.Blurhash)
	}
}

func TestImageMetadataRepo_GetBatch(t *testing.T) {
	db := setupImageMetadataDB(t)
	defer db.Close()
	repo := NewImageMetadataRepo(db)
	ctx := context.Background()

	// Seed data
	repo.Upsert(ctx, &model.ImageMetadata{Path: "/p1.jpg", Blurhash: "B1", Width: 100, Height: 100})
	repo.Upsert(ctx, &model.ImageMetadata{Path: "/p2.jpg", Blurhash: "B2", Width: 200, Height: 200})

	// GetBatch
	res, err := repo.GetBatch(ctx, []string{"/p1.jpg", "/p2.jpg", "/missing.jpg", "/p2.jpg", ""})
	if err != nil {
		t.Fatalf("GetBatch failed: %v", err)
	}

	if len(res) != 2 {
		t.Errorf("expected 2 results, got %d", len(res))
	}
	if res["/p1.jpg"].Blurhash != "B1" {
		t.Errorf("expected B1 for /p1.jpg, got %v", res["/p1.jpg"])
	}
	if res["/p2.jpg"].Blurhash != "B2" {
		t.Errorf("expected B2 for /p2.jpg, got %v", res["/p2.jpg"])
	}
	if res["/missing.jpg"] != nil {
		t.Errorf("expected nil for missing, got %v", res["/missing.jpg"])
	}
}
