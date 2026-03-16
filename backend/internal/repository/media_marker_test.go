package repository

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/thawng/velox/internal/model"
)

func TestMediaMarkerRepo(t *testing.T) {
	ctx := context.Background()

	// Use in-memory database
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("opening db: %v", err)
	}
	defer db.Close()

	// Create tables
	if _, err := db.Exec(`
		CREATE TABLE media_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT
		);
		CREATE TABLE media_markers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_file_id INTEGER NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
			marker_type TEXT NOT NULL CHECK (marker_type IN ('intro','credits')),
			start_sec REAL NOT NULL,
			end_sec REAL NOT NULL,
			source TEXT NOT NULL CHECK (source IN ('chapter','fingerprint','manual')),
			confidence REAL NOT NULL DEFAULT 1.0,
			label TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT valid_start CHECK (start_sec >= 0),
			CONSTRAINT valid_end CHECK (end_sec > start_sec)
		);
		CREATE UNIQUE INDEX idx_marker_unique ON media_markers(media_file_id, marker_type, source, start_sec, end_sec);
		CREATE INDEX idx_marker_lookup ON media_markers(media_file_id, marker_type);
		INSERT INTO media_files (id) VALUES (1), (2);
	`); err != nil {
		t.Fatalf("creating tables: %v", err)
	}

	repo := NewMediaMarkerRepo(db)

	t.Run("Create", func(t *testing.T) {
		marker := &model.MediaMarker{
			MediaFileID: 1,
			MarkerType:  "intro",
			StartSec:    10.5,
			EndSec:      85.0,
			Source:      "chapter",
			Confidence:  1.0,
			Label:       "Opening Credits",
		}

		if err := repo.Create(ctx, marker); err != nil {
			t.Fatalf("creating marker: %v", err)
		}

		if marker.ID == 0 {
			t.Error("expected ID to be set after create")
		}
		if marker.CreatedAt == "" {
			t.Error("expected CreatedAt to be set")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		// Create a marker first
		marker := &model.MediaMarker{
			MediaFileID: 1,
			MarkerType:  "credits",
			StartSec:    1200.0,
			EndSec:      1280.0,
			Source:      "chapter",
			Confidence:  1.0,
			Label:       "End Credits",
		}
		if err := repo.Create(ctx, marker); err != nil {
			t.Fatalf("creating marker: %v", err)
		}

		// Get it back
		got, err := repo.GetByID(ctx, marker.ID)
		if err != nil {
			t.Fatalf("getting marker: %v", err)
		}

		if got.MarkerType != "credits" {
			t.Errorf("expected type=credits, got %s", got.MarkerType)
		}
		if got.StartSec != 1200.0 {
			t.Errorf("expected start=1200, got %f", got.StartSec)
		}
	})

	t.Run("GetByMediaFileID", func(t *testing.T) {
		markers, err := repo.GetByMediaFileID(ctx, 1)
		if err != nil {
			t.Fatalf("getting markers: %v", err)
		}

		// Should have 2 markers from previous tests
		if len(markers) != 2 {
			t.Errorf("expected 2 markers, got %d", len(markers))
		}

		// Should be sorted by start_sec
		if len(markers) >= 2 && markers[0].StartSec > markers[1].StartSec {
			t.Error("expected markers sorted by start_sec ASC")
		}
	})

	t.Run("GetByType", func(t *testing.T) {
		markers, err := repo.GetByType(ctx, 1, "intro")
		if err != nil {
			t.Fatalf("getting markers by type: %v", err)
		}

		if len(markers) != 1 {
			t.Errorf("expected 1 intro marker, got %d", len(markers))
		}
	})

	t.Run("GetBestByType", func(t *testing.T) {
		// Create manual marker (higher priority)
		manualMarker := &model.MediaMarker{
			MediaFileID: 1,
			MarkerType:  "intro",
			StartSec:    11.0,
			EndSec:      84.0,
			Source:      "manual",
			Confidence:  1.0,
			Label:       "User defined intro",
		}
		if err := repo.Create(ctx, manualMarker); err != nil {
			t.Fatalf("creating manual marker: %v", err)
		}

		// Get best should return manual (higher priority)
		best, err := repo.GetBestByType(ctx, 1, "intro")
		if err != nil {
			t.Fatalf("getting best marker: %v", err)
		}

		if best.Source != "manual" {
			t.Errorf("expected best source=manual, got %s", best.Source)
		}
	})

	t.Run("DeleteBySource", func(t *testing.T) {
		// Create marker for file 2
		marker := &model.MediaMarker{
			MediaFileID: 2,
			MarkerType:  "intro",
			StartSec:    5.0,
			EndSec:      30.0,
			Source:      "chapter",
			Confidence:  1.0,
			Label:       "Chapter intro",
		}
		if err := repo.Create(ctx, marker); err != nil {
			t.Fatalf("creating marker: %v", err)
		}

		// Delete all chapter markers for file 2
		if err := repo.DeleteBySource(ctx, 2, "chapter"); err != nil {
			t.Fatalf("deleting by source: %v", err)
		}

		// Verify deleted
		markers, err := repo.GetByMediaFileID(ctx, 2)
		if err != nil {
			t.Fatalf("getting markers: %v", err)
		}
		if len(markers) != 0 {
			t.Errorf("expected 0 markers after delete, got %d", len(markers))
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 99999)
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
	})
}
