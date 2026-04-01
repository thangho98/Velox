package service

import (
	"context"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// CatalogService orchestrates catalog lookups used by browse/search views.
type CatalogService struct {
	mediaRepo  *repository.MediaRepo
	seriesRepo *repository.SeriesRepo
	genreRepo  *repository.GenreRepo
	personRepo *repository.PersonRepo
}

// NewCatalogService creates a new catalog service.
func NewCatalogService(
	mediaRepo *repository.MediaRepo,
	seriesRepo *repository.SeriesRepo,
	genreRepo *repository.GenreRepo,
	personRepo *repository.PersonRepo,
) *CatalogService {
	return &CatalogService{
		mediaRepo:  mediaRepo,
		seriesRepo: seriesRepo,
		genreRepo:  genreRepo,
		personRepo: personRepo,
	}
}

func (s *CatalogService) ListGenres(ctx context.Context, typeFilter string) ([]model.Genre, error) {
	if typeFilter != "" {
		return s.genreRepo.ListWithFilter(ctx, typeFilter)
	}
	return s.genreRepo.List(ctx)
}

func (s *CatalogService) ListMediaGenres(ctx context.Context, id int64) ([]model.Genre, error) {
	return s.genreRepo.ListByMediaID(ctx, id)
}

func (s *CatalogService) ListMediaCredits(ctx context.Context, id int64) ([]model.CreditWithPerson, error) {
	return s.personRepo.ListCreditsByMedia(ctx, id)
}

func (s *CatalogService) ListSeriesGenres(ctx context.Context, id int64) ([]model.Genre, error) {
	return s.genreRepo.ListBySeriesID(ctx, id)
}

func (s *CatalogService) ListSeriesCredits(ctx context.Context, id int64) ([]model.CreditWithPerson, error) {
	return s.personRepo.ListCreditsBySeries(ctx, id)
}

func (s *CatalogService) Search(ctx context.Context, query string, limit int) (*model.SearchResult, error) {
	movieResults, err := s.mediaRepo.ListFiltered(ctx, model.MediaListFilter{
		Search:    query,
		MediaType: "movie",
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}

	seriesResults, err := s.seriesRepo.ListFiltered(ctx, model.SeriesListFilter{
		Search: query,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	result := &model.SearchResult{
		Movies: movieResults,
		Series: seriesResults,
	}
	if result.Movies == nil {
		result.Movies = []model.MediaListItem{}
	}
	if result.Series == nil {
		result.Series = []model.SeriesListItem{}
	}

	return result, nil
}
