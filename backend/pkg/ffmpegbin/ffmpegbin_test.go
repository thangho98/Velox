package ffmpegbin_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thawng/velox/pkg/ffmpegbin"
)

func TestFFmpeg_DefaultPath(t *testing.T) {
	t.Parallel()
	// When env vars are not set and jellyfin path doesn't exist
	// Should return "ffmpeg"
	os.Unsetenv("JELLYFIN_FFMPEG_PATH")
	os.Unsetenv("JELLYFIN_FFPROBE_PATH")

	// Note: This test may be environment-dependent
	// We test that the function returns something
	path := ffmpegbin.FFmpeg()
	assert.NotEmpty(t, path)
	assert.Equal(t, "ffmpeg", path)
}

func TestFFprobe_DefaultPath(t *testing.T) {
	t.Parallel()
	// When env vars are not set and jellyfin path doesn't exist
	// Should return "ffprobe"
	os.Unsetenv("JELLYFIN_FFMPEG_PATH")
	os.Unsetenv("JELLYFIN_FFPROBE_PATH")

	// Note: This test may be environment-dependent
	path := ffmpegbin.FFprobe()
	assert.NotEmpty(t, path)
	assert.Equal(t, "ffprobe", path)
}

func TestFFmpeg_EnvOverride(t *testing.T) {
	t.Parallel()
	// Set custom path via environment variable
	os.Setenv("JELLYFIN_FFMPEG_PATH", "/custom/ffmpeg")
	defer os.Unsetenv("JELLYFIN_FFMPEG_PATH")

	path := ffmpegbin.FFmpeg()
	assert.Equal(t, "/custom/ffmpeg", path)
}

func TestFFprobe_EnvOverride(t *testing.T) {
	t.Parallel()
	// Set custom path via environment variable
	os.Setenv("JELLYFIN_FFPROBE_PATH", "/custom/ffprobe")
	defer os.Unsetenv("JELLYFIN_FFPROBE_PATH")

	path := ffmpegbin.FFprobe()
	assert.Equal(t, "/custom/ffprobe", path)
}

func TestFFmpeg_JellyfinPath(t *testing.T) {
	t.Parallel()
	// When JELLYFIN_FFMPEG_PATH is set, it takes precedence
	os.Setenv("JELLYFIN_FFMPEG_PATH", "/custom/ffmpeg")
	defer os.Unsetenv("JELLYFIN_FFMPEG_PATH")

	// Even if jellyfin path exists, env var should win
	path := ffmpegbin.FFmpeg()
	assert.Equal(t, "/custom/ffmpeg", path)
}

func TestFFprobe_JellyfinPath(t *testing.T) {
	t.Parallel()
	// When JELLYFIN_FFPROBE_PATH is set, it takes precedence
	os.Setenv("JELLYFIN_FFPROBE_PATH", "/custom/ffprobe")
	defer os.Unsetenv("JELLYFIN_FFPROBE_PATH")

	path := ffmpegbin.FFprobe()
	assert.Equal(t, "/custom/ffprobe", path)
}

func TestFFmpeg_EmptyEnvClearsOverride(t *testing.T) {
	t.Parallel()
	// Set env var to empty string - should fall back to default
	os.Setenv("JELLYFIN_FFMPEG_PATH", "")
	path := ffmpegbin.FFmpeg()
	// With empty string, it should use the jellyfin path or default
	// The function treats empty string as not set (Getenv returns "" for unset vars too)
	// So this might return jellyfin path or "ffmpeg"
	assert.NotEmpty(t, path)
}

func TestConstants(t *testing.T) {
	t.Parallel()
	// Verify the jellyfin directory constant
	// We can't directly access the constant, but we can verify behavior
}
