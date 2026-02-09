# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

RescueTime Linux Activity Tracker - a native Go application that monitors active window usage on Hyprland/Wayland and submits time tracking data to RescueTime. Currently Hyprland-specific via `hyprctl` command.

## Build and Run Commands

```bash
# Build the binary
go build -o active-window .

# Single window query (current active window)
./active-window

# Monitor window changes (display only)
./active-window -monitor

# Track time with summary on exit (Ctrl+C)
./active-window -track

# Track and submit to RescueTime API
./active-window -track -submit

# Custom intervals
./active-window -track -submit -interval 500ms -submission-interval 5m

# Debug logging
./active-window -track -log-level debug

# Log to file
./active-window -track -log-file /tmp/rescuetime.log

# Live stats every 10 seconds
./active-window -track -stats -stats-interval 10s

# Full feature test
./active-window -track -submit -stats -stats-interval 10s -log-level debug -submission-interval 2m

# Custom config file
./active-window -track -config /path/to/config.json
```

## Environment Setup

Requires `.env` file with:
- `RESCUE_TIME_API_KEY` - for legacy offline time API
- `RESCUE_TIME_ACCOUNT_KEY` - for native client API (optional)
- `RESCUE_TIME_DATA_KEY` - for native client API (optional)

## Configuration

JSON config at `~/.config/rescuetime-linux/config.json` (respects `XDG_CONFIG_HOME`).

Precedence: defaults < config file < CLI flags < env vars (secrets only)

```json
{
  "polling_interval": "200ms",
  "submission_interval": "15m",
  "merge_threshold": "30s",
  "min_duration": "10s",
  "log_level": "info",
  "log_file": "",
  "env_file_path": ".env",
  "persistence_path": "",
  "legacy_api_endpoint": "https://www.rescuetime.com/anapi/offline_time_post",
  "native_api_endpoint": "https://api.rescuetime.com/api/resource/user_client_events"
}
```

## Architecture

Multi-file Go application (`package main`) with these source files:

| File | Purpose |
|------|---------|
| `active-window.go` | HyprlandWindow, window helpers, monitorWindowChanges, main |
| `tracker.go` | ActivitySession, ActivitySummary, ActivityTracker + all methods |
| `api.go` | API types, activation, submission (parallel with bounded concurrency) |
| `config.go` | Config struct, Duration JSON wrapper, XDG-aware config loading |
| `logging.go` | Structured logging setup with log/slog |
| `persistence.go` | PersistenceStore for saving/loading unsubmitted sessions |
| `titleclean.go` | Rule-based window title cleaning (notification counts, browser names) |

### Data Structures
- `HyprlandWindow` - JSON structure from `hyprctl activewindow -j`
- `ActivitySession` - single continuous session with start/end times
- `ActivitySummary` - aggregated time per application
- `ActivityTracker` - thread-safe session management with `sync.RWMutex`
- `Config` / `Duration` - JSON-serializable configuration with duration strings
- `PersistenceStore` - atomic file-based session persistence
- `TitleCleaner` - regex rule pipeline for window title normalization
- `SubmissionResult` - structured batch submission outcome

### Key Behaviors
- Polls `hyprctl` every 200ms (configurable)
- Merges sessions if same app within 30 seconds (configurable)
- Ignores sessions under 10 seconds (configurable)
- Cleans window titles (strips notification counts, browser suffixes)
- Submits to RescueTime every 15 minutes (configurable) with parallel bounded concurrency (3 goroutines)
- Persists unsubmitted sessions to `~/.local/share/rescuetime-linux/sessions.json` (respects `XDG_DATA_HOME`)
- Recovers persisted sessions on startup
- Handles SIGINT/SIGTERM for graceful shutdown with final data submission
- Optional live stats display (top-5 apps by duration)

### API Integration (dual-mode)
1. **Native API** (preferred when credentials available): `POST https://api.rescuetime.com/api/resource/user_client_events` with RFC 3339 timestamps
2. **Legacy API** (fallback): `POST https://www.rescuetime.com/anapi/offline_time_post` with `YYYY-MM-DD HH:MM:SS` format

Both use exponential backoff retry (3 attempts: 1s, 2s, 4s).

## Testing

```bash
# Short tracking session
./active-window -track
# Switch windows for ~30s, then Ctrl+C

# Test API with short interval
./active-window -track -submit -submission-interval 2m

# Crash recovery test
./active-window -track &
sleep 30 && kill -9 $!
./active-window -track -submit -submission-interval 30s
# Should log "loaded saved sessions" and submit recovered data
```

HTTP test requests in `rescuetime-auth.http`.

## Platform Requirements

- Linux with Wayland
- Hyprland compositor (uses `hyprctl`)
- Go 1.22+ (uses `log/slog`, integer range)
- Checks for `WAYLAND_DISPLAY` environment variable

## Reference Documentation

- `context/rescuetime/api-docs.md` - RescueTime API documentation
- `RescueTime-Complete-Authentication-Reverse-Engineering-Report.md` - native client API reverse engineering
- `context/todo/implementation-plan.md` - detailed implementation roadmap (Phases 4-6 complete, Phases 7-8 pending)
