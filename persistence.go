package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// PersistenceStore handles saving and loading unsubmitted sessions to disk
type PersistenceStore struct {
	filePath string
}

// dataDir returns the data directory path, respecting XDG_DATA_HOME
func dataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "rescuetime-linux")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "share", "rescuetime-linux")
}

// DefaultPersistencePath returns the default sessions file path
func DefaultPersistencePath() string {
	dir := dataDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "sessions.json")
}

// NewPersistenceStore creates a new persistence store, ensuring the parent directory exists
func NewPersistenceStore(filePath string) (*PersistenceStore, error) {
	if filePath == "" {
		filePath = DefaultPersistencePath()
	}
	if filePath == "" {
		return nil, fmt.Errorf("unable to determine persistence path")
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create persistence directory %s: %v", dir, err)
	}

	return &PersistenceStore{filePath: filePath}, nil
}

// SaveSessions writes sessions to disk using atomic write (write .tmp, then rename)
func (ps *PersistenceStore) SaveSessions(sessions []ActivitySession) error {
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sessions: %v", err)
	}

	tmpPath := ps.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %v", err)
	}

	if err := os.Rename(tmpPath, ps.filePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %v", err)
	}

	slog.Debug("saved sessions to disk", "count", len(sessions), "path", ps.filePath)
	return nil
}

// LoadSessions reads sessions from disk. Returns empty slice if file is missing.
func (ps *PersistenceStore) LoadSessions() ([]ActivitySession, error) {
	data, err := os.ReadFile(ps.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []ActivitySession{}, nil
		}
		return nil, fmt.Errorf("failed to read sessions file: %v", err)
	}

	var sessions []ActivitySession
	if err := json.Unmarshal(data, &sessions); err != nil {
		slog.Warn("corrupted sessions file, starting fresh",
			"path", ps.filePath, "error", err)
		return []ActivitySession{}, nil
	}

	slog.Info("loaded saved sessions", "count", len(sessions), "path", ps.filePath)
	return sessions, nil
}

// Clear removes the persistence file after successful submission
func (ps *PersistenceStore) Clear() error {
	err := os.Remove(ps.filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove sessions file: %v", err)
	}
	slog.Debug("cleared persistence file", "path", ps.filePath)
	return nil
}
