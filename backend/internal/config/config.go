package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Host              string
	Port              string
	DataDir           string
	DatabasePath      string
	TranscodePath     string
	SubtitleCachePath string
	TrickplayPath     string
	CORSOrigin        string

	// Hardware transcoding (Plan E Phase 01)
	HWAccel       string // auto|videotoolbox|vaapi|nvenc|qsv|none
	MaxTranscodes int    // max concurrent FFmpeg jobs

	// Trickplay thumbnails (Plan E Phase 03)
	TrickplayEnabled  bool
	TrickplayInterval int // seconds between thumbnail frames

	// File watcher (Phase 03)
	FileWatcherEnabled bool
}

func Load() *Config {
	dataDir := envOrDefault("VELOX_DATA_DIR", defaultDataDir())

	return &Config{
		Host:              envOrDefault("VELOX_HOST", "0.0.0.0"),
		Port:              envOrDefault("VELOX_PORT", "8080"),
		DataDir:           dataDir,
		DatabasePath:      filepath.Join(dataDir, "velox.db"),
		TranscodePath:     filepath.Join(dataDir, "transcode"),
		SubtitleCachePath: filepath.Join(dataDir, "subtitles"),
		TrickplayPath:     filepath.Join(dataDir, "trickplay"),
		CORSOrigin:        envOrDefault("VELOX_CORS_ORIGIN", "http://localhost:5173"),

		HWAccel:           envOrDefault("VELOX_HW_ACCEL", "auto"),
		MaxTranscodes:     envOrDefaultInt("VELOX_MAX_TRANSCODES", 2),
		TrickplayEnabled:  envOrDefaultBool("VELOX_TRICKPLAY_ENABLED", false),
		TrickplayInterval: envOrDefaultInt("VELOX_TRICKPLAY_INTERVAL", 10),

		FileWatcherEnabled: envOrDefaultBool("VELOX_FILE_WATCHER", true),
	}
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func envOrDefaultBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "yes":
		return true
	case "0", "false", "no":
		return false
	}
	return fallback
}

func defaultDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".velox")
}
