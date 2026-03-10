package service

import (
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

func (s *LibraryService) List() ([]model.Library, error) {
	return s.repo.List()
}

func (s *LibraryService) Create(name, path string) (*model.Library, error) {
	return s.repo.Create(name, path)
}

func (s *LibraryService) Delete(id int64) error {
	return s.repo.Delete(id)
}

func (s *LibraryService) Scan(id int64) error {
	return s.scanner.ScanLibrary(id)
}
