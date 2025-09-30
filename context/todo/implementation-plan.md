# RescueTime Activity Recorder - Implementation Plan

## Project Overview
Create a comprehensive activity tracking system that monitors window focus changes on Hyprland/Wayland, accumulates time spent in each application, and automatically submits this data to RescueTime every 15 minutes using their Offline Time POST API.

## Key Requirements Analysis
- **Window Monitoring**: Already working with `./active-window -monitor`
- **Time Tracking**: Accumulate time spent per application/window
- **Data Parsing**: Extract application name and activity details from window info
- **API Integration**: Submit to RescueTime's Offline Time POST API every 15 minutes
- **Data Format**: Use application class as `activity_name` and window title as `activity_details`

## Implementation Tasks

### Phase 1: Core Infrastructure Setup ✅
- [x] **Task 1.1**: Active window monitoring (COMPLETED)
  - Working Hyprland/Wayland window detection
  - Real-time window focus change monitoring
  - JSON-based window information extraction

### Phase 2: Time Tracking & Data Management
- [x] **Task 2.1**: Create time tracking data structures
  - **Details**: Design Go structs to store activity sessions with start/end times
  - **Files**: Extend `active-window.go` or create separate `activity-tracker.go`
  - **Acceptance Criteria**: 
    - Track when an activity starts and ends
    - Calculate duration for each activity session
    - Handle window focus changes seamlessly

- [ ] **Task 2.2**: Implement activity session management
  - **Details**: Track continuous sessions per application, merge short switches
  - **Logic**: If user switches back to same app within 30 seconds, merge sessions
  - **Acceptance Criteria**:
    - Start new session when window focus changes
    - End previous session with accurate timing
    - Merge brief interruptions (< 30 seconds)

- [ ] **Task 2.3**: Create data aggregation system
  - **Details**: Accumulate total time per application over 15-minute periods
  - **Data Structure**: Map of application -> total duration for current period
  - **Acceptance Criteria**:
    - Aggregate multiple sessions per application
    - Reset accumulation every 15 minutes after API submission
    - Handle edge cases (sessions spanning multiple periods)

### Phase 3: RescueTime API Integration
- [ ] **Task 3.1**: Configure API credentials
  - **Details**: Set up RescueTime API key management
  - **Implementation**: Environment variable or config file for API key
  - **Security**: Never log or expose API key in plain text
  - **Acceptance Criteria**: Secure API key storage and retrieval

- [ ] **Task 3.2**: Implement RescueTime API client
  - **Details**: Create HTTP client for Offline Time POST API
  - **Endpoint**: `https://www.rescuetime.com/anapi/offline_time_post`
  - **Format**: JSON POST with required fields:
    - `start_time`: "YYYY-MM-DD HH:MM:SS"
    - `duration`: minutes (integer)
    - `activity_name`: application class (e.g., "dev.warp.Warp")
    - `activity_details`: window title/description
  - **Acceptance Criteria**: Successfully POST activity data to RescueTime

- [ ] **Task 3.3**: Implement retry logic and error handling
  - **Details**: Handle API failures gracefully with exponential backoff
  - **Edge Cases**: Network issues, API rate limits, malformed data
  - **Acceptance Criteria**: 
    - Retry failed requests up to 3 times
    - Log errors without exposing sensitive data
    - Continue tracking even if API submission fails

### Phase 4: Data Processing & Formatting
- [ ] **Task 4.1**: Parse window information for RescueTime format
  - **Details**: Extract meaningful activity names and details from window data
  - **Examples**:
    - `jetbrains-phpstorm` + `can-eye-budget – README.md` → activity: "jetbrains-phpstorm", details: "can-eye-budget – README.md"
    - `wavebox` + `Inbox (1,064) - robert@mrwilde.com - MrWilde Mail` → activity: "wavebox", details: "MrWilde Mail - Inbox"
  - **Acceptance Criteria**: Clean, readable activity names and descriptions

- [ ] **Task 4.2**: Implement activity duration calculation
  - **Details**: Convert session times to minutes for RescueTime API
  - **Edge Cases**: Sessions less than 1 minute, sessions spanning multiple periods
  - **Acceptance Criteria**: Accurate minute-based duration calculation

- [ ] **Task 4.3**: Create data payload formatting
  - **Details**: Format accumulated activity data into RescueTime API payload
  - **Structure**: Array of activity objects for batch submission
  - **Acceptance Criteria**: Valid JSON payload matching API specification

### Phase 5: Scheduling & Background Operation
- [ ] **Task 5.1**: Implement 15-minute submission timer
  - **Details**: Create background goroutine that submits data every 15 minutes
  - **Implementation**: Use `time.Ticker` for consistent intervals
  - **Acceptance Criteria**: Reliable 15-minute intervals regardless of system activity

- [ ] **Task 5.2**: Add graceful shutdown handling
  - **Details**: Handle SIGINT/SIGTERM to submit final data before exit
  - **Implementation**: Signal handlers to flush remaining activity data
  - **Acceptance Criteria**: No data loss when application is terminated

