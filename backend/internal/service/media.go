package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

type MediaService struct {
	repo          *repository.MediaRepo
	mediaFileRepo *repository.MediaFileRepo
	episodeRepo   *repository.EpisodeRepo
	seasonRepo    *repository.SeasonRepo
}

func NewMediaService(repo *repository.MediaRepo, mediaFileRepo *repository.MediaFileRepo) *MediaService {
	return &MediaService{
		repo:          repo,
		mediaFileRepo: mediaFileRepo,
	}
}

// SetEpisodeRepo sets the episode repo for episode info enrichment.
func (s *MediaService) SetEpisodeRepo(r *repository.EpisodeRepo) { s.episodeRepo = r }

// SetSeasonRepo sets the season repo for episode info enrichment.
func (s *MediaService) SetSeasonRepo(r *repository.SeasonRepo) { s.seasonRepo = r }

func (s *MediaService) List(ctx context.Context, libraryID int64, mediaType string, limit, offset int) ([]model.Media, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.List(ctx, libraryID, mediaType, limit, offset)
}

func (s *MediaService) Get(ctx context.Context, id int64) (*model.Media, error) {
	media, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return media, err
}

func (s *MediaService) GetWithFiles(ctx context.Context, id int64) (*model.MediaWithFiles, error) {
	media, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	files, err := s.mediaFileRepo.ListByMediaID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &model.MediaWithFiles{
		Media: *media,
		Files: files,
	}

	// Enrich with episode/season info when applicable
	if media.MediaType == "episode" && s.episodeRepo != nil && s.seasonRepo != nil {
		if ep, err := s.episodeRepo.GetByMediaID(ctx, id); err == nil {
			result.SeriesID = ep.SeriesID
			result.SeasonID = ep.SeasonID
			result.EpisodeNumber = ep.EpisodeNumber
			if season, err := s.seasonRepo.GetByID(ctx, ep.SeasonID); err == nil {
				result.SeasonNumber = season.SeasonNumber
			}
		}
	}

	return result, nil
}

func (s *MediaService) Search(ctx context.Context, query string, limit int) ([]model.Media, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.Search(ctx, query, limit)
}

// ListVersions returns all physical files for the given media, ordered by is_primary DESC.
// Returns ErrNotFound if the media ID does not exist.
func (s *MediaService) ListVersions(ctx context.Context, mediaID int64) ([]model.MediaFile, error) {
	if _, err := s.repo.GetByID(ctx, mediaID); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return s.mediaFileRepo.ListByMediaID(ctx, mediaID)
}
