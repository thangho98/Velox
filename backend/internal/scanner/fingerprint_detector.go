package scanner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// FingerprintDetectorConfig holds configuration for the fingerprint detector.
type FingerprintDetectorConfig struct {
	MaxBitDiff       int     // Max Hamming distance for matching (default 10)
	MinEpisodesMatch int     // Min episodes that must match to accept (default 2)
	MinIntroDuration float64 // Min intro duration in seconds (default 15)
	MaxIntroDuration float64 // Max intro duration in seconds (default 120)
}

// DefaultFingerprintConfig returns sensible defaults.
func DefaultFingerprintConfig() FingerprintDetectorConfig {
	return FingerprintDetectorConfig{
		MaxBitDiff:       10,
		MinEpisodesMatch: 2,
		MinIntroDuration: 15,
		MaxIntroDuration: 120,
	}
}

// FingerprintDetector uses chromaprint audio fingerprinting to detect intro/credits.
// It compares audio fingerprints across episodes in the same season.
type FingerprintDetector struct {
	markerRepo    *repository.MediaMarkerRepo
	fpRepo        *repository.AudioFingerprintRepo
	episodeRepo   *repository.EpisodeRepo
	seasonRepo    *repository.SeasonRepo
	mediaFileRepo *repository.MediaFileRepo
	config        FingerprintDetectorConfig
	fpcalcAvail   bool
}

// NewFingerprintDetector creates a new chromaprint-based detector.
func NewFingerprintDetector(
	markerRepo *repository.MediaMarkerRepo,
	fpRepo *repository.AudioFingerprintRepo,
	episodeRepo *repository.EpisodeRepo,
	seasonRepo *repository.SeasonRepo,
	mediaFileRepo *repository.MediaFileRepo,
) *FingerprintDetector {
	avail := CheckFpcalc()
	if !avail {
		slog.Warn("fpcalc not found — chromaprint detection disabled. Install: brew install chromaprint")
	}
	return &FingerprintDetector{
		markerRepo:    markerRepo,
		fpRepo:        fpRepo,
		episodeRepo:   episodeRepo,
		seasonRepo:    seasonRepo,
		mediaFileRepo: mediaFileRepo,
		config:        DefaultFingerprintConfig(),
		fpcalcAvail:   avail,
	}
}

func (d *FingerprintDetector) Name() string        { return "fingerprint" }
func (d *FingerprintDetector) Confidence() float64 { return 0.75 }

// Detect finds intro markers by comparing this file's audio fingerprint with
// other episodes in the same season.
func (d *FingerprintDetector) Detect(ctx context.Context, fileID int64, filePath string) ([]DetectedMarker, error) {
	if !d.fpcalcAvail {
		return nil, nil
	}

	// Skip if higher-priority markers exist
	existing, err := d.markerRepo.GetByMediaFileID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("checking existing markers: %w", err)
	}
	for _, m := range existing {
		if m.Source == "manual" || m.Source == "chapter" {
			slog.Debug("fingerprint: skipping, higher priority markers exist", "file_id", fileID)
			return nil, nil
		}
	}

	// Find season for this episode
	file, err := d.mediaFileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("getting media file: %w", err)
	}

	episode, err := d.episodeRepo.GetByMediaID(ctx, file.MediaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Debug("fingerprint: not an episode, skipping", "file_id", fileID)
			return nil, nil // Not an episode (movie) — can't cross-compare
		}
		return nil, fmt.Errorf("getting episode: %w", err)
	}

	if episode.SeasonID == 0 {
		return nil, nil
	}

	// Get all episodes in the same season
	episodes, err := d.episodeRepo.ListBySeasonID(ctx, episode.SeasonID)
	if err != nil {
		return nil, fmt.Errorf("listing season episodes: %w", err)
	}

	if len(episodes) < d.config.MinEpisodesMatch {
		slog.Debug("fingerprint: not enough episodes in season", "count", len(episodes), "min", d.config.MinEpisodesMatch)
		return nil, nil
	}

	// Get primary files for each episode and extract/cache fingerprints
	fingerprints := make(map[int64][]uint32)
	for _, ep := range episodes {
		epFile, err := d.mediaFileRepo.GetPrimaryByMediaID(ctx, ep.MediaID)
		if err != nil {
			slog.Debug("fingerprint: no primary file for episode", "media_id", ep.MediaID, "error", err)
			continue
		}

		fp, err := d.getOrExtractFingerprint(ctx, epFile.ID, epFile.FilePath, epFile.Duration)
		if err != nil {
			slog.Warn("fingerprint: extraction failed", "file_id", epFile.ID, "error", err)
			continue
		}
		if fp != nil {
			fingerprints[epFile.ID] = fp
		}
	}

	if len(fingerprints) < d.config.MinEpisodesMatch {
		return nil, nil
	}

	// Find season intro
	result := FindSeasonIntro(fingerprints, d.config.MinEpisodesMatch, d.config.MaxBitDiff)
	if result == nil {
		slog.Info("fingerprint: no intro found for season", "season_id", episode.SeasonID)
		return nil, nil
	}

	// Get this file's specific timestamps
	ts, ok := result.PerFile[fileID]
	if !ok {
		// Use consensus timestamps
		ts = IntroTimestamps{Start: result.Start, End: result.End}
	}

	// Validate duration
	dur := ts.End - ts.Start
	if dur < d.config.MinIntroDuration || dur > d.config.MaxIntroDuration {
		return nil, nil
	}

	slog.Info("fingerprint: detected intro",
		"file_id", fileID,
		"start", ts.Start,
		"end", ts.End,
		"confidence", result.Confidence,
		"match_count", result.MatchCount,
	)

	return []DetectedMarker{
		{
			Type:       "intro",
			StartSec:   ts.Start,
			EndSec:     ts.End,
			Source:     "fingerprint",
			Confidence: result.Confidence,
			Label:      fmt.Sprintf("chromaprint (%d/%d episodes)", result.MatchCount, len(fingerprints)),
		},
	}, nil
}

// getOrExtractFingerprint retrieves cached fingerprint or extracts a new one.
func (d *FingerprintDetector) getOrExtractFingerprint(ctx context.Context, fileID int64, filePath string, duration float64) ([]uint32, error) {
	// Check cache
	cached, err := d.fpRepo.GetByMediaFileID(ctx, fileID, "intro_region")
	if err == nil && cached != nil && len(cached.Fingerprint) > 0 {
		return BytesToFingerprint(cached.Fingerprint), nil
	}

	// Extract new fingerprint
	slog.Debug("fingerprint: extracting", "file_id", fileID)
	fp, regionDur, err := ExtractIntroRegion(ctx, filePath, duration)
	if err != nil {
		return nil, err
	}
	if len(fp) == 0 {
		return nil, nil
	}

	// Cache in DB
	afp := &model.AudioFingerprint{
		MediaFileID: fileID,
		Region:      "intro_region",
		Fingerprint: FingerprintToBytes(fp),
		DurationSec: regionDur,
		SampleCount: len(fp),
	}
	if err := d.fpRepo.Upsert(ctx, afp); err != nil {
		slog.Warn("fingerprint: failed to cache", "file_id", fileID, "error", err)
	}

	return fp, nil
}
