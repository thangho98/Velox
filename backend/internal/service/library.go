package service

import (
	"context"
	"log"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/scanner"
)

type LibraryService struct {
	repo        *repository.LibraryRepo
	scanJobRepo *repository.ScanJobRepo
	pipeline    *scanner.Pipeline
}

func NewLibraryService(repo *repository.LibraryRepo, scanJobRepo *repository.ScanJobRepo, pipeline *scanner.Pipeline) *LibraryService {
	return &LibraryService{repo: repo, scanJobRepo: scanJobRepo, pipeline: pipeline}
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

// Scan creates a scan job and runs the pipeline asynchronously.
// Returns the queued job immediately so the caller can poll status.
func (s *LibraryService) Scan(ctx context.Context, id int64) (*model.ScanJob, error) {
	job, err := s.pipeline.CreateJob(ctx, id)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := s.pipeline.RunJob(context.Background(), job); err != nil {
			log.Printf("scan library %d job %d: %v", id, job.ID, err)
		}
	}()

	return job, nil
}

func (s *LibraryService) ScanJobs(ctx context.Context, libraryID int64) ([]model.ScanJob, error) {
	return s.scanJobRepo.ListByLibrary(ctx, libraryID, 10)
}
