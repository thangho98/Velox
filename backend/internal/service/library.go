package service

import (
	"context"
	"fmt"
	"log"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/scanner"
)

type LibraryService struct {
	repo            *repository.LibraryRepo
	scanJobRepo     *repository.ScanJobRepo
	pipeline        *scanner.Pipeline
	notificationSvc *NotificationService
}

func NewLibraryService(repo *repository.LibraryRepo, scanJobRepo *repository.ScanJobRepo, pipeline *scanner.Pipeline) *LibraryService {
	return &LibraryService{repo: repo, scanJobRepo: scanJobRepo, pipeline: pipeline}
}

func (s *LibraryService) SetNotificationService(svc *NotificationService) {
	s.notificationSvc = svc
}

func (s *LibraryService) List(ctx context.Context) ([]model.Library, error) {
	return s.repo.List(ctx)
}

func (s *LibraryService) Create(ctx context.Context, name, libType string, paths []string) (*model.Library, error) {
	return s.repo.Create(ctx, name, libType, paths)
}

func (s *LibraryService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// Scan creates a scan job and runs the pipeline asynchronously.
// Returns the queued job immediately so the caller can poll status.
func (s *LibraryService) Scan(ctx context.Context, id int64, force bool) (*model.ScanJob, error) {
	job, err := s.pipeline.CreateJob(ctx, id)
	if err != nil {
		return nil, err
	}

	go func() {
		bgCtx := context.Background()
		runErr := s.pipeline.RunJob(bgCtx, job, force)
		if runErr != nil {
			log.Printf("scan library %d job %d: %v", id, job.ID, runErr)
		}
		// RunJob populates job.TotalFiles, job.NewFiles, job.Errors after completion
		if s.notificationSvc != nil {
			libName := fmt.Sprintf("Library #%d", id)
			if lib, err := s.repo.GetByID(bgCtx, id); err == nil {
				libName = lib.Name
			}
			if err := s.notificationSvc.NotifyScanComplete(bgCtx, nil, id, libName, job.TotalFiles, job.NewFiles, job.Errors); err != nil {
				log.Printf("scan notify library %d: %v", id, err)
			}
		}
	}()

	return job, nil
}

func (s *LibraryService) ScanJobs(ctx context.Context, libraryID int64) ([]model.ScanJob, error) {
	return s.scanJobRepo.ListByLibrary(ctx, libraryID, 10)
}
