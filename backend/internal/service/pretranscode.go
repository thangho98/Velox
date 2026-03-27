package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// PretranscodeService manages offline encoding of media files.
type PretranscodeService struct {
	repo            *repository.PretranscodeRepo
	mediaFileRepo   *repository.MediaFileRepo
	settingsRepo    *repository.AppSettingsRepo
	libraryRepo     *repository.LibraryRepo
	notificationSvc *NotificationService

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

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel
	s.stopCh = make(chan struct{})
	s.running.Store(true)

	go s.schedulerLoop(ctx)
	log.Println("pretranscode: scheduler started")
}

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
func (s *PretranscodeService) Pause() { s.paused.Store(true) }

// Resume resumes the scheduler.
func (s *PretranscodeService) Resume() { s.paused.Store(false) }

// IsPaused returns whether the scheduler is paused.
func (s *PretranscodeService) IsPaused() bool { return s.paused.Load() }

// IsRunning returns whether the scheduler is active.
func (s *PretranscodeService) IsRunning() bool { return s.running.Load() }

func (s *PretranscodeService) schedulerLoop(ctx context.Context) {
	defer close(s.stopCh)

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

		// Check if feature is enabled
		enabled, _ := s.settingsRepo.Get(ctx, model.SettingPretranscodeEnabled)
		if enabled != "true" {
			s.sleep(ctx, 30*time.Second)
			continue
		}

		// Check schedule
		if !s.isInSchedule(ctx) {
			s.sleep(ctx, 60*time.Second)
			continue
		}

		// Pick next job
		job, err := s.repo.PickNextJob(ctx)
		if err != nil {
			log.Printf("pretranscode: pick job error: %v", err)
			s.sleep(ctx, 10*time.Second)
			continue
		}
		if job == nil {
			s.sleep(ctx, 30*time.Second)
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

func (s *PretranscodeService) processJob(ctx context.Context, job *model.PretranscodeQueueItem) {
	profile, err := s.repo.GetProfile(ctx, job.ProfileID)
	if err != nil {
		log.Printf("pretranscode: get profile %d: %v", job.ProfileID, err)
		_ = s.repo.CompleteJob(ctx, job.ID, "failed")
		return
	}

	mf, err := s.mediaFileRepo.GetByID(ctx, job.MediaFileID)
	if err != nil {
		log.Printf("pretranscode: get media file %d: %v", job.MediaFileID, err)
		_ = s.repo.CompleteJob(ctx, job.ID, "failed")
		return
	}

	// Skip if source resolution < profile height (no upscale)
	if mf.Height < profile.Height {
		log.Printf("pretranscode: skip %s — source %dp < profile %dp", filepath.Base(mf.FilePath), mf.Height, profile.Height)
		_ = s.repo.CompleteJob(ctx, job.ID, "done")
		return
	}

	// Skip if source is already H.264+AAC at same or lower resolution
	if s.shouldSkipEncode(mf, profile) {
		log.Printf("pretranscode: skip %s — already compatible", filepath.Base(mf.FilePath))
		_ = s.repo.CompleteJob(ctx, job.ID, "done")
		return
	}

	// Check source file exists
	if _, err := os.Stat(mf.FilePath); os.IsNotExist(err) {
		log.Printf("pretranscode: source missing: %s", mf.FilePath)
		_ = s.repo.CompleteJob(ctx, job.ID, "failed")
		return
	}

	// Check disk space before encoding (rough estimate: need at least bitrate * duration bytes)
	estimatedSize := int64(float64(profile.VideoBitrate+profile.AudioBitrate) * mf.Duration / 8 * 1000)
	freeSpace := diskFreeSpace(s.outputBaseDir)
	if freeSpace > 0 && freeSpace < estimatedSize*2 {
		log.Printf("pretranscode: disk low (free: %d MB, need ~%d MB) — pausing", freeSpace/1024/1024, estimatedSize/1024/1024)
		s.Pause()
		_ = s.repo.CompleteJob(ctx, job.ID, "failed")
		return
	}

	s.currentFile.Store(filepath.Base(mf.FilePath))
	s.currentSpeed.Store("")

	outputDir := filepath.Join(s.OutputDir(), strconv.FormatInt(mf.ID, 10))
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("pretranscode: mkdir %s: %v", outputDir, err)
		_ = s.repo.CompleteJob(ctx, job.ID, "failed")
		return
	}
	outputPath := filepath.Join(outputDir, profile.Name+".mp4")

	// Create/update file record as encoding
	now := time.Now().UTC().Format(time.RFC3339)
	ptFile := &model.PretranscodeFile{
		MediaFileID: mf.ID,
		ProfileID:   profile.ID,
		FilePath:    outputPath,
		Status:      "encoding",
		StartedAt:   now,
	}
	fileID, err := s.repo.UpsertFile(ctx, ptFile)
	if err != nil {
		log.Printf("pretranscode: upsert file record: %v", err)
		_ = s.repo.CompleteJob(ctx, job.ID, "failed")
		return
	}

	// Run FFmpeg encode
	errEncode := s.encodeFile(ctx, mf.FilePath, outputPath, profile, mf.AudioCodec)

	if errEncode != nil {
		log.Printf("pretranscode: encode failed for %s (%s): %v", filepath.Base(mf.FilePath), profile.Name, errEncode)
		completed := time.Now().UTC().Format(time.RFC3339)
		_ = s.repo.UpdateFileStatus(ctx, fileID, "failed", errEncode.Error(), "", completed)
		_ = s.repo.CompleteJob(ctx, job.ID, "failed")
		_ = os.Remove(outputPath)
		s.currentFile.Store("")
		return
	}

	// Success: update file record
	stat, _ := os.Stat(outputPath)
	fileSize := int64(0)
	if stat != nil {
		fileSize = stat.Size()
	}
	completed := time.Now().UTC().Format(time.RFC3339)
	_ = s.repo.UpdateFileStatus(ctx, fileID, "ready", "", "", completed)
	// Update file_size
	s.repo.UpsertFile(ctx, &model.PretranscodeFile{
		MediaFileID:  mf.ID,
		ProfileID:    profile.ID,
		FilePath:     outputPath,
		FileSize:     fileSize,
		DurationSecs: mf.Duration,
		Status:       "ready",
		StartedAt:    now,
		CompletedAt:  completed,
	})
	_ = s.repo.CompleteJob(ctx, job.ID, "done")

	log.Printf("pretranscode: done %s → %s (%.1f MB)", filepath.Base(mf.FilePath), profile.Name, float64(fileSize)/1024/1024)
	s.currentFile.Store("")
	s.currentSpeed.Store("")

	// Notify when batch is done (check if queue is empty)
	if s.notificationSvc != nil {
		_, queued, _, _, _, _ := s.repo.QueueStats(ctx)
		if queued == 0 {
			total, _, _, done, failed, _ := s.repo.QueueStats(ctx)
			msg := fmt.Sprintf("Pre-transcode batch complete: %d/%d done", done, total)
			if failed > 0 {
				msg += fmt.Sprintf(", %d failed", failed)
			}
			_ = s.notificationSvc.NotifyTranscodeComplete(ctx, 0, 0, msg, failed == 0, "Pre-transcode", 0)
		}
	}
}

func (s *PretranscodeService) shouldSkipEncode(mf *model.MediaFile, profile *model.PretranscodeProfile) bool {
	vc := strings.ToLower(mf.VideoCodec)
	ac := strings.ToLower(mf.AudioCodec)
	isH264 := vc == "h264" || vc == "avc" || vc == "avc1"
	isAAC := ac == "aac"
	return isH264 && isAAC && mf.Height <= profile.Height
}

func (s *PretranscodeService) encodeFile(ctx context.Context, inputPath, outputPath string, profile *model.PretranscodeProfile, sourceAudioCodec string) error {
	// Try HW accel first, fallback to software
	if s.hwAccel != "" {
		err := s.runFFmpeg(ctx, inputPath, outputPath, profile, s.hwAccel, sourceAudioCodec)
		if err == nil {
			return nil
		}
		log.Printf("pretranscode: HW encode failed (%s), retrying software: %v", s.hwAccel, err)
		_ = os.Remove(outputPath)
	}
	return s.runFFmpeg(ctx, inputPath, outputPath, profile, "", sourceAudioCodec)
}

func (s *PretranscodeService) runFFmpeg(ctx context.Context, inputPath, outputPath string, profile *model.PretranscodeProfile, hwAccel, sourceAudioCodec string) error {
	args := []string{"-hide_banner", "-loglevel", "error", "-stats", "-y"}

	// Input args (HW accel init)
	switch hwAccel {
	case "vaapi":
		args = append(args, "-vaapi_device", "/dev/dri/renderD128")
	case "nvenc":
		args = append(args, "-hwaccel", "cuda", "-hwaccel_output_format", "cuda")
	case "qsv":
		args = append(args, "-hwaccel", "qsv", "-hwaccel_output_format", "qsv")
	case "videotoolbox":
		args = append(args, "-hwaccel", "videotoolbox")
	}

	args = append(args, "-i", inputPath)

	// Video filter (scale)
	switch hwAccel {
	case "vaapi":
		args = append(args, "-vf", fmt.Sprintf("format=nv12,hwupload,scale_vaapi=-2:%d", profile.Height))
	case "nvenc":
		args = append(args, "-vf", fmt.Sprintf("scale_cuda=-2:%d", profile.Height))
	case "qsv":
		args = append(args, "-vf", fmt.Sprintf("scale_qsv=-2:%d", profile.Height))
	default:
		args = append(args, "-vf", fmt.Sprintf("scale=-2:%d", profile.Height))
	}

	// Video codec
	switch hwAccel {
	case "vaapi":
		args = append(args, "-c:v", "h264_vaapi")
	case "nvenc":
		args = append(args, "-c:v", "h264_nvenc", "-preset", "p4", "-tune", "hq")
	case "qsv":
		args = append(args, "-c:v", "h264_qsv", "-preset", "medium")
	case "videotoolbox":
		args = append(args, "-c:v", "h264_videotoolbox")
	default:
		args = append(args, "-c:v", "libx264", "-preset", "medium", "-crf", "22")
	}

	// Bitrate (for HW encoders that don't support CRF)
	if hwAccel != "" {
		args = append(args, "-b:v", fmt.Sprintf("%dk", profile.VideoBitrate))
	}

	// Audio: copy if source is already the target codec, otherwise transcode
	srcAudio := strings.ToLower(sourceAudioCodec)
	if srcAudio == profile.AudioCodec || srcAudio == "aac" && profile.AudioCodec == "aac" {
		args = append(args, "-c:a", "copy")
	} else {
		args = append(args, "-c:a", profile.AudioCodec, "-b:a", fmt.Sprintf("%dk", profile.AudioBitrate), "-ac", "2")
	}

	// Output flags
	args = append(args, "-movflags", "+faststart", "-map", "0:v:0", "-map", "0:a:0", outputPath)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errOutput := stderr.String()
		// Extract last useful line from stderr
		lines := strings.Split(strings.TrimSpace(errOutput), "\n")
		errMsg := err.Error()
		if len(lines) > 0 {
			last := lines[len(lines)-1]
			if len(last) > 200 {
				last = last[:200]
			}
			errMsg = last
		}
		return fmt.Errorf("ffmpeg: %s", errMsg)
	}
	return nil
}

