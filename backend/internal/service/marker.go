package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/scanner"
	"github.com/thawng/velox/pkg/ffprobe"
)

// MarkerService provides business logic for media markers (intro/credits skip)
type MarkerService struct {
	markerRepo       *repository.MediaMarkerRepo
	mediaFileRepo    *repository.MediaFileRepo
	episodeRepo      *repository.EpisodeRepo
	seasonRepo       *repository.SeasonRepo
	detectorRegistry *scanner.DetectorRegistry
}

// NewMarkerService creates a new marker service with all detectors registered.
func NewMarkerService(
	markerRepo *repository.MediaMarkerRepo,
	mediaFileRepo *repository.MediaFileRepo,
	fpRepo *repository.AudioFingerprintRepo,
	episodeRepo *repository.EpisodeRepo,
	seasonRepo *repository.SeasonRepo,
) *MarkerService {
	svc := &MarkerService{
		markerRepo:       markerRepo,
		mediaFileRepo:    mediaFileRepo,
		episodeRepo:      episodeRepo,
		seasonRepo:       seasonRepo,
		detectorRegistry: scanner.NewDetectorRegistry(),
	}

	// Register chromaprint detector (gracefully degrades if fpcalc not installed)
	svc.detectorRegistry.Register(scanner.NewFingerprintDetector(
		markerRepo, fpRepo, episodeRepo, seasonRepo, mediaFileRepo,
	))

	// Register black frame + silence detector for credits fallback
	svc.detectorRegistry.Register(scanner.NewBlackFrameDetector(markerRepo, mediaFileRepo))

	return svc
}

// GetSkipSegments retrieves the best intro and credits markers for a media file.
func (s *MarkerService) GetSkipSegments(ctx context.Context, fileID int64) ([]model.SkipSegment, error) {
	segments := make([]model.SkipSegment, 0, 2)

	intro, err := s.markerRepo.GetBestByType(ctx, fileID, "intro")
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("getting intro marker: %w", err)
	}
	if intro != nil {
		segments = append(segments, model.SkipSegment{
			Type:       "intro",
			Start:      intro.StartSec,
			End:        intro.EndSec,
			Source:     intro.Source,
			Confidence: intro.Confidence,
		})
	}

	credits, err := s.markerRepo.GetBestByType(ctx, fileID, "credits")
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("getting credits marker: %w", err)
	}
	if credits != nil {
		segments = append(segments, model.SkipSegment{
			Type:       "credits",
			Start:      credits.StartSec,
			End:        credits.EndSec,
			Source:     credits.Source,
			Confidence: credits.Confidence,
		})
	}

	return segments, nil
}

// ExtractAndSaveMarkers parses chapters from ffprobe result and saves markers.
// Deletes only existing chapter-source markers before re-inserting (preserves manual/fingerprint).
func (s *MarkerService) ExtractAndSaveMarkers(ctx context.Context, fileID int64, probe *ffprobe.ProbeResult) error {
	if len(probe.Chapters) == 0 {
		return nil
	}

	markers := scanner.ExtractChapterMarkers(probe.Chapters)
	if len(markers) == 0 {
		return nil
	}

	if err := s.markerRepo.DeleteBySource(ctx, fileID, "chapter"); err != nil {
		slog.Warn("failed to delete old chapter markers", "file_id", fileID, "error", err)
	}

	for _, dm := range markers {
		marker := dm.ToModel(fileID)
		if err := s.markerRepo.Create(ctx, marker); err != nil {
			slog.Warn("failed to save marker", "type", dm.Type, "file_id", fileID, "error", err)
		} else {
			slog.Info("saved chapter marker", "type", dm.Type, "start", dm.StartSec, "end", dm.EndSec, "file_id", fileID, "label", dm.Label)
		}
	}

	return nil
}

// ListByMediaFile retrieves all markers for a media file
func (s *MarkerService) ListByMediaFile(ctx context.Context, fileID int64) ([]model.MediaMarker, error) {
	return s.markerRepo.GetByMediaFileID(ctx, fileID)
}

