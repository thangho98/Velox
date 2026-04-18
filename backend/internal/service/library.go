package service

import (
	"context"
	"fmt"
	"log"

	"github.com/thawng/velox/internal/cloudstorage"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/scanner"
)

type LibraryService struct {
	repo                *repository.LibraryRepo
	scanJobRepo         *repository.ScanJobRepo
	pipeline            *scanner.Pipeline
	notificationSvc     *NotificationService
	pretranscodeSvc     *PretranscodeService
	imagemetaSvc        libraryImagemetaCoordinator
	cloudRegistry       *cloudstorage.Registry
	storageProviderRepo *repository.StorageProviderRepo
	cloudScanner        *scanner.Scanner
	cloudSecret         []byte
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
	if isNilInterface(svc) {
		s.imagemetaSvc = nil
		return
	}
	s.imagemetaSvc = svc
}

// SetCloud wires the cloud storage dependencies for cloud library scanning.
func (s *LibraryService) SetCloud(registry *cloudstorage.Registry, spRepo *repository.StorageProviderRepo, cloudScanner *scanner.Scanner, secret []byte) {
	s.cloudRegistry = registry
	s.storageProviderRepo = spRepo
	s.cloudScanner = cloudScanner
	s.cloudSecret = secret
}

func (s *LibraryService) List(ctx context.Context) ([]model.Library, error) {
	return s.repo.List(ctx)
}

func (s *LibraryService) Create(ctx context.Context, name, libType string, paths []string) (*model.Library, error) {
	return s.repo.Create(ctx, name, libType, paths)
}

// CreateCloud creates a library backed by a cloud storage provider. The caller
// is responsible for validating the URL shape against the provider before
// persisting — the repo layer trusts its inputs.
func (s *LibraryService) CreateCloud(ctx context.Context, name, libType string, storageProviderID int64, sourceURL string) (*model.Library, error) {
	return s.repo.CreateCloud(ctx, name, libType, storageProviderID, sourceURL)
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

		lib, err := s.repo.GetByID(bgCtx, id)
		if err != nil {
			log.Printf("scan library %d: fetch failed: %v", id, err)
			s.scanJobRepo.Fail(bgCtx, job.ID, "Failed to fetch library")
			return
		}

		if lib.IsCloudLibrary() {
			s.runCloudScan(bgCtx, lib, job, force)
		} else {
			runErr := s.pipeline.RunJob(bgCtx, job, force)
			if runErr != nil {
				log.Printf("scan library %d job %d: %v", id, job.ID, runErr)
			}
		}

		// Notify after completion
		if s.notificationSvc != nil {
			libName := lib.Name
			if err := s.notificationSvc.NotifyScanComplete(bgCtx, nil, id, libName, job.TotalFiles, job.NewFiles, job.Errors); err != nil {
				log.Printf("scan notify library %d: %v", id, err)
			}
		}

		// Enqueue audio-remux for any new non-AAC files discovered by scan (only relevant for local)
		if !lib.IsCloudLibrary() && s.pretranscodeSvc != nil {
			s.pretranscodeSvc.EnqueueAudioRemux(bgCtx)
		}

		// Backfill blurhash / dimensions for any image paths this library owns.
		if s.imagemetaSvc != nil {
			go s.backfillLibraryImages(context.Background(), id, force)
		}
	}()

	return job, nil
}

func (s *LibraryService) runCloudScan(ctx context.Context, lib *model.Library, job *model.ScanJob, force bool) {
	if s.cloudRegistry == nil || s.cloudScanner == nil || s.storageProviderRepo == nil {
		_ = s.scanJobRepo.Fail(ctx, job.ID, "Cloud storage features are not fully configured")
		return
	}

	if err := s.scanJobRepo.Start(ctx, job.ID); err != nil {
		log.Printf("scan library %d: start job failed: %v", lib.ID, err)
		return
	}

	provRow, err := s.storageProviderRepo.GetByID(ctx, *lib.StorageProviderID)
	if err != nil {
		_ = s.scanJobRepo.Fail(ctx, job.ID, fmt.Sprintf("fetch provider: %v", err))
		return
	}

	driver, err := s.cloudRegistry.Get(provRow.ProviderType)
	if err != nil {
		_ = s.scanJobRepo.Fail(ctx, job.ID, fmt.Sprintf("get provider driver: %v", err))
		return
	}

	creds, err := cloudstorage.DecryptCredentials(s.cloudSecret, provRow.CredentialsEncrypted)
	if err != nil {
		_ = s.scanJobRepo.Fail(ctx, job.ID, fmt.Sprintf("decrypt credentials: %v", err))
		return
	}

	provider, err := driver.NewProvider(creds)
	if err != nil {
		_ = s.scanJobRepo.Fail(ctx, job.ID, fmt.Sprintf("instantiate provider: %v", err))
		return
	}

	var sourceURL string
	if lib.SourceURL != nil {
		sourceURL = *lib.SourceURL
	}
	folderID, err := provider.ParseFolderURL(sourceURL)
	if err != nil {
		_ = s.scanJobRepo.Fail(ctx, job.ID, fmt.Sprintf("parse folder URL: %v", err))
		return
	}

	totalFiles, newFiles, errCount, err := s.cloudScanner.ScanCloudLibrary(ctx, lib.ID, provider, folderID, force)
	if err != nil {
		_ = s.scanJobRepo.Fail(ctx, job.ID, err.Error())
		return
	}

	if err := s.scanJobRepo.Complete(ctx, job.ID, totalFiles, newFiles, errCount, ""); err != nil {
		log.Printf("scan library %d: completing job failed: %v", lib.ID, err)
	}
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
