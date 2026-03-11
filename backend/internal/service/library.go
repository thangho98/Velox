package service

import (
	"context"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/scanner"
)

type LibraryService struct {
	repo    *repository.LibraryRepo
	scanner *scanner.Scanner
}

func NewLibraryService(repo *repository.LibraryRepo, scanner *scanner.Scanner) *LibraryService {
	return &LibraryService{repo: repo, scanner: scanner}
}

func (s *LibraryService) List(ctx context.Context) ([]model.Library, error) {
	return s.repo.List(ctx)
}

func (s *LibraryService) Create(ctx context.Context, name, path, libType string) (*model.Library, error) {
	return s.repo.Create(ctx, name, path, libType)
}

func (s *LibraryService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *LibraryService) Scan(ctx context.Context, id int64) error {
	return s.scanner.ScanLibrary(ctx, id)
}