// EnqueueLibrary enqueues all eligible media files in a library for encoding.
func (s *PretranscodeService) EnqueueLibrary(ctx context.Context, libraryID int64) (int, error) {
	profiles, err := s.repo.ListEnabledProfiles(ctx)
	if err != nil {
		return 0, fmt.Errorf("listing profiles: %w", err)
	}
	if len(profiles) == 0 {
		return 0, fmt.Errorf("no enabled profiles")
	}

	total := 0
	for _, p := range profiles {
		files, err := s.repo.ListMediaFilesForEnqueue(ctx, libraryID, p.ID, p.Height)
		if err != nil {
			return total, fmt.Errorf("listing files for profile %s: %w", p.Name, err)
		}
		for _, f := range files {
			if err := s.repo.EnqueueJob(ctx, f.ID, p.ID, 0); err != nil {
				log.Printf("pretranscode: enqueue %d/%s: %v", f.ID, p.Name, err)
				continue
			}
			total++
		}
	}
	return total, nil
}

// EnqueueAllLibraries enqueues all libraries.
func (s *PretranscodeService) EnqueueAllLibraries(ctx context.Context) (int, error) {
	libs, err := s.libraryRepo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("listing libraries: %w", err)
	}
	total := 0
	for _, lib := range libs {
		n, err := s.EnqueueLibrary(ctx, lib.ID)
		if err != nil {
			log.Printf("pretranscode: enqueue library %d: %v", lib.ID, err)
		}
		total += n
	}
	return total, nil
}

