# RescueTime Linux Activity Tracker

A native Linux activity tracker for [RescueTime](https://www.rescuetime.com) that monitors active window usage on Hyprland/Wayland compositors and submits time tracking data via the RescueTime API.

> **Status:** Phases 1-6 complete. Dual-mode API active. Tests and packaging pending (Phases 7-8).

## Features

- **Hyprland/Wayland Support** — Monitors active window focus changes using `hyprctl`
- **Smart Session Tracking** — Automatically tracks time spent in each application
- **Intelligent Merging** — Merges brief window switches to the same app (< 30 seconds)
- **Session Filtering** — Ignores very short sessions (< 10 seconds) to reduce noise
- **Dual-Mode API Submission** — Native client API first, automatic legacy fallback
- **Parallel Bounded Concurrency** — Submits activities via 3 concurrent goroutines
- **Session Persistence** — Saves unsubmitted sessions to disk for crash recovery
- **JSON Configuration** — XDG-compliant config file with CLI flag overrides
- **Structured Logging** — `log/slog`-based logging with configurable levels and file output
- **Window Title Cleaning** — Regex pipeline strips notification counts and browser suffixes
- **Live Statistics** — Optional periodic display of top-5 apps by duration
- **Graceful Shutdown** — Submits final data on exit (SIGINT/SIGTERM)
- **Retry Logic** — Exponential backoff for failed API submissions (3 attempts: 1s, 2s, 4s)
- **XDG-Compliant Paths** — Respects `XDG_CONFIG_HOME` and `XDG_DATA_HOME`
- **Automatic Submission** — Sends activity data to RescueTime every 15 minutes (configurable)

## Requirements

- **OS:** Linux with Wayland or X11
- **Compositor:** Hyprland (with `hyprctl` command)
- **Runtime:** Go 1.22+ (uses `log/slog`, integer range)
- **RescueTime Account:** Free or paid account with API access

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/robwilde/rescuetime-linux.git
cd rescuetime-linux

# Build the binary
go build -o active-window .

# Create environment file
cp .env.example .env
# Edit .env and add your RescueTime API key
```

### Environment Setup

Create a `.env` file in the project directory (or copy from `.env.example`):

```bash
# Required: Legacy offline time API key
# Get from: https://www.rescuetime.com/anapi/manage
RESCUE_TIME_API_KEY=your_api_key_here

# Optional: Native client API credentials
# Obtained via the /activate endpoint (see reverse engineering report)
RESCUE_TIME_ACCOUNT_KEY=your_account_key_here
RESCUE_TIME_DATA_KEY=your_data_key_here
```

**Getting your API key:**
1. Log in to [RescueTime](https://www.rescuetime.com)
2. Navigate to Settings → API & Integrations
3. Generate or copy your API key

When native credentials (`RESCUE_TIME_ACCOUNT_KEY` / `RESCUE_TIME_DATA_KEY`) are present, the application uses the native client API with automatic fallback to the legacy API on failure.

## Configuration

The application loads configuration from a JSON file at `~/.config/rescuetime-linux/config.json` (respects `XDG_CONFIG_HOME`).

**Precedence:** defaults < config file < CLI flags < env vars (secrets only)

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

| Option | Default | Description |
|--------|---------|-------------|
| `polling_interval` | `200ms` | How often to poll `hyprctl` for window changes |
| `submission_interval` | `15m` | How often to submit activity data to RescueTime |
| `merge_threshold` | `30s` | Merge sessions of the same app if gap is within this |
| `min_duration` | `10s` | Ignore sessions shorter than this |
| `log_level` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `log_file` | `""` | Path to log file (empty = stdout only) |
| `env_file_path` | `.env` | Path to environment file with API credentials |
| `persistence_path` | `""` | Path for session persistence (empty = XDG default) |
| `legacy_api_endpoint` | `https://www.rescuetime.com/anapi/offline_time_post` | Legacy API endpoint |
| `native_api_endpoint` | `https://api.rescuetime.com/api/resource/user_client_events` | Native API endpoint |

## Usage

### Basic Commands

```bash
# Single window query (shows current active window)
./active-window

# Monitor window changes (display only, no tracking)
./active-window -monitor

# Track time and display summary on exit (Ctrl+C)
./active-window -track

# Track and submit to RescueTime API (production mode)
./active-window -track -submit

# Custom polling interval (default: 200ms)
./active-window -monitor -interval 500ms

# Custom submission interval (default: 15m)
./active-window -track -submit -submission-interval 5m
```

### Advanced Commands

```bash
# Debug logging
./active-window -track -log-level debug

# Log to file
./active-window -track -log-file /tmp/rescuetime.log

# Live stats every 10 seconds
./active-window -track -stats -stats-interval 10s

# Custom config file
./active-window -track -config /path/to/config.json

# Full feature test
./active-window -track -submit -stats -stats-interval 10s -log-level debug -submission-interval 2m
```

### CLI Flags Reference

| Flag | Default | Description |
|------|---------|-------------|
| `-monitor` | `false` | Continuously monitor for window changes |
| `-track` | `false` | Monitor and track time spent in applications |
| `-submit` | `false` | Submit activity data to RescueTime API |
| `-config` | `""` | Config file path (default: `~/.config/rescuetime-linux/config.json`) |
| `-interval` | `""` | Polling interval override (e.g., `100ms`, `1s`) |
| `-submission-interval` | `""` | Submission interval override (e.g., `15m`, `1h`) |
| `-stats` | `false` | Show periodic activity statistics |
| `-stats-interval` | `60s` | Interval for live stats display (e.g., `10s`, `60s`) |
| `-log-level` | `""` | Log level override: `debug`, `info`, `warn`, `error` |
| `-log-file` | `""` | Log file path override |

### Running as a Service

**Systemd service (recommended for autostart):**

```ini
# ~/.config/systemd/user/rescuetime.service
[Unit]
Description=RescueTime Activity Tracker
After=hyprland-session.target

[Service]
Type=simple
ExecStart=/path/to/active-window -track -submit -log-file /tmp/rescuetime.log
Restart=on-failure
RestartSec=10
EnvironmentFile=%h/.config/rescuetime-linux/.env
Environment="WAYLAND_DISPLAY=wayland-1"

[Install]
WantedBy=default.target
```

Enable and start:
```bash
systemctl --user enable rescuetime.service
systemctl --user start rescuetime.service
```

## Architecture

### Source Files

| File | Purpose |
|------|---------|
| `active-window.go` | `HyprlandWindow` struct, window helpers, `monitorWindowChanges`, `main` |
| `tracker.go` | `ActivitySession`, `ActivitySummary`, `ActivityTracker` and all tracking methods |
| `api.go` | API types, activation, submission (parallel with bounded concurrency) |
| `config.go` | `Config` struct, `Duration` JSON wrapper, XDG-aware config loading |
| `logging.go` | Structured logging setup with `log/slog` |
| `persistence.go` | `PersistenceStore` for saving/loading unsubmitted sessions |
| `titleclean.go` | Rule-based window title cleaning (notification counts, browser names) |

### Core Components

**1. Window Monitoring** (`active-window.go`)
- Polls `hyprctl activewindow -j` for active window data
- Returns structured `HyprlandWindow` information
- Configurable polling interval (default: 200ms)

**2. Activity Tracking** (`tracker.go`)
- Thread-safe session management with `sync.RWMutex`
- Automatic session start/end on window focus changes
- Session merging for brief interruptions (< 30s configurable)
- Filters out sessions shorter than minimum duration (< 10s configurable)

**3. Window Title Cleaning** (`titleclean.go`)
- Regex pipeline strips notification counts like `(3)` or `(1,064)`
- Removes trailing browser names (Firefox, Chrome, Brave, etc.)
- Normalizes whitespace and truncates to 255 characters (API limit)

**4. API Submission** (`api.go`)
- Dual-mode: native client API first, legacy fallback
- Bounded concurrency with 3 goroutines via semaphore channel
- Exponential backoff retry (3 attempts: 1s, 2s, 4s)
- 10-second HTTP timeout per request
- Distinguishes retryable (5xx) vs non-retryable (4xx) errors
- Filters activities under 1 minute before submission

**5. Session Persistence** (`persistence.go`)
- Atomic file writes (write `.tmp`, then rename)
- XDG-compliant path: `~/.local/share/rescuetime-linux/sessions.json`
- Automatic recovery on startup; cleared after successful submission

**6. Configuration** (`config.go`)
- JSON config file with `Duration` wrapper for human-readable time strings
- XDG-compliant config directory
- CLI flag overrides via `ApplyCLIFlags()`

**7. Structured Logging** (`logging.go`)
- `log/slog` with configurable levels (`debug`, `info`, `warn`, `error`)
- Multi-writer: stdout + optional log file
- Replaces all `fmt.Printf` for internal logging

### Key Data Structures

```go
// Single continuous session with an application
type ActivitySession struct {
    StartTime   time.Time     `json:"start_time"`
    EndTime     time.Time     `json:"end_time"`
    AppClass    string        `json:"app_class"`
    WindowTitle string        `json:"window_title"`
    Duration    time.Duration `json:"duration"`
    Active      bool          `json:"active"`     // true if session is currently ongoing
}

// Aggregated time across multiple sessions
type ActivitySummary struct {
    AppClass        string        `json:"app_class"`
    ActivityDetails string        `json:"activity_details"`
    TotalDuration   time.Duration `json:"total_duration"`
    SessionCount    int           `json:"session_count"`
    FirstSeen       time.Time     `json:"first_seen"`
    LastSeen        time.Time     `json:"last_seen"`
}

// RescueTime API payload (legacy offline time API)
type RescueTimePayload struct {
    StartTime       string `json:"start_time"`        // "YYYY-MM-DD HH:MM:SS"
    Duration        int    `json:"duration"`           // minutes
    ActivityName    string `json:"activity_name"`      // app class
    ActivityDetails string `json:"activity_details"`   // window title
}

// Native RescueTime user_client_events API format
type UserClientEventPayload struct {
    UserClientEvent UserClientEvent `json:"user_client_event"`
}

type UserClientEvent struct {
    EventDescription string `json:"event_description"` // application class
    StartTime        string `json:"start_time"`        // RFC 3339
    EndTime          string `json:"end_time"`          // RFC 3339
    WindowTitle      string `json:"window_title"`
    Application      string `json:"application"`
}

// Batch submission outcome
type SubmissionResult struct {
    Succeeded      int
    Failed         int
    NativeSuccess  int
    LegacyFallback int
    Errors         []error
}
```

## API Integration

### Dual-Mode Submission

The application uses a **native-first with legacy-fallback** strategy:

1. **Check credentials** — If `RESCUE_TIME_ACCOUNT_KEY` or `RESCUE_TIME_DATA_KEY` is set, attempt the native API
2. **Native API** — `POST https://api.rescuetime.com/api/resource/user_client_events` with RFC 3339 timestamps
3. **Fallback** — If native API fails, automatically retry with the legacy API
4. **Legacy-only** — If no native credentials, use the legacy API directly

Activities under 1 minute are filtered out before submission. All submissions use bounded concurrency (3 goroutines) and exponential backoff retry (3 attempts).

### Legacy API

- **Endpoint:** `https://www.rescuetime.com/anapi/offline_time_post`
- **Auth:** Query parameter `?key=API_KEY`
- **Method:** POST JSON
- **Timestamps:** `YYYY-MM-DD HH:MM:SS` format
- **Duration:** In minutes (rounded up)

### Native Client API

- **Endpoint:** `https://api.rescuetime.com/api/resource/user_client_events`
- **Auth:** Query parameter `?key=ACCOUNT_KEY` (with Bearer token fallback on 401)
- **Method:** POST JSON
- **Timestamps:** RFC 3339 format (`2025-10-02T14:00:00Z`)
- **Duration:** Derived from start/end time

Documentation: `RescueTime-Complete-Authentication-Reverse-Engineering-Report.md`

## Development Status

### Completed (Phases 1-6)
- ✅ Hyprland/Wayland window detection via `hyprctl`
- ✅ Real-time window focus monitoring
- ✅ Activity session tracking with start/end times
- ✅ Session merging for brief interruptions
- ✅ Activity summarization and aggregation
- ✅ Graceful shutdown with summary display
- ✅ RescueTime API integration (legacy offline time POST)
- ✅ Automatic 15-minute submission timer
- ✅ API error handling with exponential backoff
- ✅ Environment-based configuration (.env file)
- ✅ Complete reverse engineering of native client API
- ✅ Session persistence across restarts (atomic file writes)
- ✅ JSON configuration file with XDG support
- ✅ Structured logging via `log/slog` (multi-writer, configurable levels)
- ✅ Window title cleaning (regex pipeline)
- ✅ Live activity statistics display (`-stats`)
- ✅ Dual-mode API submission (native-first + legacy fallback)
- ✅ Parallel bounded concurrency (3 goroutines)

### Pending (Phases 7-8)
- ⏸️ Unit and integration tests
- ⏸️ Performance profiling
- ⏸️ Distribution packaging (AUR, Flatpak, etc.)

Detailed implementation plan: `context/todo/implementation-plan.md`

## Testing

### Manual Testing

```bash
# Short tracking session to verify window detection
./active-window -track
# Switch between windows for ~30 seconds, then Ctrl+C to see summary

# Test API submission with short interval (2 minutes)
./active-window -track -submit -submission-interval 2m
# Use windows for 2+ minutes, verify API submission succeeds
```

### Crash Recovery Test

```bash
# Start tracking in the background
./active-window -track &
# Use windows for 30 seconds, then force-kill
sleep 30 && kill -9 $!

# Restart with submission — should recover saved sessions
./active-window -track -submit -submission-interval 30s
# Look for "loaded saved sessions" in the log output
```

### Debug and Stats Test

```bash
# Debug logging to see all internal activity
./active-window -track -log-level debug

# Live stats with 10-second refresh
./active-window -track -stats -stats-interval 10s
```

### API Testing

HTTP requests for testing authentication and endpoints are in `rescuetime-auth.http` (use with REST client or curl).

## Platform Notes

- **Hyprland-specific:** Uses `hyprctl activewindow -j` command
- **Not portable:** Requires modifications for X11, Windows, or macOS
- **Display required:** Checks for `WAYLAND_DISPLAY` or `DISPLAY` environment variable

## Contributing

Contributions are welcome! Areas of interest:
- Support for other Wayland compositors (Sway, etc.)
- X11/Xorg support
- Unit and integration tests
- Performance profiling and optimization
- Distribution packaging (AUR, Flatpak, etc.)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Documentation

- [Complete Authentication Reverse Engineering Report](RescueTime-Complete-Authentication-Reverse-Engineering-Report.md)
- [RescueTime API Documentation](context/rescuetime/api-docs.md)
- [Implementation Plan](context/todo/implementation-plan.md)
- [Environment Variable Template](.env.example)
- [HTTP Request Examples](rescuetime-auth.http)

## Acknowledgments

- RescueTime for providing time tracking services
- Hyprland compositor for clean JSON output via `hyprctl`
