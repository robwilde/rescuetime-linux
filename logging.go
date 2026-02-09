package main

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// parseLogLevel converts a string level name to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// setupLogger creates a multi-writer slog handler (stdout + optional file).
// Returns a cleanup function that should be deferred by the caller.
func setupLogger(level slog.Level, logFilePath string) func() {
	var writers []io.Writer
	writers = append(writers, os.Stdout)

	var logFile *os.File
	cleanup := func() {}

	if logFilePath != "" {
		f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			// Fall back to stdout-only logging
			slog.Warn("failed to open log file, using stdout only", "path", logFilePath, "error", err)
		} else {
			logFile = f
			writers = append(writers, logFile)
			cleanup = func() {
				logFile.Close()
			}
		}
	}

	writer := io.MultiWriter(writers...)
	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))

	return cleanup
}
