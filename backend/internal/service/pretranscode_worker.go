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
	"syscall"
	"time"

	"github.com/thawng/velox/internal/model"
)

// pickAudioRemuxJob picks the next queued job for the audio-remux profile (video_codec='copy').
func (s *PretranscodeService) pickAudioRemuxJob(ctx context.Context) (*model.PretranscodeQueueItem, error) {
	profile, err := s.repo.GetAudioRemuxProfile(ctx)
	if err != nil || profile == nil {
		return nil, err
	}
	return s.repo.PickNextJobForProfile(ctx, profile.ID)
}

// EnqueueAudioRemux enqueues audio-remux jobs for all non-AAC media files.
func (s *PretranscodeService) EnqueueAudioRemux(ctx context.Context) {
	profile, err := s.repo.GetAudioRemuxProfile(ctx)
	if err != nil || profile == nil {
		return
	}
	files, err := s.repo.ListNonAACMediaFiles(ctx, profile.ID)
	if err != nil {
		log.Printf("pretranscode: list non-AAC files: %v", err)
		return
	}
	if len(files) == 0 {
		return
	}
	enqueued := 0
	for _, f := range files {
		if err := s.repo.EnqueueJob(ctx, f.ID, profile.ID, 100); err != nil {
			continue
		}
		enqueued++
	}
	if enqueued > 0 {
		log.Printf("pretranscode: enqueued %d audio-remux jobs", enqueued)
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

	isAudioRemux := profile.VideoCodec == "copy"

	// For video pretranscode: skip if source resolution < profile height (no upscale)
	if !isAudioRemux && mf.Height < profile.Height {
		log.Printf("pretranscode: skip %s — source %dp < profile %dp", filepath.Base(mf.FilePath), mf.Height, profile.Height)
		_ = s.repo.CompleteJob(ctx, job.ID, "done")
		return
	}

	// For audio remux: skip if audio is already AAC
	if isAudioRemux && strings.EqualFold(mf.AudioCodec, "aac") {
		log.Printf("pretranscode: skip %s — audio already AAC", filepath.Base(mf.FilePath))
		_ = s.repo.CompleteJob(ctx, job.ID, "done")
		return
	}

	// For audio remux: detect if source video codec is not browser-compatible.
	// HEVC/VP9/AV1 sources need full transcode (video → H.264) instead of copy.
	needsVideoTranscode := false
	if isAudioRemux {
		vc := strings.ToLower(mf.VideoCodec)
		isH264 := vc == "h264" || vc == "avc" || vc == "avc1"
		if !isH264 {
			needsVideoTranscode = true
		}
	}

	// For video pretranscode: skip if source is already H.264+AAC at same or lower resolution
	if !isAudioRemux && s.shouldSkipEncode(mf, profile) {
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

	// Check disk space before encoding
	videoBitrate := profile.VideoBitrate
	if videoBitrate == 0 && needsVideoTranscode {
		videoBitrate = estimateBitrateForHeight(mf.Height)
	}
	estimatedSize := int64(float64(videoBitrate+profile.AudioBitrate) * mf.Duration / 8 * 1000)
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
	var errEncode error
	if isAudioRemux && !needsVideoTranscode {
		errEncode = s.runAudioRemux(ctx, mf.FilePath, outputPath, profile)
	} else if isAudioRemux && needsVideoTranscode {
		errEncode = s.runUniversalTranscode(ctx, mf.FilePath, outputPath, profile, mf.Height)
	} else {
		errEncode = s.encodeFile(ctx, mf.FilePath, outputPath, profile, mf.AudioCodec)
	}

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

	// Notify when batch is done
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

// runAudioRemux remuxes a media file: copies video stream, transcodes audio to AAC.
func (s *PretranscodeService) runAudioRemux(ctx context.Context, inputPath, outputPath string, profile *model.PretranscodeProfile) error {
	args := []string{
		"-hide_banner", "-loglevel", "error", "-stats", "-y",
		"-probesize", "5000000", "-analyzeduration", "10000000",
		"-i", inputPath,
		"-map", "0:v:0", "-map", "0:a:0",
		"-c:v", "copy",
		"-c:a", "aac", "-b:a", fmt.Sprintf("%dk", profile.AudioBitrate), "-ac", "2",
		"-movflags", "+faststart",
		outputPath,
	}
	cmd := niceFFmpeg(ctx, args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runUniversalTranscode transcodes non-H.264 sources to H.264+AAC at source resolution.
func (s *PretranscodeService) runUniversalTranscode(ctx context.Context, inputPath, outputPath string, profile *model.PretranscodeProfile, sourceHeight int) error {
	bitrate := estimateBitrateForHeight(sourceHeight)
	// Try HW accel first
	if s.hwAccel != "" {
		err := s.runUniversalTranscodeWith(ctx, inputPath, outputPath, profile, sourceHeight, bitrate, s.hwAccel)
		if err == nil {
			return nil
		}
		log.Printf("pretranscode: HW universal transcode failed (%s), retrying software: %v", s.hwAccel, err)
		_ = os.Remove(outputPath)
	}
	return s.runUniversalTranscodeWith(ctx, inputPath, outputPath, profile, sourceHeight, bitrate, "")
}

func (s *PretranscodeService) runUniversalTranscodeWith(ctx context.Context, inputPath, outputPath string, profile *model.PretranscodeProfile, sourceHeight, bitrate int, hwAccel string) error {
	args := []string{"-hide_banner", "-loglevel", "error", "-stats", "-y"}

	switch hwAccel {
	case "vaapi":
		args = append(args, "-vaapi_device", "/dev/dri/renderD128")
	case "nvenc":
		args = append(args, "-hwaccel", "cuda", "-hwaccel_output_format", "cuda")
	}

	args = append(args, "-i", inputPath)

	switch hwAccel {
	case "vaapi":
		args = append(args, "-vf", "format=nv12,hwupload,scale_vaapi=format=nv12", "-c:v", "h264_vaapi", "-b:v", fmt.Sprintf("%dk", bitrate))
	case "nvenc":
		args = append(args, "-c:v", "h264_nvenc", "-preset", "p4", "-b:v", fmt.Sprintf("%dk", bitrate))
	default:
		args = append(args, "-c:v", "libx264", "-preset", "medium", "-crf", "20")
	}

	args = append(args, "-c:a", "aac", "-b:a", fmt.Sprintf("%dk", profile.AudioBitrate), "-ac", "2",
		"-movflags", "+faststart", "-map", "0:v:0", "-map", "0:a:0", outputPath)

	cmd := niceFFmpeg(ctx, args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// estimateBitrateForHeight returns a rough video bitrate (kbps) for H.264 at a given height.
func estimateBitrateForHeight(height int) int {
	switch {
	case height >= 2160:
		return 20000
	case height >= 1440:
		return 12000
	case height >= 1080:
		return 8000
	case height >= 720:
		return 4000
	case height >= 480:
		return 1500
	default:
		return 800
	}
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

	cmd := niceFFmpeg(ctx, args...)

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

// niceFFmpeg wraps an FFmpeg command with nice -n 19 for lowest CPU priority.
// Pretranscode runs in background — it should never starve NAS or realtime transcode.
func niceFFmpeg(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, "nice", append([]string{"-n", "19", "ffmpeg"}, args...)...)
}

// diskFreeSpace returns free bytes on the filesystem containing path.
func diskFreeSpace(path string) int64 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0
	}
	return int64(stat.Bavail) * int64(stat.Bsize)
}
