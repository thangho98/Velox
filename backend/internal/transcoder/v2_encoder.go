package transcoder

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/thawng/velox/internal/model"
)

type V2EncoderOpts struct {
	InputPath         string
	OutputDir         string
	Prefix            string
	StartSegNum       int
	AudioTracks       []model.AudioTrack
	VideoCopy         bool
	SubtitleStreamIdx int
	MaxHeight         int
	HwAccel           string
	SegLength         float64
}

func getJellyfinFFmpegPath() string {
	if path := os.Getenv("JELLYFIN_FFMPEG_PATH"); path != "" {
		return path
	}
	// Default installation path for jellyfin-ffmpeg on Debian/Ubuntu
	if _, err := os.Stat("/usr/lib/jellyfin-ffmpeg/ffmpeg"); err == nil {
		return "/usr/lib/jellyfin-ffmpeg/ffmpeg"
	}
	// Fallback to standard ffmpeg
	return "ffmpeg"
}

// StartHLSV2Encoder launches a single ffmpeg process per session to transcode video and audio.
func StartHLSV2Encoder(ctx context.Context, opts V2EncoderOpts) (*exec.Cmd, error) {
	args := []string{"-hide_banner", "-loglevel", "warning"}

	startOffset := float64(opts.StartSegNum) * opts.SegLength

	if opts.VideoCopy {
		args = append(args, ffmpegInputProbeArgs()...)
	} else {
		args = append(args, buildFFmpegInputArgs(opts.HwAccel)...)
	}

	if startOffset > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", startOffset))
	}
	args = append(args, "-i", opts.InputPath)

	// CRITICAL: preserve source PTS so segment timestamps match the original
	// video timeline. When ffmpeg is restarted mid-video via -ss, shifting PTS
	// to zero (via -avoid_negative_ts make_zero) causes hls.js to place the
	// first emitted segment at player time 0 instead of at the seek position,
	// so video.currentTime set to the user's resume point points to an empty
	// range and playback freezes. This matches Jellyfin's DynamicHlsController.
	args = append(args,
		"-copyts",
		"-avoid_negative_ts", "disabled",
	)

	if opts.VideoCopy {
		args = append(args,
			"-map_metadata", "-1", "-map_chapters", "-1",
			"-max_muxing_queue_size", "4096",
		)
	}

	hdr := isHDRFile(opts.InputPath)

	if opts.VideoCopy {
		args = append(args, "-map", "0:v:0", "-c:v", "copy")
	} else if opts.SubtitleStreamIdx >= 0 {
		args = append(args, buildImageSubtitleBurnInVideoOnlyArgs(opts.HwAccel, hdr, opts.SubtitleStreamIdx)...)
	} else {
		args = append(args, "-map", "0:v:0")
		encArgs := buildVideoEncodeArgs(opts.HwAccel, hdr, opts.SubtitleStreamIdx, opts.InputPath)
		if opts.MaxHeight > 0 {
			scaleFilter := hwScaleFilter(opts.HwAccel, opts.MaxHeight)
			inserted := false
			for i, arg := range encArgs {
				if arg == "-vf" && i+1 < len(encArgs) {
					if opts.HwAccel == "vaapi" {
						encArgs[i+1] = strings.Replace(encArgs[i+1], "scale_vaapi=format=nv12", scaleFilter, 1)
					} else {
						encArgs[i+1] = scaleFilter + "," + encArgs[i+1]
					}
					inserted = true
					break
				}
			}
			if !inserted {
				encArgs = append([]string{"-vf", scaleFilter}, encArgs...)
			}
		}
		args = append(args, encArgs...)
	}

	var videoHLSArgs []string
	if opts.VideoCopy {
		videoHLSArgs = []string{
			"-an",
			"-muxdelay", "0", "-muxpreload", "0",
			"-f", "hls", "-max_delay", "5000000",
			"-hls_segment_type", "fmp4",
			"-hls_fmp4_init_filename", opts.Prefix + "video_-1.mp4",
			"-hls_time", fmt.Sprintf("%v", opts.SegLength),
			"-hls_list_size", "0", "-hls_playlist_type", "event",
			"-start_number", fmt.Sprintf("%d", opts.StartSegNum),
			"-hls_segment_filename", filepath.Join(opts.OutputDir, opts.Prefix+"video_%d.m4s"),
			filepath.Join(opts.OutputDir, opts.Prefix+"video.m3u8"), // ffmpeg output real playlist here
		}
	} else {
		videoHLSArgs = []string{
			"-an",
			"-f", "hls", "-max_delay", "5000000",
			"-hls_segment_type", "fmp4",
			"-hls_fmp4_init_filename", opts.Prefix + "video_-1.mp4",
			"-hls_time", fmt.Sprintf("%v", opts.SegLength),
			"-hls_list_size", "0", "-hls_playlist_type", "event",
			"-start_number", fmt.Sprintf("%d", opts.StartSegNum),
			"-hls_segment_filename", filepath.Join(opts.OutputDir, opts.Prefix+"video_%d.m4s"),
			filepath.Join(opts.OutputDir, opts.Prefix+"video.m3u8"),
		}
	}
	args = append(args, videoHLSArgs...)

	for i, track := range opts.AudioTracks {
		args = append(args,
			"-map", fmt.Sprintf("0:%d", track.StreamIndex),
			"-c:a", "aac", "-b:a", "192k", "-ac", "2",
			"-f", "hls", "-hls_segment_type", "fmp4",
			"-hls_fmp4_init_filename", fmt.Sprintf("%saudio%d_-1.mp4", opts.Prefix, i),
			"-hls_time", fmt.Sprintf("%v", opts.SegLength),
			"-hls_list_size", "0", "-hls_playlist_type", "event",
			"-start_number", fmt.Sprintf("%d", opts.StartSegNum),
			"-hls_segment_filename", filepath.Join(opts.OutputDir, fmt.Sprintf("%saudio%d_%%d.m4s", opts.Prefix, i)),
			filepath.Join(opts.OutputDir, fmt.Sprintf("%saudio%d.m3u8", opts.Prefix, i)),
		)
	}

	cmd := exec.CommandContext(ctx, getJellyfinFFmpegPath(), args...)

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("create ffmpeg stderr pipe: %w", err)
	}
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			slog.Warn("ffmpeg_v2", "prefix", opts.Prefix, "line", scanner.Text())
		}
	}()

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}