- [ ] **Task 5.3**: Add persistence for reliability
  - **Details**: Optional local storage to survive application restarts
  - **Implementation**: JSON file or SQLite database for session backup
  - **Acceptance Criteria**: Resume tracking after unexpected shutdown

### Phase 6: Configuration & Logging
- [ ] **Task 6.1**: Create configuration system
  - **Details**: Configurable submission interval, API endpoints, logging level
  - **Format**: YAML or JSON config file with sensible defaults
  - **Options**:
    - `submission_interval`: default 15 minutes
    - `polling_interval`: default 200ms
    - `merge_threshold`: default 30 seconds
    - `api_endpoint`: RescueTime API URL
    - `log_level`: info/debug/error
  - **Acceptance Criteria**: Flexible configuration without code changes

- [ ] **Task 6.2**: Implement structured logging
  - **Details**: Comprehensive logging for debugging and monitoring
  - **Categories**: Activity tracking, API submissions, errors, performance
  - **Format**: Structured logs with timestamps, levels, and context
  - **Acceptance Criteria**: Clear logs for troubleshooting without sensitive data

- [ ] **Task 6.3**: Add activity statistics and reporting
  - **Details**: Optional real-time statistics display
  - **Metrics**: Current session, daily totals, submission status
  - **Implementation**: Optional verbose mode or separate stats command
  - **Acceptance Criteria**: Useful insights into tracking performance

### Phase 7: Testing & Validation
- [ ] **Task 7.1**: Create unit tests for core functions
  - **Details**: Test activity tracking, duration calculation, API formatting
  - **Coverage**: All critical path functions
  - **Acceptance Criteria**: >90% test coverage for core functionality

- [ ] **Task 7.2**: Implement integration testing
  - **Details**: Test full workflow with mock RescueTime API
  - **Scenarios**: Window switching, API submission, error recovery
  - **Acceptance Criteria**: End-to-end functionality verification

- [ ] **Task 7.3**: Performance validation
  - **Details**: Ensure minimal system impact during continuous monitoring
  - **Metrics**: CPU usage, memory usage, response time
  - **Acceptance Criteria**: <1% CPU usage, <10MB memory usage

### Phase 8: Documentation & Deployment
- [ ] **Task 8.1**: Create user documentation
  - **Details**: Installation guide, configuration options, troubleshooting
  - **Format**: README.md with clear examples
  - **Acceptance Criteria**: User can set up and run the system from documentation

- [ ] **Task 8.2**: Create systemd service configuration
  - **Details**: Allow running as background system service
  - **Implementation**: Service file with proper dependencies and restart policies
  - **Acceptance Criteria**: Automatic startup and reliable background operation

- [ ] **Task 8.3**: Package for distribution
  - **Details**: Create release builds and installation scripts
  - **Options**: Binary releases, AUR package for Arch Linux
  - **Acceptance Criteria**: Easy installation for end users

## Technical Architecture

### Data Flow
1. **Window Monitor** → detects focus changes every 200ms
2. **Activity Tracker** → accumulates session durations
3. **Data Aggregator** → summarizes activities every 15 minutes
4. **API Client** → submits data to RescueTime
5. **Logger** → records all operations for debugging

### Key Data Structures
```go
type ActivitySession struct {
    StartTime    time.Time
    EndTime      time.Time
    AppClass     string
    WindowTitle  string
    Duration     time.Duration
}

type ActivitySummary struct {
    AppClass        string
    ActivityDetails string
    TotalDuration   time.Duration
    SessionCount    int
}

type RescueTimePayload struct {
    StartTime       string `json:"start_time"`
    Duration        int    `json:"duration"`
    ActivityName    string `json:"activity_name"`
    ActivityDetails string `json:"activity_details"`
}
```

### Configuration Schema
```yaml
api:
  key: "${RESCUETIME_API_KEY}"
  endpoint: "https://www.rescuetime.com/anapi/offline_time_post"
  
tracking:
  submission_interval: "15m"
  polling_interval: "200ms"
  merge_threshold: "30s"
  min_duration: "10s"
  
logging:
  level: "info"
  file: "rescuetime-tracker.log"
  
storage:
  persist_sessions: true
  backup_file: "sessions.json"
```

## Success Criteria
- [x] Active window detection working on Hyprland/Wayland
- [ ] Accurate time tracking per application
- [ ] Successful RescueTime API integration
- [ ] Reliable 15-minute data submission
- [ ] Minimal system resource usage
- [ ] Comprehensive error handling and logging
- [ ] User-friendly configuration and deployment

## Timeline Estimate
- **Phase 2-3**: 2-3 days (Core tracking and API integration)
- **Phase 4-5**: 1-2 days (Data processing and scheduling)
- **Phase 6-7**: 1-2 days (Configuration and testing)
- **Phase 8**: 1 day (Documentation and deployment)
- **Total**: 5-8 days for complete implementation

## Next Steps
1. Start with **Task 2.1**: Create time tracking data structures
2. Test with short intervals (1-2 minutes) before using 15-minute intervals
3. Use RescueTime API in test mode first to avoid polluting real data
4. Implement logging early for debugging complex timing issues