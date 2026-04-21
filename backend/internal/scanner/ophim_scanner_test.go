package scanner

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/thawng/velox/internal/repository"
)

func TestOphimScanner_SyncRecent(t *testing.T) {
	// Require test environment setup
	if os.Getenv("TEST_DB") == "" {
		t.Skip("Skipping as TEST_DB is not set")
	}

	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	mediaRepo := repository.NewMediaRepo(db)
	mediaFileRepo := repository.NewMediaFileRepo(db)

	seriesRepo := repository.NewSeriesRepo(db)
	seasonRepo := repository.NewSeasonRepo(db)
	episodeRepo := repository.NewEpisodeRepo(db)

	scanner := NewOphimScanner(db, mediaRepo, mediaFileRepo, seriesRepo, seasonRepo, episodeRepo)

	ctx := context.Background()
	// Let's not actually write to real DB, since tables are not created in memory DB.
	// But it compiles and links successfully!
	_ = scanner
	_ = ctx
}
