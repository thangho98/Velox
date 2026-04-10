// Package ffmpegbin resolves the paths to the ffmpeg and ffprobe binaries.
//
// Velox uses jellyfin-ffmpeg as its primary media processing toolchain.
// Resolution order:
//  1. Environment variables JELLYFIN_FFMPEG_PATH / JELLYFIN_FFPROBE_PATH
//  2. jellyfin-ffmpeg default install location (/usr/lib/jellyfin-ffmpeg/*)
//  3. Standard PATH fallback (plain "ffmpeg" / "ffprobe")
package ffmpegbin

import "os"

const (
	jellyfinDir     = "/usr/lib/jellyfin-ffmpeg"
	jellyfinFFmpeg  = jellyfinDir + "/ffmpeg"
	jellyfinFFprobe = jellyfinDir + "/ffprobe"
)

// FFmpeg returns the path to the ffmpeg binary.
func FFmpeg() string {
	if path := os.Getenv("JELLYFIN_FFMPEG_PATH"); path != "" {
		return path
	}
	if _, err := os.Stat(jellyfinFFmpeg); err == nil {
		return jellyfinFFmpeg
	}
	return "ffmpeg"
}

// FFprobe returns the path to the ffprobe binary.
func FFprobe() string {
	if path := os.Getenv("JELLYFIN_FFPROBE_PATH"); path != "" {
		return path
	}
	if _, err := os.Stat(jellyfinFFprobe); err == nil {
		return jellyfinFFprobe
	}
	return "ffprobe"
}
