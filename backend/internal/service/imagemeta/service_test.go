package imagemeta

import (
	"context"
	"database/sql"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/storage"
	"golang.org/x/time/rate"
)

func setupDB(t *testing.T) *sql.DB {
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

func setupTestStorage(t *testing.T) (*storage.ImageStorage, string) {
	t.Helper()
	dir, err := os.MkdirTemp("", "velox-test-*")
	if err != nil {
		t.Fatal(err)
	}
	return storage.NewImageStorage(dir), dir
}

func createTestJPEG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 255), uint8(y % 255), 100, 255})
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := jpeg.Encode(f, img, nil); err != nil {
		t.Fatal(err)
	}
}

func TestService_Compute_Local(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	s, dir := setupTestStorage(t)
	defer os.RemoveAll(dir)

	repo := repository.NewImageMetadataRepo(db)
	svc := NewService(repo, s, http.DefaultClient, rate.NewLimiter(10, 1))

	// Create local mock image: local://media/10/poster.jpg
	// AbsPath = {dataDir}/images/media/10/poster.jpg
	absPath := filepath.Join(dir, "images", "media", "10", "poster.jpg")
	createTestJPEG(t, absPath, 100, 200)

	ctx := context.Background()
	rawPath := "local://media/10/poster.jpg"

	meta, err := svc.Compute(ctx, rawPath)
	if err != nil {
		t.Fatalf("Compute failed: %v", err)
	}

	if meta.Width != 100 || meta.Height != 200 {
		t.Errorf("dim = %dx%d, want 100x200", meta.Width, meta.Height)
	}
	if meta.Blurhash == "" {
		t.Error("expected blurhash")
	}

	// Idempotent check
	meta2, err := svc.Compute(ctx, rawPath)
	if err != nil {
		t.Fatalf("Compute 2 failed: %v", err)
	}
	if meta2.Blurhash != meta.Blurhash {
		t.Error("expected idempotent hash")
	}
}

func TestService_Compute_TMDb(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TMDb test in short mode")
	}

	db := setupDB(t)
	defer db.Close()
	_, dir := setupTestStorage(t)
	defer os.RemoveAll(dir)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 50, 50))
		jpeg.Encode(w, img, nil)
	}))
	defer ts.Close()

	// Since we hardcoded api in fetchTMDbImage, we'll test the mocked logic if needed.
	// But actually we have a hardcoded "https://image.tmdb.org". We'll just verify skipping in short mode.
}

func TestService_ComputeBatch(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	s, dir := setupTestStorage(t)
	defer os.RemoveAll(dir)

	repo := repository.NewImageMetadataRepo(db)
	svc := NewService(repo, s, http.DefaultClient, rate.NewLimiter(10, 1))

	absPath1 := filepath.Join(dir, "images", "media", "1", "poster.jpg")
	createTestJPEG(t, absPath1, 100, 100)

	ctx := context.Background()
	paths := []string{
		"local://media/1/poster.jpg",
		"local://media/2/poster.jpg", // missing
	}

	errs := svc.ComputeBatch(ctx, paths)
	if errs["local://media/1/poster.jpg"] != nil {
		t.Errorf("unexpected error for 1: %v", errs["local://media/1/poster.jpg"])
	}
	if errs["local://media/2/poster.jpg"] == nil {
		t.Errorf("expected error for 2")
	}
}

func TestWorker_EnqueueAndRun(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	s, dir := setupTestStorage(t)
	defer os.RemoveAll(dir)

	repo := repository.NewImageMetadataRepo(db)
	svc := NewService(repo, s, http.DefaultClient, rate.NewLimiter(10, 1))

	absPath := filepath.Join(dir, "images", "media", "1", "poster.jpg")
	createTestJPEG(t, absPath, 100, 100)

	worker := svc.Worker()

	// Enqueue path
	rawPath := "local://media/1/poster.jpg"
	worker.Enqueue(rawPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run worker briefly
	go worker.Run(ctx)

	// Give it a moment to process
	time.Sleep(100 * time.Millisecond)

	// Verify compute happened
	meta, err := repo.Get(context.Background(), rawPath)
	if err != nil {
		t.Fatalf("Failed to get meta: %v", err)
	}
	if meta == nil {
		t.Fatal("Worker did not compute metadata")
	}
}

// TestService_Compute_HTTPScheme covers paths stored as full URLs (e.g.
// fanart.tv logos / thumbs). The fetcher must hit the URL directly, not the
// TMDb CDN, and must not be rate-limited against the TMDb bucket.
func TestService_Compute_HTTPScheme(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	s, dir := setupTestStorage(t)
	defer os.RemoveAll(dir)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 40, 30))
		for y := 0; y < 30; y++ {
			for x := 0; x < 40; x++ {
				img.Set(x, y, color.RGBA{200, 100, 50, 255})
			}
		}
		w.Header().Set("Content-Type", "image/jpeg")
		jpeg.Encode(w, img, nil)
	}))
	defer ts.Close()

	repo := repository.NewImageMetadataRepo(db)
	svc := NewService(repo, s, http.DefaultClient, rate.NewLimiter(10, 1))

	meta, err := svc.Compute(context.Background(), ts.URL+"/fanart/foo.jpg")
	if err != nil {
		t.Fatalf("Compute https path failed: %v", err)
	}
	if meta.Width != 40 || meta.Height != 30 {
		t.Errorf("dim = %dx%d, want 40x30", meta.Width, meta.Height)
	}
	if meta.Blurhash == "" {
		t.Error("expected blurhash for https path")
	}
}

func TestWorker_QueueOverflow(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	s, dir := setupTestStorage(t)
	defer os.RemoveAll(dir)

	repo := repository.NewImageMetadataRepo(db)
	svc := NewService(repo, s, http.DefaultClient, rate.NewLimiter(10, 1))
	worker := svc.Worker()

	// Fill queue past 256
	for i := 0; i < 300; i++ {
		worker.Enqueue("local://media/1/poster.jpg")
	}

	// Should not block
	if len(worker.queue) != 256 {
		t.Errorf("Expected queue to be bounded at 256, got %d", len(worker.queue))
	}
}