// CancelAll cancels queued jobs, kills current FFmpeg process, and pauses.
func (s *PretranscodeService) CancelAll(ctx context.Context) (int64, error) {
	s.Pause()
	// Cancel the scheduler context to kill any running FFmpeg process
	if s.cancelFn != nil {
		s.cancelFn()
	}
	// Wait for scheduler loop to fully exit (ensures processJob error handler is done)
	if s.running.Load() {
		<-s.stopCh
		s.running.Store(false)
	}
	// Now safe to clean up DB — use background ctx since scheduler ctx is cancelled
	bgCtx := context.Background()
	n, err := s.repo.CancelAllQueued(bgCtx)
	// Reset interrupted encoding jobs back to queued so they can be retried
	_, _ = s.repo.ResetEncodingJobs(bgCtx)
	s.repo.ResetEncodingFiles(bgCtx)
	// Restart scheduler loop (paused — won't pick jobs until Resume)
	s.Start()
	return n, err
}

// GetStatus returns the current status.
func (s *PretranscodeService) GetStatus(ctx context.Context) (*model.PretranscodeStatus, error) {
	enabled, _ := s.settingsRepo.Get(ctx, model.SettingPretranscodeEnabled)
	schedule, _ := s.settingsRepo.Get(ctx, model.SettingPretranscodeSchedule)
	concurrencyStr, _ := s.settingsRepo.Get(ctx, model.SettingPretranscodeConcurrency)
	concurrency := 1
	if n, err := strconv.Atoi(concurrencyStr); err == nil && n > 0 {
		concurrency = n
	}

	total, queued, encoding, done, failed, err := s.repo.QueueStats(ctx)
	if err != nil {
		return nil, err
	}

	diskUsed, _ := s.repo.TotalDiskUsed(ctx)

	cf, _ := s.currentFile.Load().(string)
	cs, _ := s.currentSpeed.Load().(string)

	return &model.PretranscodeStatus{
		Enabled:     enabled == "true",
		Schedule:    schedule,
		Concurrency: concurrency,
		Paused:      s.paused.Load(),
		Total:       total,
		Done:        done,
		Encoding:    encoding,
		Failed:      failed,
		Queued:      queued,
		DiskUsed:    diskUsed,
		CurrentFile: cf,
		Speed:       cs,
	}, nil
}