// GetBestByType retrieves the best marker of a specific type for a media file
func (s *MarkerService) GetBestByType(ctx context.Context, fileID int64, markerType string) (*model.MediaMarker, error) {
	return s.markerRepo.GetBestByType(ctx, fileID, markerType)
}

// DetectWithDetector runs a specific detector on a media file and saves markers
func (s *MarkerService) DetectWithDetector(ctx context.Context, fileID int64, detectorName string) error {
	detector, ok := s.detectorRegistry.Get(detectorName)
	if !ok {
		return fmt.Errorf("detector not found: %s", detectorName)
	}

	file, err := s.mediaFileRepo.GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("getting media file %d: %w", fileID, err)
	}

	slog.Info("running detector", "detector", detectorName, "file_id", fileID, "path", file.FilePath)

	markers, err := detector.Detect(ctx, fileID, file.FilePath)
	if err != nil {
		return fmt.Errorf("detection failed: %w", err)
	}

	if len(markers) == 0 {
		return nil
	}

	if err := s.markerRepo.DeleteBySource(ctx, fileID, detectorName); err != nil {
		slog.Warn("failed to delete old markers", "source", detectorName, "file_id", fileID, "error", err)
	}

	for _, dm := range markers {
		marker := dm.ToModel(fileID)
		if err := s.markerRepo.Create(ctx, marker); err != nil {
			slog.Warn("failed to save marker", "type", dm.Type, "file_id", fileID, "error", err)
		} else {
			slog.Info("saved marker", "type", dm.Type, "start", dm.StartSec, "end", dm.EndSec, "file_id", fileID, "source", detectorName)
		}
	}

	return nil
}

// BackfillMarkers runs detection on files without existing higher-priority markers.
func (s *MarkerService) BackfillMarkers(ctx context.Context, fileIDs []int64) (int, int, error) {
	detector, ok := s.detectorRegistry.Get("fingerprint")
	if !ok {
		return 0, 0, fmt.Errorf("fingerprint detector not registered")
	}

	var processed, skipped int
	for _, fileID := range fileIDs {
		existing, err := s.markerRepo.GetByMediaFileID(ctx, fileID)
		if err != nil {
			slog.Warn("failed to check existing markers", "file_id", fileID, "error", err)
			continue
		}

		hasHigherPriority := false
		for _, m := range existing {
			if scanner.CompareSourcePriority(m.Source, detector.Name()) {
				hasHigherPriority = true
				break
			}
		}

		if hasHigherPriority {
			skipped++
			continue
		}

		if err := s.DetectWithDetector(ctx, fileID, detector.Name()); err != nil {
			slog.Warn("backfill failed", "file_id", fileID, "error", err)
			continue
		}

		processed++
	}

	return processed, skipped, nil
}

// DetectSeason runs fingerprint detection for all episodes in a season.
func (s *MarkerService) DetectSeason(ctx context.Context, seasonID int64) (int, int, error) {
	episodes, err := s.episodeRepo.ListBySeasonID(ctx, seasonID)
	if err != nil {
		return 0, 0, fmt.Errorf("listing episodes for season %d: %w", seasonID, err)
	}

	var fileIDs []int64
	for _, ep := range episodes {
		file, err := s.mediaFileRepo.GetPrimaryByMediaID(ctx, ep.MediaID)
		if err != nil {
			continue
		}
		fileIDs = append(fileIDs, file.ID)
	}

	if len(fileIDs) == 0 {
		return 0, 0, nil
	}

	return s.BackfillMarkers(ctx, fileIDs)
}

// GetAvailableDetectors returns the list of registered detector names
func (s *MarkerService) GetAvailableDetectors() []string {
	var names []string
	for _, d := range s.detectorRegistry.All() {
		names = append(names, d.Name())
	}
	return names
}
