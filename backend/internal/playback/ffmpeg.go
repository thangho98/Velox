package playback

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// FFmpegArgs holds FFmpeg command arguments
type FFmpegArgs struct {
	Global   []string
	Input    []string
	Video    []string
	Audio    []string
	Subtitle []string
	Output   []string
}

// BuildFFmpegArgs builds FFmpeg arguments based on playback decision
func BuildFFmpegArgs(decision PlaybackDecision, inputPath string, outputPath string, segmentPattern string) FFmpegArgs {
	args := FFmpegArgs{
		Global: []string{
			"-hide_banner",
			"-loglevel", "warning",
			"-stats",
		},
		Input: []string{
			"-i", inputPath,
		},
	}

	// Build video arguments
	args.Video = buildVideoArgs(decision)

	// Subtitle burn-in: prepend -vf filter to video args so the subtitle is
	// rendered into the video stream during transcode.
	// Uses the absolute FFprobe stream index stored in SubtitleStreamIndex.
	if decision.SubtitleAction == SubtitleBurnIn {
		filter := fmt.Sprintf("subtitles='%s':si=%d", inputPath, decision.SubtitleStreamIndex)
		args.Video = append([]string{"-vf", filter}, args.Video...)
	}

	// Build audio arguments
	args.Audio = buildAudioArgs(decision)

	// Build subtitle arguments
	args.Subtitle = buildSubtitleArgs(decision)

	// Build output arguments
	if segmentPattern != "" {
		// HLS output
		args.Output = buildHLSArgs(decision, outputPath, segmentPattern)
	} else {
		// Single file output
		args.Output = []string{outputPath}
	}

	return args
}

// buildVideoArgs builds video encoding arguments
func buildVideoArgs(decision PlaybackDecision) []string {
	if decision.VideoAction == VideoCopy {
		return []string{"-c:v", "copy"}
	}

	// Transcoding required
	codec := decision.VideoCodec
	if codec == "" {
		codec = CodecH264 // Default
	}

	args := []string{"-c:v"}

	switch codec {
	case CodecH264:
		args = append(args, "libx264",
			"-preset", "fast",
			"-crf", "22",
			"-profile:v", "high",
			"-level", "4.1",
		)
	case CodecH265:
		args = append(args, "libx265",
			"-preset", "fast",
			"-crf", "26",
			"-tag:v", "hvc1", // For Safari compatibility
		)
	case CodecVP9:
		args = append(args, "libvpx-vp9",
			"-deadline", "good",
			"-cpu-used", "4",
			"-crf", "31",
			"-b:v", "0", // Constrained quality mode
		)
	case CodecAV1:
		args = append(args, "libsvtav1",
			"-preset", "6",
			"-crf", "30",
		)
	default:
		args = append(args, "libx264",
			"-preset", "fast",
			"-crf", "22",
		)
	}

	// Add resolution scaling if needed
	if decision.EstimatedBitrate > 0 {
		// Replace CRF with explicit bitrate control
		bitrateStr := fmt.Sprintf("%dk", decision.EstimatedBitrate)
		replaced := false
		for i := 0; i < len(args)-1; i++ {
			if args[i] == "-crf" {
				// Remove the "-crf <value>" pair and insert "-b:v <bitrate>" in its place
				newArgs := make([]string, 0, len(args)+2)
				newArgs = append(newArgs, args[:i]...)
				newArgs = append(newArgs, "-b:v", bitrateStr)
				newArgs = append(newArgs, args[i+2:]...)
				// Add maxrate for VBR
				newArgs = append(newArgs, "-maxrate", fmt.Sprintf("%dk", int(float64(decision.EstimatedBitrate)*1.5)))
				newArgs = append(newArgs, "-bufsize", fmt.Sprintf("%dk", decision.EstimatedBitrate*2))
				args = newArgs
				replaced = true
				break
			}
		}
		if !replaced {
			// No CRF present (e.g. VP9 constrained quality mode already uses -b:v 0)
			args = append(args, "-maxrate", fmt.Sprintf("%dk", decision.EstimatedBitrate))
			args = append(args, "-bufsize", fmt.Sprintf("%dk", decision.EstimatedBitrate*2))
		}
	}

	// Add pixel format for compatibility
	args = append(args, "-pix_fmt", "yuv420p")

	return args
}

// buildAudioArgs builds audio encoding arguments
func buildAudioArgs(decision PlaybackDecision) []string {
	if decision.AudioAction == AudioCopy {
		return []string{"-c:a", "copy"}
	}

	codec := decision.AudioCodec
	if codec == "" {
		codec = CodecAAC
	}

	args := []string{"-c:a"}

	switch codec {
	case CodecAAC:
		args = append(args, "aac",
			"-b:a", "192k",
			"-ac", "2", // Stereo
		)
	case CodecOpus:
		args = append(args, "libopus",
			"-b:a", "128k",
			"-ac", "2",
		)
	case CodecMP3:
		args = append(args, "libmp3lame",
			"-b:a", "192k",
			"-ac", "2",
		)
	default:
		args = append(args, "aac",
			"-b:a", "192k",
			"-ac", "2",
		)
	}

	return args
}