// EstimateStorage estimates disk usage for pre-transcoding a library.
func (s *PretranscodeService) EstimateStorage(ctx context.Context, libraryID int64) (*model.StorageEstimate, error) {
	profiles, err := s.repo.ListEnabledProfiles(ctx)
	if err != nil {
		return nil, err
	}

	avgDuration, err := s.repo.AvgDurationByLibrary(ctx, libraryID)
	if err != nil {
		return nil, err
	}
	if avgDuration <= 0 {
		avgDuration = 5400 // default 90 min
	}

	estimate := &model.StorageEstimate{}

	for _, p := range profiles {
		count, err := s.repo.CountMediaFilesInLibrary(ctx, libraryID)
		if err != nil {
			return nil, err
		}
		// Estimated size = bitrate (kbps) * duration (s) / 8 * 1000 (bytes)
		bytesPerFile := int64(float64(p.VideoBitrate+p.AudioBitrate) * avgDuration / 8 * 1000)
		totalBytes := bytesPerFile * int64(count)

		estimate.Profiles = append(estimate.Profiles, model.ProfileEstimate{
			ProfileID:   p.ID,
			ProfileName: p.Name,
			Height:      p.Height,
			EstimatedGB: float64(totalBytes) / (1024 * 1024 * 1024),
			FileCount:   count,
		})
		estimate.TotalBytes += totalBytes
		estimate.FileCount = count
	}

	estimate.DiskFreeBytes = diskFreeSpace(s.outputBaseDir)
	return estimate, nil
}

// CleanupAll deletes all pre-transcode files from disk and DB.
func (s *PretranscodeService) CleanupAll(ctx context.Context) (int, error) {
	paths, err := s.repo.DeleteAllFiles(ctx)
	if err != nil {
		return 0, err
	}
	_ = s.repo.ClearQueue(ctx)

	removed := 0
	for _, p := range paths {
		if err := os.Remove(p); err == nil {
			removed++
		}
	}

	// Remove empty directories
	ptDir := s.OutputDir()
	if entries, err := os.ReadDir(ptDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				subDir := filepath.Join(ptDir, e.Name())
				if subEntries, err := os.ReadDir(subDir); err == nil && len(subEntries) == 0 {
					_ = os.Remove(subDir)
				}
			}
		}
	}
	return removed, nil
}

// CleanupByMediaFile removes pre-transcode files for a specific media file.
func (s *PretranscodeService) CleanupByMediaFile(ctx context.Context, mediaFileID int64) error {
	files, err := s.repo.ListReadyFilesByMedia(ctx, mediaFileID)
	if err != nil {
		return err
	}
	for _, f := range files {
		_ = os.Remove(f.FilePath)
	}
	return nil
}

// GetProfile returns a single profile by ID.
func (s *PretranscodeService) GetProfile(ctx context.Context, id int64) (*model.PretranscodeProfile, error) {
	return s.repo.GetProfile(ctx, id)
}

// ListProfiles returns all profiles.
func (s *PretranscodeService) ListProfiles(ctx context.Context) ([]model.PretranscodeProfile, error) {
	return s.repo.ListProfiles(ctx)
}

