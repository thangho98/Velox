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
	pretranscodeSvc *PretranscodeService
	imagemetaSvc    libraryImagemetaCoordinator
}

// libraryImagemetaCoordinator is the subset of imagemeta.Service the library
// scan path needs. Declared here (vs. importing the concrete type) to keep the
// service testable without pulling imagemeta into tests.
type libraryImagemetaCoordinator interface {
	ComputeBatch(ctx context.Context, paths []string) map[string]error
	InvalidatePaths(ctx context.Context, paths []string) error
}

func NewLibraryService(repo *repository.LibraryRepo, scanJobRepo *repository.ScanJobRepo, pipeline *scanner.Pipeline) *LibraryService {
	return &LibraryService{repo: repo, scanJobRepo: scanJobRepo, pipeline: pipeline}
}

func (s *LibraryService) SetNotificationService(svc *NotificationService) {
	s.notificationSvc = svc
}

func (s *LibraryService) SetPretranscodeService(svc *PretranscodeService) {
	s.pretranscodeSvc = svc
}

// SetImagemetaService wires the blurhash/dimensions service so post-scan hooks
// can backfill missing image metadata for the library.
func (s *LibraryService) SetImagemetaService(svc libraryImagemetaCoordinator) {
	s.imagemetaSvc = svc
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
		// Enqueue audio-remux for any new non-AAC files discovered by scan
		if s.pretranscodeSvc != nil {
			s.pretranscodeSvc.EnqueueAudioRemux(bgCtx)
		}
		// Backfill blurhash / dimensions for any image paths this library owns.
		// Incremental scan only picks up paths without a computed row (Compute
		// is idempotent). Force scan invalidates existing rows first so every
		// image is recomputed from scratch.
		if s.imagemetaSvc != nil {
			go s.backfillLibraryImages(context.Background(), id, force)
		}
	}()

	return job, nil
}

// backfillLibraryImages enumerates every image path owned by a library and
// drives the imagemeta worker to compute missing (or, when force=true, all)
// blurhash + dimension rows. Runs in its own goroutine off the scan path.
func (s *LibraryService) backfillLibraryImages(ctx context.Context, libID int64, force bool) {
	paths, err := s.repo.AllImagePaths(ctx, libID)
	if err != nil {
		log.Printf("scan library %d: list image paths: %v", libID, err)
		return
	}
	if len(paths) == 0 {
		return
	}
	if force {
		if err := s.imagemetaSvc.InvalidatePaths(ctx, paths); err != nil {
			// Log and continue — Compute() is still safe, it just won't recompute
			// paths whose rows we failed to delete.
			log.Printf("scan library %d: invalidate image metadata: %v", libID, err)
		}
	}
	errs := s.imagemetaSvc.ComputeBatch(ctx, paths)
	log.Printf("scan library %d blurhash: processed=%d failed=%d force=%t",
		libID, len(paths), len(errs), force)
}

func (s *LibraryService) ScanJobs(ctx context.Context, libraryID int64) ([]model.ScanJob, error) {
	return s.scanJobRepo.ListByLibrary(ctx, libraryID, 10)
}
