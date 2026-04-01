package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

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

// GetStatus returns the current status using the service's own repos.
func (s *PretranscodeService) GetStatus(ctx context.Context) (*model.PretranscodeStatus, error) {
	return s.GetStatusWith(ctx, s.repo, s.settingsRepo)
}

// GetStatusWith returns the current status using the provided repos.
// This allows HTTP handlers to query status via the main DB connection
// without competing with the background scheduler's dedicated DB.
func (s *PretranscodeService) GetStatusWith(ctx context.Context, repo *repository.PretranscodeRepo, settingsRepo *repository.AppSettingsRepo) (*model.PretranscodeStatus, error) {
	enabled, _ := settingsRepo.Get(ctx, model.SettingPretranscodeEnabled)
	schedule, _ := settingsRepo.Get(ctx, model.SettingPretranscodeSchedule)
	concurrencyStr, _ := settingsRepo.Get(ctx, model.SettingPretranscodeConcurrency)
	concurrency := 1
	if n, err := strconv.Atoi(concurrencyStr); err == nil && n > 0 {
		concurrency = n
	}

	total, queued, encoding, done, failed, err := repo.QueueStats(ctx)
	if err != nil {
		return nil, err
	}

	diskUsed, _ := repo.TotalDiskUsed(ctx)

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
	cmd := niceFFmpeg(ctx, "-y",
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
