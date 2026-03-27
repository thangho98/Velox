package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Host              string
	Port              string
	DataDir           string
	DatabasePath      string
	TranscodePath     string
	SubtitleCachePath string
	TrickplayPath     string
	PretranscodePath  string
	CORSOrigin        string

	// Hardware transcoding (Plan E Phase 01)
	HWAccel       string // auto|videotoolbox|vaapi|nvenc|qsv|none
	MaxTranscodes int    // max concurrent FFmpeg jobs

	// Trickplay thumbnails (Plan E Phase 03)
	TrickplayEnabled  bool
	TrickplayInterval int // seconds between thumbnail frames

	// File watcher (Phase 03)
	FileWatcherEnabled bool

	// Built-in API keys from env (optional - for open-source distribution)
	// If set, these act as default keys when user hasn't configured custom keys
	TMDbAPIKey   string // VELOX_TMDB_API_KEY
	OMDbAPIKey   string // VELOX_OMDB_API_KEY
	TVDBAPIKey   string // VELOX_TVDB_API_KEY
	FanartAPIKey string // VELOX_FANART_API_KEY
	SubdlAPIKey  string // VELOX_SUBDL_API_KEY
}

// LoadDotEnv loads the first .env file found from a small set of common paths.
// Existing process environment variables always win over file values.
func LoadDotEnv() error {
	candidates := []string{
		".env",
		filepath.Join("backend", ".env"),
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		return loadDotEnvFile(candidate)
	}

	return nil
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
		PretranscodePath:  envOrDefault("VELOX_PRETRANSCODE_DIR", filepath.Join(dataDir, "pretranscode")),
		CORSOrigin:        envOrDefault("VELOX_CORS_ORIGIN", "http://localhost:5173"),

		HWAccel:           envOrDefault("VELOX_HW_ACCEL", "auto"),
		MaxTranscodes:     envOrDefaultInt("VELOX_MAX_TRANSCODES", 2),
		TrickplayEnabled:  envOrDefaultBool("VELOX_TRICKPLAY_ENABLED", false),
		TrickplayInterval: envOrDefaultInt("VELOX_TRICKPLAY_INTERVAL", 10),

		FileWatcherEnabled: envOrDefaultBool("VELOX_FILE_WATCHER", true),

		// Built-in API keys from env (optional)
		TMDbAPIKey:   envOrDefault("VELOX_TMDB_API_KEY", ""),
		OMDbAPIKey:   envOrDefault("VELOX_OMDB_API_KEY", ""),
		TVDBAPIKey:   envOrDefault("VELOX_TVDB_API_KEY", ""),
		FanartAPIKey: envOrDefault("VELOX_FANART_API_KEY", ""),
		SubdlAPIKey:  envOrDefault("VELOX_SUBDL_API_KEY", ""),
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

func loadDotEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}

		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				if unquoted, err := strconv.Unquote(value); err == nil {
					value = unquoted
				} else {
					value = value[1 : len(value)-1]
				}
			}
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}
