package service

import (
	"context"
	"errors"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// SeriesService orchestrates series, season, and episode read operations.
type SeriesService struct {
	seriesRepo    *repository.SeriesRepo
	seasonRepo    *repository.SeasonRepo
	episodeRepo   *repository.EpisodeRepo
	mediaFileRepo *repository.MediaFileRepo
}

// NewSeriesService creates a new series service.
func NewSeriesService(
	seriesRepo *repository.SeriesRepo,
	seasonRepo *repository.SeasonRepo,
	episodeRepo *repository.EpisodeRepo,
	mediaFileRepo *repository.MediaFileRepo,
) *SeriesService {
	return &SeriesService{
		seriesRepo:    seriesRepo,
		seasonRepo:    seasonRepo,
		episodeRepo:   episodeRepo,
		mediaFileRepo: mediaFileRepo,
	}
}

func (s *SeriesService) ListFiltered(ctx context.Context, filter model.SeriesListFilter) ([]model.SeriesListItem, error) {
	return s.seriesRepo.ListFiltered(ctx, filter)
}

func (s *SeriesService) GetAlphabet(ctx context.Context, filter model.SeriesListFilter) ([]model.AlphabetCount, error) {
	return s.seriesRepo.GetAlphabet(ctx, filter)
}

func (s *SeriesService) Get(ctx context.Context, id int64) (*model.Series, error) {
	series, err := s.seriesRepo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return series, err
}

func (s *SeriesService) Search(ctx context.Context, query string, limit int) ([]model.Series, error) {
	return s.seriesRepo.Search(ctx, query, limit)
}

func (s *SeriesService) ListSeasons(ctx context.Context, seriesID int64) ([]model.Season, error) {
	return s.seasonRepo.ListBySeriesID(ctx, seriesID)
}

func (s *SeriesService) ListEpisodes(ctx context.Context, seasonID int64) ([]model.Episode, error) {
	episodes, err := s.episodeRepo.ListBySeasonID(ctx, seasonID)
	if err != nil {
		return nil, err
	}
	for i := range episodes {
		if files, err := s.mediaFileRepo.ListByMediaID(ctx, episodes[i].MediaID); err == nil {
			episodes[i].MediaFiles = files
		}
	}
	return episodes, nil
}
