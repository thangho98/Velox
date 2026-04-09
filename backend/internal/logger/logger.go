package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Setup initializes the global slog default logger.
// Call once from main() before any logging happens.
//
// VELOX_LOG_LEVEL controls the minimum level:
//
//	debug, info (default), warn, error
//
// VELOX_LOG_FORMAT controls output format:
//
//	text (default), json
func Setup() {
	level := parseLevel(os.Getenv("VELOX_LOG_LEVEL"))
	format := strings.ToLower(os.Getenv("VELOX_LOG_FORMAT"))

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

// New creates a named child logger. Use this in constructors:
//
//	log := logger.New("stream")   // [stream] prefix in output
func New(name string) *slog.Logger {
	return slog.Default().With("component", name)
}

// NewWith creates a child logger with additional default fields:
//
//	log := logger.NewWith("hls", "media_id", id, "session", ssid)
func NewWith(name string, args ...any) *slog.Logger {
	return slog.Default().With(append([]any{"component", name}, args...)...)
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Debugf is a convenience for quick debug logging without creating a named logger.
// Only emits when VELOX_LOG_LEVEL=debug.
func Debugf(format string, args ...any) {
	slog.Debug(fmt.Sprintf(format, args...))
}

// Infof is a convenience for quick info logging without creating a named logger.
func Infof(format string, args ...any) {
	slog.Info(fmt.Sprintf(format, args...))
}
