package service

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

func setupMarkerTestDB(t *testing.T) (*sql.DB, *repository.MediaMarkerRepo, *repository.MediaFileRepo, *repository.AudioFingerprintRepo, *repository.EpisodeRepo, *repository.SeasonRepo) {
	t.Helper()

	db, err := sql.Open("sqlite3", "file::memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("opening db: %v", err)
	}

	schema := `
	CREATE TABLE libraries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		media_type TEXT NOT NULL,
		paths TEXT NOT NULL DEFAULT '[]',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE media (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		library_id INTEGER NOT NULL REFERENCES libraries(id),
		media_type TEXT NOT NULL,
		title TEXT NOT NULL,
		sort_title TEXT NOT NULL,
		release_date TEXT,
		tmdb_id INTEGER,
		imdb_id TEXT,
		overview TEXT,
		runtime INTEGER DEFAULT 0,
		is_hidden BOOLEAN DEFAULT 0
	);
	CREATE TABLE media_files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
		file_path TEXT NOT NULL UNIQUE,
		file_size INTEGER DEFAULT 0,
		duration REAL DEFAULT 0,
		width INTEGER DEFAULT 0,
		height INTEGER DEFAULT 0,
		video_codec TEXT DEFAULT '',
		video_profile TEXT DEFAULT '',
		video_level INTEGER DEFAULT 0,
		video_fps REAL DEFAULT 0,
		audio_codec TEXT DEFAULT '',
		container TEXT DEFAULT '',
		bitrate INTEGER DEFAULT 0,
		fingerprint TEXT DEFAULT '',
		is_primary INTEGER DEFAULT 1,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_verified_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("creating schema: %v", err)
	}

	if _, err := db.Exec(`
		INSERT INTO libraries (id, name, media_type, paths) VALUES (1, 'Test', 'movie', '["/test"]');
		INSERT INTO media (id, library_id, media_type, title, sort_title) VALUES (1, 1, 'movie', 'Test Movie', 'Test Movie');
		INSERT INTO media_files (id, media_id, file_path, file_size) VALUES (1, 1, '/test/movie1.mkv', 1000);
		INSERT INTO media_files (id, media_id, file_path, file_size) VALUES (2, 1, '/test/movie2.mkv', 2000);
		INSERT INTO media_files (id, media_id, file_path, file_size) VALUES (3, 1, '/test/movie3.mkv', 3000);
	`); err != nil {
		t.Fatalf("inserting test data: %v", err)
	}

	return db, repository.NewMediaMarkerRepo(db), repository.NewMediaFileRepo(db),
		repository.NewAudioFingerprintRepo(db), repository.NewEpisodeRepo(db), repository.NewSeasonRepo(db)
}

func TestGetSkipSegments(t *testing.T) {
	ctx := context.Background()
	db, markerRepo, mediaFileRepo, fpRepo, episodeRepo, seasonRepo := setupMarkerTestDB(t)
	defer db.Close()

	svc := NewMarkerService(markerRepo, mediaFileRepo, fpRepo, episodeRepo, seasonRepo)

	// No markers → empty segments
	segments, err := svc.GetSkipSegments(ctx, 1)
	if err != nil {
		t.Fatalf("GetSkipSegments: %v", err)
	}
	if len(segments) != 0 {
		t.Errorf("expected 0 segments, got %d", len(segments))
	}

	// Add intro + credits markers
	markerRepo.Create(ctx, &model.MediaMarker{
		MediaFileID: 1, MarkerType: "intro", StartSec: 10, EndSec: 85,
		Source: "chapter", Confidence: 1.0, Label: "Opening",
	})
	markerRepo.Create(ctx, &model.MediaMarker{
		MediaFileID: 1, MarkerType: "credits", StartSec: 2500, EndSec: 2580,
		Source: "chapter", Confidence: 1.0, Label: "End Credits",
	})

	segments, err = svc.GetSkipSegments(ctx, 1)
	if err != nil {
		t.Fatalf("GetSkipSegments: %v", err)
	}
	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}
	if segments[0].Type != "intro" || segments[0].Start != 10 || segments[0].End != 85 {
		t.Errorf("intro segment mismatch: %+v", segments[0])
	}
	if segments[1].Type != "credits" || segments[1].Start != 2500 || segments[1].End != 2580 {
		t.Errorf("credits segment mismatch: %+v", segments[1])
	}
}

func TestGetSkipSegments_SourcePriority(t *testing.T) {
	ctx := context.Background()
	db, markerRepo, mediaFileRepo, fpRepo, episodeRepo, seasonRepo := setupMarkerTestDB(t)
	defer db.Close()

	svc := NewMarkerService(markerRepo, mediaFileRepo, fpRepo, episodeRepo, seasonRepo)

	// Add fingerprint intro (low priority)
	markerRepo.Create(ctx, &model.MediaMarker{
		MediaFileID: 1, MarkerType: "intro", StartSec: 10, EndSec: 80,
		Source: "fingerprint", Confidence: 0.75, Label: "Detected",
	})
	// Add chapter intro (higher priority)
	markerRepo.Create(ctx, &model.MediaMarker{
		MediaFileID: 1, MarkerType: "intro", StartSec: 12, EndSec: 86,
		Source: "chapter", Confidence: 1.0, Label: "Opening",
	})

	segments, err := svc.GetSkipSegments(ctx, 1)
	if err != nil {
		t.Fatalf("GetSkipSegments: %v", err)
	}
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}
	// Should return chapter (higher priority), not fingerprint
	if segments[0].Source != "chapter" || segments[0].Start != 12 {
		t.Errorf("expected chapter source with start=12, got %+v", segments[0])
	}
}

func TestBackfillMarkers_SkipsWhenHigherPriorityExists(t *testing.T) {
	ctx := context.Background()
	db, markerRepo, mediaFileRepo, fpRepo, episodeRepo, seasonRepo := setupMarkerTestDB(t)
	defer db.Close()

	svc := NewMarkerService(markerRepo, mediaFileRepo, fpRepo, episodeRepo, seasonRepo)

	// File 1: manual marker (highest priority)
	markerRepo.Create(ctx, &model.MediaMarker{
		MediaFileID: 1, MarkerType: "intro", StartSec: 10, EndSec: 85,
		Source: "manual", Confidence: 1.0, Label: "User defined",
	})
	// File 2: chapter marker (higher than fingerprint)
	markerRepo.Create(ctx, &model.MediaMarker{
		MediaFileID: 2, MarkerType: "intro", StartSec: 12, EndSec: 86,
		Source: "chapter", Confidence: 1.0, Label: "Opening",
	})
	// File 3: no markers

	processed, skipped, err := svc.BackfillMarkers(ctx, []int64{1, 2, 3})
	if err != nil {
		t.Fatalf("backfill failed: %v", err)
	}
	if skipped != 2 {
		t.Errorf("expected skipped=2, got %d", skipped)
	}
	if processed != 1 {
		t.Errorf("expected processed=1, got %d", processed)
	}
}

func TestBackfillMarkers_AllowsFingerprintWhenNoMarkers(t *testing.T) {
	ctx := context.Background()
	db, markerRepo, mediaFileRepo, fpRepo, episodeRepo, seasonRepo := setupMarkerTestDB(t)
	defer db.Close()

	svc := NewMarkerService(markerRepo, mediaFileRepo, fpRepo, episodeRepo, seasonRepo)

	// File 3 has no markers — fingerprint detector is a stub so processed=1 but no markers saved
	processed, skipped, err := svc.BackfillMarkers(ctx, []int64{3})
	if err != nil {
		t.Fatalf("backfill failed: %v", err)
	}
	if skipped != 0 {
		t.Errorf("expected skipped=0, got %d", skipped)
	}
	if processed != 1 {
		t.Errorf("expected processed=1, got %d", processed)
	}
}

func TestDetectWithDetector_NotFound(t *testing.T) {
	ctx := context.Background()
	db, markerRepo, mediaFileRepo, fpRepo, episodeRepo, seasonRepo := setupMarkerTestDB(t)
	defer db.Close()

	svc := NewMarkerService(markerRepo, mediaFileRepo, fpRepo, episodeRepo, seasonRepo)

	err := svc.DetectWithDetector(ctx, 1, "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent detector")
	}
}

func TestGetAvailableDetectors(t *testing.T) {
	_, markerRepo, mediaFileRepo, fpRepo, episodeRepo, seasonRepo := setupMarkerTestDB(t)

	svc := NewMarkerService(markerRepo, mediaFileRepo, fpRepo, episodeRepo, seasonRepo)
	names := svc.GetAvailableDetectors()

	foundFingerprint := false
	for _, name := range names {
		if name == "fingerprint" {
			foundFingerprint = true
			break
		}
	}
	if !foundFingerprint {
		t.Errorf("expected 'fingerprint' detector, got: %v", names)
	}
}
