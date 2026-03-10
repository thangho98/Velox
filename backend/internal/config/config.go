package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Host          string
	Port          string
	DatabasePath  string
	TranscodePath string
	CORSOrigin    string
}

func Load() *Config {
	dataDir := envOrDefault("VELOX_DATA_DIR", defaultDataDir())

	return &Config{
		Host:          envOrDefault("VELOX_HOST", "0.0.0.0"),
		Port:          envOrDefault("VELOX_PORT", "8080"),
		DatabasePath:  filepath.Join(dataDir, "velox.db"),
		TranscodePath: filepath.Join(dataDir, "transcode"),
		CORSOrigin:    envOrDefault("VELOX_CORS_ORIGIN", "http://localhost:5173"),
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

func defaultDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".velox")
}
