# RescueTime Linux Activity Tracker

A native Linux activity tracker for [RescueTime](https://www.rescuetime.com) that monitors active window usage on Hyprland/Wayland compositors and submits time tracking data via the RescueTime API.

> **Status:** Core functionality complete (Phase 1-3). Native client API integration in progress.

## Features

- **Hyprland/Wayland Support** - Monitors active window focus changes using `hyprctl`
- **Smart Session Tracking** - Automatically tracks time spent in each application
- **Intelligent Merging** - Merges brief window switches to the same app (< 30 seconds)
- **Session Filtering** - Ignores very short sessions (< 10 seconds) to reduce noise
- **Automatic Submission** - Sends activity data to RescueTime every 15 minutes (configurable)
- **Graceful Shutdown** - Submits final data on exit (SIGINT/SIGTERM)
- **Retry Logic** - Exponential backoff for failed API submissions

## Requirements

- **OS:** Linux with Wayland
- **Compositor:** Hyprland (with `hyprctl` command)
- **Runtime:** Go 1.16+ (for building)
- **RescueTime Account:** Free or paid account with API access

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/robwilde/rescuetime-linux.git
cd rescuetime-linux

# Build the binary
go build -o active-window active-window.go

# Create environment file
cp .env.example .env
# Edit .env and add your RescueTime API key
```

### Environment Setup

Create a `.env` file in the project directory:

```bash
RESCUE_TIME_API_KEY=your_api_key_here
```

**Getting your API key:**
1. Log in to [RescueTime](https://www.rescuetime.com)
2. Navigate to Settings → API & Integrations
3. Generate or copy your API key

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

### Running as a Service

**Systemd service (recommended for autostart):**

```ini
# ~/.config/systemd/user/rescuetime.service
[Unit]
Description=RescueTime Activity Tracker
After=hyprland-session.target

[Service]
Type=simple
ExecStart=/path/to/active-window -track -submit
Restart=on-failure
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

### Core Components

**1. Window Monitoring** (`active-window.go:236-251`)
- Polls `hyprctl activewindow -j` for active window data
- Returns structured `HyprlandWindow` information
- Configurable polling interval (default: 200ms)

**2. Activity Tracking** (`ActivityTracker`)
- Thread-safe session management with `sync.RWMutex`
- Automatic session start/end on window focus changes
- Session merging for brief interruptions (< 30s)
- Filters out sessions shorter than 10 seconds

**3. Data Aggregation** (`GetActivitySummaries()`)
- Aggregates multiple sessions per application
- Calculates total duration and session counts
- Includes currently active session in real-time

**4. API Submission** (`submitToRescueTime()`)
- Posts to RescueTime Offline Time API
- Exponential backoff retry (3 attempts: 1s, 2s, 4s)
- 10-second HTTP timeout per request
- Distinguishes retryable (5xx) vs non-retryable (4xx) errors

### Key Data Structures

```go
// Single continuous session with an application
type ActivitySession struct {
    AppClass    string
    WindowTitle string
    StartTime   time.Time
    EndTime     time.Time
    Duration    time.Duration
}

// Aggregated time across multiple sessions
type ActivitySummary struct {
    AppClass       string
    TotalDuration  time.Duration
    SessionCount   int
    LastWindowTitle string
    MostRecentTime time.Time
}

// RescueTime API payload
type RescueTimePayload struct {
    StartTime       string `json:"start_time"`        // "YYYY-MM-DD HH:MM:SS"
    Duration        int    `json:"duration"`          // minutes
    ActivityName    string `json:"activity_name"`     // app class
    ActivityDetails string `json:"activity_details"`  // window title
}
```

## API Integration

### Current Implementation (Legacy API)

Uses the public **Offline Time POST API**:
- **Endpoint:** `https://www.rescuetime.com/anapi/offline_time_post`
- **Auth:** Query parameter `?key=API_KEY`
- **Method:** POST JSON
- **Max duration:** 4 hours per entry

### Native Client API (Reverse Engineered)

Documentation available in `RescueTime-Complete-Authentication-Reverse-Engineering-Report.md`

**Authentication Flow:**
```bash
POST https://www.rescuetime.com/activate
Content-Type: application/json
Accept: application/json

{
  "username": "your@email.com",
  "password": "your_password",
  "computer_name": "my-linux-machine"
}

# Response:
{
  "account_key": "186c3aa4fddc9204ea5e6cb2dfb50fa2",  // 32-char hex
  "data_key": "B633XlfzSI__qItgt7BG8IGlvFJLYoQT69seoVwt"   // 44-char base64
}
```

**Event Submission:**
```bash
POST https://api.rescuetime.com/api/resource/user_client_events
Authorization: Bearer {data_key}
Content-Type: application/json; charset=utf-8

{
  "user_client_event": {
    "event_description": "firefox",
    "start_time": "2025-10-02T14:00:00Z",
    "end_time": "2025-10-02T14:05:00Z",
    "window_title": "GitHub - robwilde/rescuetime-linux",
    "application": "firefox"
  }
}
```

## Development Status

### Completed (Phase 1-3)
- ✅ Hyprland/Wayland window detection via `hyprctl`
- ✅ Real-time window focus monitoring
- ✅ Activity session tracking with start/end times
- ✅ Session merging for brief interruptions
- ✅ Activity summarization and aggregation
- ✅ Graceful shutdown with summary display
- ✅ RescueTime API integration (Offline Time POST)
- ✅ Automatic 15-minute submission timer
- ✅ API error handling with exponential backoff
- ✅ Environment-based configuration (.env file)
- ✅ Complete reverse engineering of native client API

### TODO (Phase 4-8)
- ⏸️ Session persistence across restarts
- ⏸️ Configuration file support (YAML/JSON)
- ⏸️ Structured logging (replace fmt.Printf)
- ⏸️ Unit tests
- ⏸️ Migration to native client API
- ⏸️ Systemd service template

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

### API Testing

HTTP requests for testing authentication and endpoints are in `rescuetime-auth.http` (use with REST client or curl).

## Platform Notes

- **Hyprland-specific:** Uses `hyprctl activewindow -j` command
- **Not portable:** Requires modifications for X11, Windows, or macOS
- **Wayland-only:** Checks for `WAYLAND_DISPLAY` environment variable

## Contributing

Contributions are welcome! Areas of interest:
- Support for other Wayland compositors (Sway, etc.)
- X11/Xorg support
- Better error handling and logging
- Unit and integration tests

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Documentation

- [Complete Authentication Reverse Engineering Report](RescueTime-Complete-Authentication-Reverse-Engineering-Report.md)
- [RescueTime API Documentation](context/rescuetime/api-docs.md)
- [HTTP Request Examples](rescuetime-auth.http)

## Acknowledgments

- RescueTime for providing time tracking services
- Hyprland compositor for clean JSON output via `hyprctl`