// buildSubtitleArgs builds subtitle encoding arguments
func buildSubtitleArgs(decision PlaybackDecision) []string {
	if decision.SubtitleAction == SubtitleNone {
		return []string{"-sn"} // Disable subtitles
	}

	if decision.SubtitleAction == SubtitleBurnIn {
		// Filter already prepended to video args in BuildFFmpegArgs.
		// Suppress the subtitle output stream so it is not duplicated.
		return []string{"-sn"}
	}

	// Text-based subtitles - copy or convert
	return []string{
		"-c:s", "webvtt",
	}
}

// buildHLSArgs builds HLS output arguments
func buildHLSArgs(decision PlaybackDecision, playlistPath string, segmentPattern string) []string {
	segmentDir := filepath.Dir(segmentPattern)
	segmentName := filepath.Base(segmentPattern)

	args := []string{
		"-f", "hls",
		"-hls_time", "6",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", filepath.Join(segmentDir, segmentName),
		"-hls_flags", "independent_segments",
	}

	// Add codec-specific optimizations
	if decision.VideoAction == VideoTranscode {
		switch decision.VideoCodec {
		case CodecH264:
			// Ensure keyframes align with segment boundaries
			args = append(args,
				"-force_key_frames", "expr:gte(t,n_forced*6)",
				"-sc_threshold", "0",
			)
		}
	}

	args = append(args, playlistPath)
	return args
}

// ToCommand converts FFmpegArgs to command-line string slice
func (a FFmpegArgs) ToCommand() []string {
	result := []string{}
	result = append(result, a.Global...)
	result = append(result, a.Input...)
	result = append(result, a.Video...)
	result = append(result, a.Audio...)
	result = append(result, a.Subtitle...)
	result = append(result, a.Output...)
	return result
}

// BuildRemuxArgs builds args for container remuxing (no transcoding)
func BuildRemuxArgs(inputPath string, outputPath string) []string {
	return []string{
		"-hide_banner",
		"-loglevel", "warning",
		"-i", inputPath,
		"-c", "copy", // Copy all streams
		"-movflags", "+faststart", // Enable faststart for MP4
		"-y", // Overwrite output
		outputPath,
	}
}

// BuildExtractSubtitleArgs builds args for extracting subtitle
func BuildExtractSubtitleArgs(inputPath string, streamIndex int, outputPath string) []string {
	return []string{
		"-hide_banner",
		"-loglevel", "error",
		"-i", inputPath,
		"-map", fmt.Sprintf("0:s:%d", streamIndex),
		"-c:s", "webvtt",
		"-y",
		outputPath,
	}
}

// BuildBurnSubtitleArgs builds args for burning subtitle into video
func BuildBurnSubtitleArgs(inputPath string, subtitlePath string, outputPath string, decision PlaybackDecision) []string {
	// First build video args
	vArgs := buildVideoArgs(decision)
	aArgs := buildAudioArgs(decision)

	// Build filter for subtitle burn-in
	subFilter := fmt.Sprintf("subtitles='%s'", subtitlePath)

	// Add filter to video args
	for i, arg := range vArgs {
		if arg == "-c:v" {
			// Insert filter before codec
			vArgs = append(vArgs[:i], append([]string{"-vf", subFilter}, vArgs[i:]...)...)
			break
		}
	}

	args := []string{
		"-hide_banner",
		"-loglevel", "warning",
		"-i", inputPath,
	}
	args = append(args, vArgs...)
	args = append(args, aArgs...)
	args = append(args, "-y", outputPath)

	return args
}

// EstimateTranscodeTime estimates transcoding time based on duration and method
func EstimateTranscodeTime(durationSec int, method PlaybackMethod) int {
	// Rough estimates (in seconds)
	switch method {
	case MethodDirectPlay:
		return 0
	case MethodDirectStream:
		return durationSec / 10 // Remuxing is fast
	case MethodTranscodeAudio:
		return durationSec / 5 // Audio transcoding
	case MethodFullTranscode:
		return durationSec * 2 // Full transcode is slow
	default:
		return durationSec
	}
}

// FormatDuration formats seconds to HH:MM:SS
func FormatDuration(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// ParseBitrate parses bitrate string (e.g., "5M", "500k") to kbps
func ParseBitrate(s string) (int, error) {
	if s == "" {
		return 0, nil
	}

	// Extract number and unit
	numStr := ""
	unit := ""
	for i, c := range s {
		if c >= '0' && c <= '9' || c == '.' {
			numStr += string(c)
		} else {
			unit = s[i:]
			break
		}
	}

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}

	switch strings.ToLower(unit) {
	case "k", "kbps":
		return int(num), nil
	case "m", "mbps":
		return int(num * 1000), nil
	case "g", "gbps":
		return int(num * 1000000), nil
	default:
		return int(num), nil
	}
}
