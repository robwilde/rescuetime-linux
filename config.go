package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// Duration wraps time.Duration for JSON marshal/unmarshal as human-readable strings
type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %v", s, err)
	}
	d.Duration = parsed
	return nil
}

// Config holds all configurable values for the application
type Config struct {
	PollingInterval    Duration `json:"polling_interval"`
	SubmissionInterval Duration `json:"submission_interval"`
	MergeThreshold     Duration `json:"merge_threshold"`
	MinDuration        Duration `json:"min_duration"`
	LogLevel           string   `json:"log_level"`
	LogFile            string   `json:"log_file"`
	EnvFilePath        string   `json:"env_file_path"`
	PersistencePath    string   `json:"persistence_path"`
	LegacyAPIEndpoint  string   `json:"legacy_api_endpoint"`
	NativeAPIEndpoint  string   `json:"native_api_endpoint"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		PollingInterval:    Duration{200 * time.Millisecond},
		SubmissionInterval: Duration{15 * time.Minute},
		MergeThreshold:     Duration{30 * time.Second},
		MinDuration:        Duration{10 * time.Second},
		LogLevel:           "info",
		LogFile:            "",
		EnvFilePath:        ".env",
		PersistencePath:    "",
		LegacyAPIEndpoint:  "https://www.rescuetime.com/anapi/offline_time_post",
		NativeAPIEndpoint:  "https://api.rescuetime.com/api/resource/user_client_events",
	}
}

// configDir returns the config directory path, respecting XDG_CONFIG_HOME
func configDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "rescuetime-linux")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "rescuetime-linux")
}

// DefaultConfigPath returns the default config file path
func DefaultConfigPath() string {
	dir := configDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "config.json")
}

// LoadConfig loads configuration from a JSON file.
// Returns the default config if the file doesn't exist.
func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		path = DefaultConfigPath()
	}
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Debug("no config file found, using defaults", "path", path)
			return cfg, nil
		}
		return cfg, fmt.Errorf("failed to read config file: %v", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse config file: %v", err)
	}

	slog.Info("loaded config file", "path", path)
	return cfg, nil
}

// ApplyCLIFlags overrides config values with CLI flag values when flags were explicitly set.
// flagSet maps flag names to their values: "interval", "submission-interval", "log-level", "log-file".
func (c *Config) ApplyCLIFlags(flags map[string]string) {
	if v, ok := flags["interval"]; ok {
		if d, err := time.ParseDuration(v); err == nil {
			c.PollingInterval = Duration{d}
		}
	}
	if v, ok := flags["submission-interval"]; ok {
		if d, err := time.ParseDuration(v); err == nil {
			c.SubmissionInterval = Duration{d}
		}
	}
	if v, ok := flags["log-level"]; ok {
		c.LogLevel = v
	}
	if v, ok := flags["log-file"]; ok {
		c.LogFile = v
	}
}