// SetProfileEnabled toggles a profile.
func (s *PretranscodeService) SetProfileEnabled(ctx context.Context, id int64, enabled bool) error {
	return s.repo.SetProfileEnabled(ctx, id, enabled)
}

// ListReadyFiles returns all ready pre-transcode files for a media file.
func (s *PretranscodeService) ListReadyFiles(ctx context.Context, mediaFileID int64) ([]model.PretranscodeFile, error) {
	return s.repo.ListReadyFilesByMedia(ctx, mediaFileID)
}

// ListReadyFilesWithProfiles returns ready files joined with their profile metadata in one query.
func (s *PretranscodeService) ListReadyFilesWithProfiles(ctx context.Context, mediaFileID int64) ([]repository.ReadyFileWithProfile, error) {
	return s.repo.ListReadyFilesWithProfiles(ctx, mediaFileID)
}

// RemuxFromHLS copies existing HLS transcode segments into a pretranscode MP4.
// Called after realtime transcode — "transcode once, instant forever".
// No-op if pretranscode disabled, no matching profile, or file already exists.
func (s *PretranscodeService) RemuxFromHLS(ctx context.Context, mediaFileID int64, height int, hlsPlaylist string) {
	// Check pretranscode enabled
	enabled, _ := s.settingsRepo.Get(ctx, model.SettingPretranscodeEnabled)
	if enabled != "true" {
		return
	}

	profile, err := s.repo.GetProfileByHeight(ctx, height)
	if err != nil || profile == nil {
		return
	}

	// Check if already ready
	existing, _ := s.repo.GetFileByMediaAndProfile(ctx, mediaFileID, profile.ID)
	if existing != nil {
		return
	}

	// Verify HLS playlist exists
	if _, err := os.Stat(hlsPlaylist); err != nil {
		return
	}

	// Output path
	outputDir := filepath.Join(s.OutputDir(), fmt.Sprintf("%d", mediaFileID))
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Printf("pretranscode: remux mkdir failed: %v", err)
		return
	}
	outputPath := filepath.Join(outputDir, profile.Name+".mp4")

	// Remux HLS → MP4 (copy streams, no re-encoding — near instant)
	cmd := exec.CommandContext(ctx, "ffmpeg", "-y",
		"-i", hlsPlaylist,
		"-c", "copy",
		"-movflags", "+faststart",
		outputPath,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("pretranscode: remux failed for file %d @ %s: %v\n%s", mediaFileID, profile.Name, err, output)
		os.Remove(outputPath)
		return
	}

	// Get file size
	stat, err := os.Stat(outputPath)
	if err != nil {
		log.Printf("pretranscode: remux stat failed: %v", err)
		return
	}

	// Get duration
	mf, _ := s.mediaFileRepo.GetByID(ctx, mediaFileID)
	var duration float64
	if mf != nil {
		duration = mf.Duration
	}

	// Upsert pretranscode file record
	ptFile := &model.PretranscodeFile{
		MediaFileID:  mediaFileID,
		ProfileID:    profile.ID,
		FilePath:     outputPath,
		FileSize:     stat.Size(),
		DurationSecs: duration,
		Status:       "ready",
		CompletedAt:  time.Now().Format(time.RFC3339),
	}
	if _, err := s.repo.UpsertFile(ctx, ptFile); err != nil {
		log.Printf("pretranscode: remux upsert failed: %v", err)
		os.Remove(outputPath)
		return
	}

	log.Printf("pretranscode: remuxed HLS → %s (%s, %d MB)", profile.Name, filepath.Base(outputPath), stat.Size()/1024/1024)
}

// FindBestPretranscode finds the best pre-transcoded file for a media file + max height.
func (s *PretranscodeService) FindBestPretranscode(ctx context.Context, mediaFileID int64, maxHeight int) (*model.PretranscodeFile, error) {
	files, err := s.repo.ListReadyFilesByMedia(ctx, mediaFileID)
	if err != nil {
		return nil, err
	}

	// files are ordered by height DESC — pick the best that fits
	for _, f := range files {
		profile, err := s.repo.GetProfile(ctx, f.ProfileID)
		if err != nil {
			continue
		}
		if maxHeight <= 0 || profile.Height <= maxHeight {
			return &f, nil
		}
	}
	return nil, nil
}

func (s *PretranscodeService) sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

// diskFreeSpace returns free bytes on the filesystem containing path.
func diskFreeSpace(path string) int64 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0
	}
	return int64(stat.Bavail) * int64(stat.Bsize)
}
