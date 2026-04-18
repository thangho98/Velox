package scanner

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/thawng/velox/internal/database"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/pkg/ffprobe"
	"github.com/thawng/velox/pkg/nameparser"
)

type metadataMatcherSpy struct {
	animeCalls   int
	movieCalls   int
	episodeCalls int
}

func (m *metadataMatcherSpy) MatchAndPersistMovie(ctx context.Context, media *model.Media, parsed nameparser.ParsedMedia, filePath string, force bool) error {
	m.movieCalls++
	return nil
}

func (m *metadataMatcherSpy) MatchAndPersistEpisode(ctx context.Context, media *model.Media, parsed nameparser.ParsedMedia, filePath string, libraryID int64, force bool) error {
	m.episodeCalls++
	return nil
}

func (m *metadataMatcherSpy) MatchAndPersistAnime(ctx context.Context, media *model.Media, parsed nameparser.ParsedMedia, filePath string, libraryID int64, force bool) error {
	m.animeCalls++
	return nil
}

func openPipelineTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "pipeline-test.db")
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

func TestPersistRoutesAnimeLibrariesToAnimeMatcher(t *testing.T) {
	t.Parallel()

	db := openPipelineTestDB(t)
	ctx := context.Background()

	libraryRepo := repository.NewLibraryRepo(db)
	mediaRepo := repository.NewMediaRepo(db)
	mediaFileRepo := repository.NewMediaFileRepo(db)
	seriesRepo := repository.NewSeriesRepo(db)
	seasonRepo := repository.NewSeasonRepo(db)
	episodeRepo := repository.NewEpisodeRepo(db)
	scanJobRepo := repository.NewScanJobRepo(db)
	subtitleRepo := repository.NewSubtitleRepo(db)
	audioTrackRepo := repository.NewAudioTrackRepo(db)

	pipeline := NewPipeline(
		db,
		libraryRepo,
		mediaRepo,
		mediaFileRepo,
		seriesRepo,
		seasonRepo,
		episodeRepo,
		scanJobRepo,
		subtitleRepo,
		audioTrackRepo,
		nil,
		nil,
	)

	spy := &metadataMatcherSpy{}
	pipeline.SetMetadataMatcher(spy)

	root := t.TempDir()
	lib, err := libraryRepo.Create(ctx, "Anime", model.LibraryTypeAnime, []string{root})
	if err != nil {
		t.Fatalf("create library: %v", err)
	}

	videoPath := filepath.Join(root, "Gintama - 01.mkv")
	if err := os.WriteFile(videoPath, []byte("anime"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	scanCtx := &ScanContext{
		LibraryID:   lib.ID,
		LibraryType: model.LibraryTypeAnime,
		ctx:         ctx,
	}

	probe := &ffprobe.ProbeResult{
		Duration:   1440,
		Width:      1920,
		Height:     1080,
		VideoCodec: "h264",
		AudioCodec: "aac",
		Container:  "matroska",
		Bitrate:    2_000_000,
	}
	parsed := nameparser.ParsedMedia{
		Title:     "Gintama",
		MediaType: "episode",
		Season:    1,
		Episode:   1,
	}

	if err := pipeline.persist(scanCtx, videoPath, "5:fingerprint", probe, parsed, 0, 0, true); err != nil {
		t.Fatalf("persist media: %v", err)
	}

	if spy.animeCalls != 1 {
		t.Fatalf("expected anime matcher to be called once, got %d", spy.animeCalls)
	}
	if spy.movieCalls != 0 || spy.episodeCalls != 0 {
		t.Fatalf("expected only anime matcher route, got movie=%d episode=%d", spy.movieCalls, spy.episodeCalls)
	}
}
