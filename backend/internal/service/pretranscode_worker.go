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
	"github.com/thawng/velox/internal/scanner"
	"github.com/thawng/velox/internal/transcoder"
	"github.com/thawng/velox/pkg/ffmpegbin"
	"github.com/thawng/velox/pkg/ffprobe"
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

	// Skip cloud files — pretranscode requires local file access
	if scanner.IsCloudPath(mf.FilePath) {
		_ = s.repo.CompleteJob(ctx, job.ID, "done")
		return
	}

	isAudioRemux := profile.VideoCodec == "copy"
	isHDR := ffprobe.IsHDRLike(mf.FilePath)

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
		if err := s.Pause(); err != nil {
			log.Printf("pretranscode: disk low auto-pause: failed to persist: %v", err)
		}
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
		errEncode = s.runUniversalTranscode(ctx, mf, outputPath, profile, isHDR)
	} else {
		errEncode = s.encodeFile(ctx, mf, outputPath, profile, mf.AudioCodec, isHDR)
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

	// Probe encoded output for actual audio metadata so the playback overlay
	// shows what is really being streamed (AAC stereo) instead of source AC3 5.1.
	var audioCodec string
	var audioChannels, audioBitrate, audioSampleRate int
	if probeResult, probeErr := ffprobe.Probe(outputPath); probeErr != nil {
		log.Printf("pretranscode: ffprobe %s failed (non-fatal): %v", filepath.Base(outputPath), probeErr)
	} else if len(probeResult.AudioTracks) > 0 {
		// Use the default track if marked, otherwise the first one.
		picked := probeResult.AudioTracks[0]
		for _, t := range probeResult.AudioTracks {
			if t.IsDefault {
				picked = t
				break
			}
		}
		audioCodec = picked.Codec
		audioChannels = picked.Channels
		audioBitrate = picked.Bitrate
		audioSampleRate = picked.SampleRate
	}

	_ = s.repo.UpdateFileStatus(ctx, fileID, "ready", "", "", completed)
	// Update file_size + audio metadata
	s.repo.UpsertFile(ctx, &model.PretranscodeFile{
		MediaFileID:     mf.ID,
		ProfileID:       profile.ID,
		FilePath:        outputPath,
		FileSize:        fileSize,
		DurationSecs:    mf.Duration,
		Status:          "ready",
		StartedAt:       now,
		CompletedAt:     completed,
		AudioCodec:      audioCodec,
		AudioChannels:   audioChannels,
		AudioBitrate:    audioBitrate,
		AudioSampleRate: audioSampleRate,
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
func (s *PretranscodeService) runUniversalTranscode(ctx context.Context, mf *model.MediaFile, outputPath string, profile *model.PretranscodeProfile, hdr bool) error {
	bitrate := estimateBitrateForHeight(mf.Height)
	if hdr {
		log.Printf("pretranscode: HDR/DV source detected for %s", filepath.Base(mf.FilePath))
		return s.runUniversalTranscodeWith(ctx, mf.FilePath, outputPath, profile, mf.Height, bitrate, "", true, mf)
	}
	// Try HW accel first
	if s.hwAccel != "" {
		err := s.runUniversalTranscodeWith(ctx, mf.FilePath, outputPath, profile, mf.Height, bitrate, s.hwAccel, false, mf)
		if err == nil {
			return nil
		}
		log.Printf("pretranscode: HW universal transcode failed (%s), retrying software: %v", s.hwAccel, err)
		_ = os.Remove(outputPath)
	}
	return s.runUniversalTranscodeWith(ctx, mf.FilePath, outputPath, profile, mf.Height, bitrate, "", false, mf)
}

func (s *PretranscodeService) runUniversalTranscodeWith(ctx context.Context, inputPath, outputPath string, profile *model.PretranscodeProfile, sourceHeight, bitrate int, hwAccel string, hdr bool, mf *model.MediaFile) error {
	args := []string{"-hide_banner", "-loglevel", "error", "-stats", "-y"}

	dvProfile := mf.DVProfile

	if !hdr {
		switch hwAccel {
		case "vaapi":
			args = append(args, "-vaapi_device", "/dev/dri/renderD128")
		case "nvenc":
			args = append(args, "-hwaccel", "cuda", "-hwaccel_output_format", "cuda")
		}
	} else {
		// HW Tonemap Path evaluates probe results
		if transcoder.ShouldUseHwTonemap(hwAccel, dvProfile, s.enableHwTonemap) {
			switch hwAccel {
			case "vaapi":
				args = append(args, "-vaapi_device", "/dev/dri/renderD128", "-hwaccel", "vaapi")
			case "nvenc":
				args = append(args, "-init_hw_device", "vulkan=vk:0", "-filter_hw_device", "vk")
			}
			if transcoder.NeedsDoviStripBSF(dvProfile) {
				args = append(args, "-bsf:v:0", "hevc_metadata=remove_dovi=1")
			}
		} else if hwAccel == "vaapi" {
			// SW HDR tonemap + HW Encode requires device init for vaapi
			args = append(args, "-vaapi_device", "/dev/dri/renderD128")
		}
	}

	args = append(args, "-i", inputPath)

	if hdr {
		if transcoder.ShouldUseHwTonemap(hwAccel, dvProfile, s.enableHwTonemap) {
			if hwAccel == "vaapi" {
				vf := buildPretranscodeHDRScaleFilter(inputPath, sourceHeight, "nv12", hwAccel, dvProfile, s.enableHwTonemap)
				args = append(args,
					"-vf", vf,
					"-c:v", "h264_vaapi",
					"-profile:v", "main",
					"-b:v", fmt.Sprintf("%dk", bitrate),
				)
			} else if hwAccel == "nvenc" {
				vf := buildPretranscodeHDRScaleFilter(inputPath, sourceHeight, "yuv420p", hwAccel, dvProfile, s.enableHwTonemap)
				args = append(args,
					"-vf", vf,
					"-c:v", "h264_nvenc",
					"-preset", "p4",
					"-profile:v", "high",
					"-b:v", fmt.Sprintf("%dk", bitrate),
					"-pix_fmt", "yuv420p",
				)
			}
		} else {
			vf := buildPretranscodeHDRScaleFilter(inputPath, sourceHeight, "yuv420p", hwAccel, dvProfile, s.enableHwTonemap)
			args = append(args,
				"-vf", vf,
				"-c:v", "libx264",
				"-preset", "medium",
				"-crf", "20",
				"-pix_fmt", "yuv420p",
			)
		}
		args = append(args, "-colorspace", "bt709", "-color_primaries", "bt709", "-color_trc", "bt709")
	} else {
		switch hwAccel {
		case "vaapi":
			args = append(args,
				"-vf", fmt.Sprintf("format=nv12,hwupload,scale_vaapi=w=-2:h=%d:format=nv12", sourceHeight),
				"-c:v", "h264_vaapi",
				"-profile:v", "main",
				"-b:v", fmt.Sprintf("%dk", bitrate),
			)
		case "nvenc":
			args = append(args,
				"-vf", fmt.Sprintf("scale_cuda=w=-2:h=%d:format=yuv420p", sourceHeight),
				"-c:v", "h264_nvenc",
				"-preset", "p4",
				"-profile:v", "high",
				"-pix_fmt", "yuv420p",
				"-b:v", fmt.Sprintf("%dk", bitrate),
			)
		case "amf":
			args = append(args,
				"-vf", fmt.Sprintf("scale=-2:%d", sourceHeight),
				"-c:v", "h264_amf",
				"-quality", "balanced",
				"-pix_fmt", "yuv420p",
				"-b:v", fmt.Sprintf("%dk", bitrate),
			)
		default:
			args = append(args, "-vf", fmt.Sprintf("scale=-2:%d", sourceHeight))
			args = append(args, "-c:v", "libx264", "-preset", "medium", "-crf", "20", "-pix_fmt", "yuv420p")
		}
	}

	args = append(args, "-c:a", "aac", "-b:a", fmt.Sprintf("%dk", profile.AudioBitrate), "-ac", "2",
		"-movflags", "+faststart", "-map", "0:v:0", "-map", "0:a:0", outputPath)

	cmd := niceFFmpeg(ctx, args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// buildPretranscodeHDRScaleFilter builds the SW tonemap filter chain using tonemapx.
// outFmt controls the final pixel format: "nv12" for VAAPI hwupload, "yuv420p" for SW/NVENC.
//
// Uses tonemapx (Jellyfin SIMD-optimized) which correctly handles both standard
// HDR10 (BT.2020/PQ) and Dolby Vision (including Profile 5 IPT-PQ-C2).
// outFmt controls the SW fallback pixel format.
func buildPretranscodeHDRScaleFilter(inputPath string, height int, outFmt string, hwAccel string, dvProfile int, enableHwTonemap bool) string {
	if transcoder.ShouldUseHwTonemap(hwAccel, dvProfile, enableHwTonemap) {
		hwFilter := transcoder.HDRToneMapFilterForHW(hwAccel, inputPath, dvProfile, enableHwTonemap)
		if hwAccel == "vaapi" {
			return hwFilter + fmt.Sprintf(",scale_vaapi=w=-2:h=%d:format=nv12", height)
		}
		// nvenc vulkan
		return hwFilter + fmt.Sprintf(",scale=-2:%d", height)
	}

	prefix := ""
	if ffprobe.NeedsHDRColorMetadataFallback(inputPath) {
		prefix = "setparams=color_primaries=bt2020:color_trc=smpte2084:colorspace=bt2020nc,"
	}
	return prefix + fmt.Sprintf(
		"tonemapx=tonemap=bt2390:desat=0:peak=100:t=bt709:m=bt709:p=bt709,format=%s,scale=-2:%d",
		outFmt, height,
	)
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

func (s *PretranscodeService) encodeFile(ctx context.Context, mf *model.MediaFile, outputPath string, profile *model.PretranscodeProfile, sourceAudioCodec string, hdr bool) error {
	if hdr {
		// HDR/DV: software tonemap is required, but try HW encode (hybrid/HW) first
		if s.hwAccel != "" {
			log.Printf("pretranscode: HDR/DV source detected, trying tonemap for %s", filepath.Base(mf.FilePath))
			err := s.runFFmpeg(ctx, mf, outputPath, profile, s.hwAccel, sourceAudioCodec, true)
			if err == nil {
				return nil
			}
			log.Printf("pretranscode: HDR HW/hybrid encode failed (%s), retrying full software: %v", s.hwAccel, err)
			_ = os.Remove(outputPath)
		} else {
			log.Printf("pretranscode: HDR/DV source detected, full SW tonemap for %s", filepath.Base(mf.FilePath))
		}
		return s.runFFmpeg(ctx, mf, outputPath, profile, "", sourceAudioCodec, true)
	}
	// Try HW accel first, fallback to software
	if s.hwAccel != "" {
		err := s.runFFmpeg(ctx, mf, outputPath, profile, s.hwAccel, sourceAudioCodec, false)
		if err == nil {
			return nil
		}
		log.Printf("pretranscode: HW encode failed (%s), retrying software: %v", s.hwAccel, err)
		_ = os.Remove(outputPath)
	}
	return s.runFFmpeg(ctx, mf, outputPath, profile, "", sourceAudioCodec, false)
}

func (s *PretranscodeService) runFFmpeg(ctx context.Context, mf *model.MediaFile, outputPath string, profile *model.PretranscodeProfile, hwAccel, sourceAudioCodec string, hdr bool) error {
	args := []string{"-hide_banner", "-loglevel", "error", "-stats", "-y"}

	dvProfile := mf.DVProfile

	// HW device init — needed for both normal HW and HDR hybrid (SW tonemap → HW encode)
	if !hdr {
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
	} else {
		// HW Tonemap Path evaluates probe results
		if transcoder.ShouldUseHwTonemap(hwAccel, dvProfile, s.enableHwTonemap) {
			switch hwAccel {
			case "vaapi":
				args = append(args, "-vaapi_device", "/dev/dri/renderD128", "-hwaccel", "vaapi")
			case "nvenc":
				args = append(args, "-init_hw_device", "vulkan=vk:0", "-filter_hw_device", "vk")
			}
			if transcoder.NeedsDoviStripBSF(dvProfile) {
				args = append(args, "-bsf:v:0", "hevc_metadata=remove_dovi=1")
			}
		} else if hwAccel == "vaapi" {
			// SW HDR tonemap + HW Encode requires device init for vaapi
			args = append(args, "-vaapi_device", "/dev/dri/renderD128")
		}
	}
	args = append(args, "-i", mf.FilePath)

	if hdr {
		// HDR/DV: software tonemap always runs on CPU.
		// When hwAccel is set, upload tonemapped frames to GPU for HW encode (hybrid).
		switch hwAccel {
		case "vaapi":
			baseFilter := buildPretranscodeHDRScaleFilter(mf.FilePath, profile.Height, "nv12", hwAccel, dvProfile, s.enableHwTonemap)
			if transcoder.ShouldUseHwTonemap(hwAccel, dvProfile, s.enableHwTonemap) {
				args = append(args, "-vf", baseFilter,
					"-c:v", "h264_vaapi", "-profile:v", "main",
					"-b:v", fmt.Sprintf("%dk", profile.VideoBitrate))
			} else {
				// SW tonemap → nv12 → hwupload to VAAPI surface
				args = append(args, "-vf", baseFilter+",hwupload",
					"-c:v", "h264_vaapi", "-profile:v", "main",
					"-b:v", fmt.Sprintf("%dk", profile.VideoBitrate))
			}
		case "nvenc":
			// NVENC accepts system memory frames directly
			args = append(args, "-vf", buildPretranscodeHDRScaleFilter(mf.FilePath, profile.Height, "yuv420p", hwAccel, dvProfile, s.enableHwTonemap),
				"-c:v", "h264_nvenc", "-preset", "p4", "-tune", "hq", "-profile:v", "high", "-pix_fmt", "yuv420p",
				"-b:v", fmt.Sprintf("%dk", profile.VideoBitrate))
		case "qsv":
			args = append(args, "-vf", buildPretranscodeHDRScaleFilter(mf.FilePath, profile.Height, "nv12", hwAccel, dvProfile, s.enableHwTonemap)+",hwupload=extra_hw_frames=64",
				"-c:v", "h264_qsv", "-preset", "medium",
				"-b:v", fmt.Sprintf("%dk", profile.VideoBitrate))
		default:
			// Full software
			args = append(args, "-vf", buildPretranscodeHDRScaleFilter(mf.FilePath, profile.Height, "yuv420p", hwAccel, dvProfile, s.enableHwTonemap),
				"-c:v", "libx264", "-preset", "medium", "-crf", "22", "-pix_fmt", "yuv420p")
		}
		args = append(args, "-colorspace", "bt709", "-color_primaries", "bt709", "-color_trc", "bt709")
	} else {
		// Video filter (scale)
		switch hwAccel {
		case "vaapi":
			args = append(args, "-vf", fmt.Sprintf("format=nv12,hwupload,scale_vaapi=w=-2:h=%d:format=nv12", profile.Height))
		case "nvenc":
			args = append(args, "-vf", fmt.Sprintf("scale_cuda=w=-2:h=%d:format=yuv420p", profile.Height))
		case "qsv":
			args = append(args, "-vf", fmt.Sprintf("scale_qsv=w=-2:h=%d:format=nv12", profile.Height))
		default:
			args = append(args, "-vf", fmt.Sprintf("scale=-2:%d", profile.Height))
		}

		// Video codec
		switch hwAccel {
		case "vaapi":
			args = append(args, "-c:v", "h264_vaapi", "-profile:v", "main")
		case "nvenc":
			args = append(args, "-c:v", "h264_nvenc", "-preset", "p4", "-tune", "hq", "-profile:v", "high", "-pix_fmt", "yuv420p")
		case "qsv":
			args = append(args, "-c:v", "h264_qsv", "-preset", "medium", "-pix_fmt", "nv12")
		case "videotoolbox":
			args = append(args, "-c:v", "h264_videotoolbox", "-pix_fmt", "yuv420p")
		case "amf":
			args = append(args, "-c:v", "h264_amf", "-quality", "balanced", "-pix_fmt", "yuv420p")
		default:
			args = append(args, "-c:v", "libx264", "-preset", "medium", "-crf", "22", "-pix_fmt", "yuv420p")
		}
	}

	// Bitrate (for HW encoders that don't support CRF)
	if hwAccel != "" && !hdr {
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
		// FFmpeg uses \r for progress and \n for errors — normalize both
		normalized := strings.ReplaceAll(errOutput, "\r\n", "\n")
		normalized = strings.ReplaceAll(normalized, "\r", "\n")
		lines := strings.Split(strings.TrimSpace(normalized), "\n")
		// Walk backwards to find the last non-progress line (actual error)
		errMsg := err.Error()
		for i := len(lines) - 1; i >= 0; i-- {
			line := strings.TrimSpace(lines[i])
			if line == "" || strings.HasPrefix(line, "frame=") || strings.HasPrefix(line, "size=") {
				continue
			}
			if len(line) > 300 {
				line = line[:300]
			}
			errMsg = line
			break
		}
		return fmt.Errorf("ffmpeg: %s", errMsg)
	}
	return nil
}

// niceFFmpeg wraps jellyfin-ffmpeg with nice -n 10 for lower CPU priority.
// Pretranscode runs in background — it should yield to realtime transcode but still get reasonable CPU time.
func niceFFmpeg(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, "nice", append([]string{"-n", "10", ffmpegbin.FFmpeg()}, args...)...)
}

// diskFreeSpace returns free bytes on the filesystem containing path.
func diskFreeSpace(path string) int64 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0
	}
	return int64(stat.Bavail) * int64(stat.Bsize)
}
