package service

import (
	"context"
	"errors"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// SeriesService orchestrates series, season, and episode read operations.
type SeriesService struct {
	seriesRepo  *repository.SeriesRepo
	seasonRepo  *repository.SeasonRepo
	episodeRepo *repository.EpisodeRepo
}

// NewSeriesService creates a new series service.
func NewSeriesService(
	seriesRepo *repository.SeriesRepo,
	seasonRepo *repository.SeasonRepo,
	episodeRepo *repository.EpisodeRepo,
) *SeriesService {
	return &SeriesService{
		seriesRepo:  seriesRepo,
		seasonRepo:  seasonRepo,
		episodeRepo: episodeRepo,
	}
}

func (s *SeriesService) ListFiltered(ctx context.Context, filter model.SeriesListFilter) ([]model.SeriesListItem, error) {
	return s.seriesRepo.ListFiltered(ctx, filter)
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
	return s.episodeRepo.ListBySeasonID(ctx, seasonID)
}
