package service

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// PretranscodeService manages offline encoding of media files.
type PretranscodeService struct {
	repo            *repository.PretranscodeRepo
	statusRepo      *repository.PretranscodeRepo
	mediaFileRepo   *repository.MediaFileRepo
	settingsRepo    *repository.AppSettingsRepo
	libraryRepo     *repository.LibraryRepo
	notificationSvc *NotificationService
	transcoder      interface{ TryActiveCount() int } // realtime transcoder — nil-safe

	outputBaseDir string
	hwAccel       string

	mu       sync.Mutex
	paused   atomic.Bool
	running  atomic.Bool
	stopCh   chan struct{}
	cancelFn context.CancelFunc

	// Progress tracking
	currentFile  atomic.Value // string
	currentSpeed atomic.Value // string
}

// NewPretranscodeService creates a new pre-transcode service.
func NewPretranscodeService(
	repo *repository.PretranscodeRepo,
	mediaFileRepo *repository.MediaFileRepo,
	settingsRepo *repository.AppSettingsRepo,
	libraryRepo *repository.LibraryRepo,
	pretranscodePath, hwAccel string,
) *PretranscodeService {
	s := &PretranscodeService{
		repo:          repo,
		mediaFileRepo: mediaFileRepo,
		settingsRepo:  settingsRepo,
		libraryRepo:   libraryRepo,
		outputBaseDir: pretranscodePath,
		hwAccel:       hwAccel,
	}
	s.currentFile.Store("")
	s.currentSpeed.Store("")
	return s
}

// SetNotificationService sets the notification service for progress notifications.
func (s *PretranscodeService) SetNotificationService(svc *NotificationService) {
	s.notificationSvc = svc
}

// SetTranscoder sets the realtime transcoder so pretranscode can yield when users are watching.
func (s *PretranscodeService) SetTranscoder(t interface{ TryActiveCount() int }) {
	s.transcoder = t
}

// SetStatusRepo configures the main-db status repo used by HTTP handlers.
func (s *PretranscodeService) SetStatusRepo(repo *repository.PretranscodeRepo) {
	s.statusRepo = repo
}

// OutputDir returns the base directory for pre-transcode files.
func (s *PretranscodeService) OutputDir() string {
	return s.outputBaseDir
}

// Start begins the background scheduler loop.
func (s *PretranscodeService) Start() {
	if s.running.Load() {
		return
	}

	// Recovery: reset any 'encoding' queue items back to 'queued' (interrupted by restart)
	s.recoverInterruptedJobs()

	// Restore persisted paused state from database.
	// Fail-safe: if DB read fails, keep paused=true to avoid unexpected resume.
	ctx := context.Background()
	pausedVal, err := s.settingsRepo.Get(ctx, model.SettingPretranscodePaused)
	if err != nil {
		log.Printf("pretranscode: failed to read paused state from DB (%v) — keeping paused", err)
		s.paused.Store(true)
	} else {
		s.paused.Store(pausedVal == "true")
		if pausedVal == "true" {
			log.Println("pretranscode: scheduler started (paused)")
		} else {
			log.Println("pretranscode: scheduler started")
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel
	s.stopCh = make(chan struct{})
	s.running.Store(true)

	go s.schedulerLoop(ctx)
}

// Stop gracefully stops the scheduler.
func (s *PretranscodeService) Stop() {
	if !s.running.Load() {
		return
	}
	if s.cancelFn != nil {
		s.cancelFn()
	}
	<-s.stopCh
	s.running.Store(false)
	log.Println("pretranscode: scheduler stopped")
}

// Pause pauses the scheduler (current job finishes, no new jobs picked up).
// Persists to database so the pause state survives restarts.
// Returns error if DB write fails; scheduler state is unchanged on failure (atomic).
func (s *PretranscodeService) Pause() error {
	ctx := context.Background()
	if err := s.settingsRepo.Set(ctx, model.SettingPretranscodePaused, "true"); err != nil {
		return err
	}
	s.paused.Store(true)
	return nil
}

// Resume resumes the scheduler.
// Persists to database so the resume state survives restarts.
// Returns error if DB write fails; scheduler state is unchanged on failure (atomic).
func (s *PretranscodeService) Resume() error {
	ctx := context.Background()
	if err := s.settingsRepo.Set(ctx, model.SettingPretranscodePaused, "false"); err != nil {
		return err
	}
	s.paused.Store(false)
	return nil
}

// IsPaused returns whether the scheduler is paused.
func (s *PretranscodeService) IsPaused() bool { return s.paused.Load() }

// IsRunning returns whether the scheduler is active.
func (s *PretranscodeService) IsRunning() bool { return s.running.Load() }

func (s *PretranscodeService) recoverInterruptedJobs() {
	ctx := context.Background()
	// Reset encoding queue items back to queued
	_, err := s.repo.ResetEncodingJobs(ctx)
	if err != nil {
		log.Printf("pretranscode: recovery reset failed: %v", err)
	}
	// Reset encoding file records back to pending
	s.repo.ResetEncodingFiles(ctx)
}

func (s *PretranscodeService) schedulerLoop(ctx context.Context) {
	defer close(s.stopCh)

	// Auto-enqueue on startup.
	s.EnqueueAudioRemux(ctx)
	enabledVal, _ := s.settingsRepo.Get(ctx, model.SettingPretranscodeEnabled)
	if enabledVal == "true" {
		n, err := s.EnqueueAllLibraries(ctx)
		if err != nil {
			log.Printf("pretranscode: auto-enqueue error: %v", err)
		} else if n > 0 {
			log.Printf("pretranscode: auto-enqueued %d video jobs on startup", n)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if s.paused.Load() {
			s.sleep(ctx, 10*time.Second)
			continue
		}

		// Yield to realtime transcodes — user experience first.
		if s.transcoder != nil && s.transcoder.TryActiveCount() > 0 {
			s.sleep(ctx, 5*time.Second)
			continue
		}

		// Audio-remux jobs always run (not gated by pretranscode_enabled).
		// Pick audio-remux job first (highest priority).
		audioJob, err := s.pickAudioRemuxJob(ctx)
		if err != nil {
			log.Printf("pretranscode: pick audio-remux job error: %v", err)
		}
		if audioJob != nil {
			s.processJob(ctx, audioJob)
			continue
		}

		// Video pretranscode: only when feature is enabled + in schedule.
		enabled, _ := s.settingsRepo.Get(ctx, model.SettingPretranscodeEnabled)
		if enabled != "true" {
			s.sleep(ctx, 30*time.Second)
			continue
		}
		if !s.isInSchedule(ctx) {
			s.sleep(ctx, 60*time.Second)
			continue
		}

		// Pick next video pretranscode job
		job, err := s.repo.PickNextJob(ctx)
		if err != nil {
			log.Printf("pretranscode: pick job error: %v", err)
			s.sleep(ctx, 10*time.Second)
			continue
		}
		if job == nil {
			s.sleep(ctx, 60*time.Second)
			continue
		}

		s.processJob(ctx, job)
	}
}

func (s *PretranscodeService) isInSchedule(ctx context.Context) bool {
	schedule, _ := s.settingsRepo.Get(ctx, model.SettingPretranscodeSchedule)
	switch schedule {
	case "night":
		hour := time.Now().Hour()
		return hour >= 0 && hour < 6
	case "idle":
		// Consider idle if no active transcode (simplified check)
		return true
	default: // "always" or empty
		return true
	}
}

func (s *PretranscodeService) sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
