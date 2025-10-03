package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// HyprlandWindow represents the JSON structure returned by hyprctl activewindow -j
type HyprlandWindow struct {
	Address   string `json:"address"`
	Mapped    bool   `json:"mapped"`
	Hidden    bool   `json:"hidden"`
	At        [2]int `json:"at"`
	Size      [2]int `json:"size"`
	Workspace struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"workspace"`
	Floating     bool   `json:"floating"`
	Pseudo       bool   `json:"pseudo"`
	Monitor      int    `json:"monitor"`
	Class        string `json:"class"`
	Title        string `json:"title"`
	InitialClass string `json:"initialClass"`
	InitialTitle string `json:"initialTitle"`
	Pid          int    `json:"pid"`
	Xwayland     bool   `json:"xwayland"`
	Pinned       bool   `json:"pinned"`
	Fullscreen   int    `json:"fullscreen"`
}

// ActivitySession represents a single continuous session with an application
type ActivitySession struct {
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	AppClass    string        `json:"app_class"`
	WindowTitle string        `json:"window_title"`
	Duration    time.Duration `json:"duration"`
	Active      bool          `json:"active"` // true if session is currently ongoing
}

// ActivitySummary represents aggregated time spent in an application
type ActivitySummary struct {
	AppClass        string        `json:"app_class"`
	ActivityDetails string        `json:"activity_details"`
	TotalDuration   time.Duration `json:"total_duration"`
	SessionCount    int           `json:"session_count"`
	FirstSeen       time.Time     `json:"first_seen"`
	LastSeen        time.Time     `json:"last_seen"`
}

// ActivityTracker manages tracking of application usage sessions
type ActivityTracker struct {
	mu             sync.RWMutex
	currentSession *ActivitySession
	sessions       []ActivitySession
	mergeThreshold time.Duration // merge sessions shorter than this threshold
	minDuration    time.Duration // ignore sessions shorter than this
}

// RescueTimePayload represents the data structure for RescueTime API (legacy offline time API)
type RescueTimePayload struct {
	StartTime       string `json:"start_time"`       // YYYY-MM-DD HH:MM:SS format
	Duration        int    `json:"duration"`         // duration in minutes
	ActivityName    string `json:"activity_name"`    // application class
	ActivityDetails string `json:"activity_details"` // window title/details
}

// UserClientEventPayload represents the native RescueTime user_client_events API format
type UserClientEventPayload struct {
	UserClientEvent UserClientEvent `json:"user_client_event"`
}

// UserClientEvent represents a single activity tracking event
type UserClientEvent struct {
	EventDescription string `json:"event_description"` // application class
	StartTime        string `json:"start_time"`        // RFC 3339 format: 2025-09-30T12:00:00Z
	EndTime          string `json:"end_time"`          // RFC 3339 format: 2025-09-30T12:01:00Z
	WindowTitle      string `json:"window_title"`      // window title
	Application      string `json:"application"`       // application class (redundant with event_description)
}

// ActivationRequest represents the payload for the /activate endpoint
type ActivationRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ActivationResponse represents the response from the /activate endpoint
type ActivationResponse struct {
	AccountKey string `json:"account_key"`
	DataKey    string `json:"data_key"`
	ApiURL     string `json:"api_url"`
	URL        string `json:"url"`
}

// activateWithRescueTime authenticates with RescueTime and retrieves account keys
func activateWithRescueTime(email, password string) (*ActivationResponse, error) {
	// Discovered through testing: endpoint uses form-encoded data with username/password fields
	url := "https://api.rescuetime.com/activate"

	// Create form-encoded payload
	formData := fmt.Sprintf("username=%s&password=%s",
		strings.ReplaceAll(email, "@", "%40"), // URL encode @ sign
		password)

	// Create request
	req, err := http.NewRequest("POST", url, strings.NewReader(formData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "RescueTime/2.16.5.1 (Linux)")

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Check for error in response
	// Response format is YAML-like: "c:\n- 0\n- RT:ok\naccount_key: xxx\nkey: xxx"
	bodyStr := string(body)
	if strings.Contains(bodyStr, "RT:error") {
		return nil, fmt.Errorf("activation failed: %s", bodyStr)
	}

	// Parse response to extract account_key
	// TODO: The response only contains account_key, not data_key
	// We need to discover how to obtain the data_key (separate endpoint? different auth flow?)
	var accountKey string
	for _, line := range strings.Split(bodyStr, "\n") {
		if strings.HasPrefix(line, "account_key:") {
			accountKey = strings.TrimSpace(strings.TrimPrefix(line, "account_key:"))
			break
		}
	}

	if accountKey == "" {
		return nil, fmt.Errorf("no account_key in response: %s", bodyStr)
	}

	// Return response with account_key
	// Note: data_key is empty - needs further investigation
	return &ActivationResponse{
		AccountKey: accountKey,
		DataKey:    "", // TODO: Discover how to obtain data_key
		ApiURL:     "api.rescuetime.com",
		URL:        "www.rescuetime.com",
	}, nil
}

// saveCredentialsToEnv saves the activation credentials to .env file
func saveCredentialsToEnv(filepath string, response *ActivationResponse) error {
	// Read existing .env file to preserve RESCUE_TIME_API_KEY if it exists
	existingVars := make(map[string]string)

	file, err := os.Open(filepath)
	if err == nil {
		// File exists, read it
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				existingVars[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
		file.Close()
	}

	// Update with new credentials
	existingVars["RESCUE_TIME_ACCOUNT_KEY"] = response.AccountKey
	existingVars["RESCUE_TIME_DATA_KEY"] = response.DataKey

	// Write back to file
	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create .env file: %v", err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)

	// Write header
	fmt.Fprintln(writer, "# RescueTime API Credentials")
	fmt.Fprintln(writer, "# Generated by active-window")
	fmt.Fprintln(writer, "")

	// Write all variables
	for key, value := range existingVars {
		fmt.Fprintf(writer, "%s=%s\n", key, value)
	}

	return writer.Flush()
}

// loadEnvFile loads environment variables from a .env file
func loadEnvFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open .env file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first '=' sign
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Set environment variable
		os.Setenv(key, value)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %v", err)
	}

	return nil
}

// summaryToPayload converts an ActivitySummary to RescueTimePayload format (legacy)
func summaryToPayload(summary ActivitySummary) RescueTimePayload {
	// Convert duration to minutes (rounded up)
	durationMinutes := int(math.Ceil(summary.TotalDuration.Minutes()))

	// Format start time as "YYYY-MM-DD HH:MM:SS"
	startTimeFormatted := summary.FirstSeen.Format("2006-01-02 15:04:05")

	return RescueTimePayload{
		StartTime:       startTimeFormatted,
		Duration:        durationMinutes,
		ActivityName:    summary.AppClass,
		ActivityDetails: summary.ActivityDetails,
	}
}

// summaryToUserClientEvent converts an ActivitySummary to UserClientEventPayload format
func summaryToUserClientEvent(summary ActivitySummary) UserClientEventPayload {
	// Calculate end time: start time + total duration
	endTime := summary.FirstSeen.Add(summary.TotalDuration)

	// Format timestamps in RFC 3339 (ISO 8601) format with UTC timezone
	startTimeFormatted := summary.FirstSeen.UTC().Format(time.RFC3339)
	endTimeFormatted := endTime.UTC().Format(time.RFC3339)

	return UserClientEventPayload{
		UserClientEvent: UserClientEvent{
			EventDescription: summary.AppClass,
			StartTime:        startTimeFormatted,
			EndTime:          endTimeFormatted,
			WindowTitle:      summary.ActivityDetails,
			Application:      summary.AppClass, // Same as EventDescription
		},
	}
}

// submitToRescueTime submits activity data to RescueTime API with retry logic (legacy offline time API)
func submitToRescueTime(apiKey string, payload RescueTimePayload) error {
	const maxRetries = 3
	const baseDelay = 1 * time.Second

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			delay := baseDelay * time.Duration(math.Pow(2, float64(attempt-1)))
			fmt.Printf("Retrying in %v... (attempt %d/%d)\n", delay, attempt+1, maxRetries)
			time.Sleep(delay)
		}

		// Convert payload to JSON
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %v", err)
		}

		// Create request
		url := fmt.Sprintf("https://www.rescuetime.com/anapi/offline_time_post?key=%s", apiKey)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %v", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		// Send request
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %v", err)
			continue
		}

		// Read response body
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Check response status
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Printf("✓ Submitted to RescueTime: %s (%d min)\n", payload.ActivityName, payload.Duration)
			return nil
		}

		lastErr = fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))

		// Don't retry on client errors (4xx)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return lastErr
		}
	}

	return fmt.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
}

// submitUserClientEvent submits activity data to native RescueTime user_client_events API
func submitUserClientEvent(apiKey string, payload UserClientEventPayload) error {
	const maxRetries = 3
	const baseDelay = 1 * time.Second

	var lastErr error
	var tryBearerAuth bool

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			delay := baseDelay * time.Duration(math.Pow(2, float64(attempt-1)))
			fmt.Printf("Retrying in %v... (attempt %d/%d)\n", delay, attempt+1, maxRetries)
			time.Sleep(delay)
		}

		// Convert payload to JSON
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %v", err)
		}

		var req *http.Request

		// Try Bearer token auth if query param auth failed with 401
		if tryBearerAuth {
			// Create request WITHOUT query parameter
			url := "https://api.rescuetime.com/api/resource/user_client_events"
			req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
			if err != nil {
				lastErr = fmt.Errorf("failed to create request: %v", err)
				continue
			}
			// Use Bearer token authentication with data_key
			// The desktop app uses the data_key as the Bearer token
			dataKey := os.Getenv("RESCUE_TIME_DATA_KEY")
			if dataKey == "" {
				dataKey = apiKey // Fallback to provided API key
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", dataKey))

			// Also try adding account_key as a query parameter along with Bearer token
			accountKey := os.Getenv("RESCUE_TIME_ACCOUNT_KEY")
			if accountKey != "" {
				req.URL.RawQuery = fmt.Sprintf("key=%s", accountKey)
			}
		} else {
			// Try query parameter authentication first with account_key
			authKey := os.Getenv("RESCUE_TIME_ACCOUNT_KEY")
			if authKey == "" {
				authKey = apiKey
			}
			url := fmt.Sprintf("https://api.rescuetime.com/api/resource/user_client_events?key=%s", authKey)
			req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
			if err != nil {
				lastErr = fmt.Errorf("failed to create request: %v", err)
				continue
			}
		}

		// Set headers matching the official app
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("User-Agent", "RescueTime/2.16.5.1 (Linux)")

		// Send request
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %v", err)
			continue
		}

		// Read response body
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Check response status
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			authMethod := "query parameter"
			if tryBearerAuth {
				authMethod = "Bearer token"
			}
			fmt.Printf("✓ Submitted to RescueTime via %s: %s (%s to %s)\n",
				authMethod,
				payload.UserClientEvent.Application,
				payload.UserClientEvent.StartTime,
				payload.UserClientEvent.EndTime)
			return nil
		}

		lastErr = fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))

		// If we got 401 with query param auth, try Bearer token auth next
		if resp.StatusCode == 401 && !tryBearerAuth {
			fmt.Println("Query parameter auth failed (401), trying Bearer token authentication...")
			tryBearerAuth = true
			continue
		}

		// Don't retry on other client errors (4xx)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return lastErr
		}
	}

	return fmt.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
}

// submitActivitiesToRescueTime submits all activity summaries to RescueTime
// Attempts native user_client_events API first if credentials are available,
// falls back to offline_time_post API if native fails or credentials are missing.
func submitActivitiesToRescueTime(apiKey string, summaries map[string]ActivitySummary) {
	if len(summaries) == 0 {
		fmt.Println("No activities to submit.")
		return
	}

	// Check if we have native API credentials
	dataKey := os.Getenv("RESCUE_TIME_DATA_KEY")
	accountKey := os.Getenv("RESCUE_TIME_ACCOUNT_KEY")
	hasNativeCredentials := dataKey != "" || accountKey != ""

	fmt.Printf("\n=== Submitting %d activities to RescueTime ===\n", len(summaries))
	if hasNativeCredentials {
		fmt.Println("[INFO] Native API credentials detected, will try native API first with legacy fallback")
	} else {
		fmt.Println("[INFO] Using legacy offline time API (no native credentials found)")
	}

	successCount := 0
	failCount := 0
	nativeSuccessCount := 0
	legacyFallbackCount := 0

	for _, summary := range summaries {
		// Skip activities with a very short duration (< 1 minute)
		if summary.TotalDuration < time.Minute {
			continue
		}

		var err error
		usedFallback := false

		if hasNativeCredentials {
			// Try native API first
			fmt.Printf("[ATTEMPT] Trying native API for %s...\n", summary.AppClass)
			payload := summaryToUserClientEvent(summary)
			err = submitUserClientEvent(apiKey, payload)

			if err != nil {
				// Native API failed, log and try legacy fallback
				fmt.Fprintf(os.Stderr, "[WARN] Native API failed for %s: %v\n", summary.AppClass, err)
				fmt.Printf("[FALLBACK] Attempting legacy API for %s...\n", summary.AppClass)

				legacyPayload := summaryToPayload(summary)
				err = submitToRescueTime(apiKey, legacyPayload)
				usedFallback = true
			} else {
				nativeSuccessCount++
			}
		} else {
			// No native credentials, use legacy API directly
			payload := summaryToPayload(summary)
			err = submitToRescueTime(apiKey, payload)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Failed to submit %s: %v\n", summary.AppClass, err)
			failCount++
		} else {
			successCount++
			if usedFallback {
				legacyFallbackCount++
			}
		}
	}

	fmt.Printf("\n=== Submission Summary ===\n")
	fmt.Printf("Total succeeded: %d, failed: %d\n", successCount, failCount)
	if hasNativeCredentials {
		fmt.Printf("Native API successes: %d\n", nativeSuccessCount)
		fmt.Printf("Legacy fallback successes: %d\n", legacyFallbackCount)
	}
}

// NewActivityTracker creates a new activity tracker with default settings
func NewActivityTracker() *ActivityTracker {
	return &ActivityTracker{
		sessions:       make([]ActivitySession, 0),
		mergeThreshold: 30 * time.Second, // merge sessions if gap is less than 30s
		minDuration:    10 * time.Second, // ignore sessions shorter than 10s
	}
}

// StartSession begins tracking a new activity session
func (at *ActivityTracker) StartSession(appClass, windowTitle string) {
	at.mu.Lock()
	defer at.mu.Unlock()

	now := time.Now()

	// End the current session if one exists
	if at.currentSession != nil && at.currentSession.Active {
		at.endCurrentSessionUnsafe(now)
	}

	// Start new session
	at.currentSession = &ActivitySession{
		StartTime:   now,
		AppClass:    appClass,
		WindowTitle: windowTitle,
		Active:      true,
	}
}

// endCurrentSessionUnsafe ends the current session (must be called with lock held)
func (at *ActivityTracker) endCurrentSessionUnsafe(endTime time.Time) {
	if at.currentSession == nil || !at.currentSession.Active {
		return
	}

	at.currentSession.EndTime = endTime
	at.currentSession.Duration = endTime.Sub(at.currentSession.StartTime)
	at.currentSession.Active = false

	// Only store sessions that meet minimum duration requirement
	if at.currentSession.Duration >= at.minDuration {
		// Check if we should merge with the last session
		if at.shouldMergeWithLastSession() {
			at.mergeWithLastSession()
		} else {
			// Store the session
			at.sessions = append(at.sessions, *at.currentSession)
		}
	}
}

// EndCurrentSession ends the currently active session
func (at *ActivityTracker) EndCurrentSession() {
	at.mu.Lock()
	defer at.mu.Unlock()
	at.endCurrentSessionUnsafe(time.Now())
}

// shouldMergeWithLastSession checks if current session should be merged with the previous one
func (at *ActivityTracker) shouldMergeWithLastSession() bool {
	if len(at.sessions) == 0 || at.currentSession == nil {
		return false
	}

	lastSession := &at.sessions[len(at.sessions)-1]

	// Can only merge sessions of the same application
	if lastSession.AppClass != at.currentSession.AppClass {
		return false
	}

	// Check if the gap between sessions is within merge threshold
	gap := at.currentSession.StartTime.Sub(lastSession.EndTime)
	return gap <= at.mergeThreshold
}

// mergeWithLastSession merges current session with the last stored session
func (at *ActivityTracker) mergeWithLastSession() {
	if len(at.sessions) == 0 || at.currentSession == nil {
		return
	}

	lastSession := &at.sessions[len(at.sessions)-1]

	// Extend the last session to include the current session
	lastSession.EndTime = at.currentSession.EndTime
	lastSession.Duration = lastSession.EndTime.Sub(lastSession.StartTime)

	// Use the most recent window title
	lastSession.WindowTitle = at.currentSession.WindowTitle
}

// GetActivitySummaries aggregates sessions by application class
func (at *ActivityTracker) GetActivitySummaries() map[string]ActivitySummary {
	at.mu.RLock()
	defer at.mu.RUnlock()

	summaries := make(map[string]ActivitySummary)

	// Process all completed sessions
	for _, session := range at.sessions {
		key := session.AppClass
		summary, exists := summaries[key]

		if !exists {
			summary = ActivitySummary{
				AppClass:        session.AppClass,
				ActivityDetails: session.WindowTitle,
				FirstSeen:       session.StartTime,
				LastSeen:        session.EndTime,
			}
		}

		// Update summary
		summary.TotalDuration += session.Duration
		summary.SessionCount++

		// Update time boundaries
		if session.StartTime.Before(summary.FirstSeen) {
			summary.FirstSeen = session.StartTime
		}
		if session.EndTime.After(summary.LastSeen) {
			summary.LastSeen = session.EndTime
			// Use the most recent window title as activity details
			summary.ActivityDetails = session.WindowTitle
		}

		summaries[key] = summary
	}

	// Include current active session if exists
	if at.currentSession != nil && at.currentSession.Active {
		key := at.currentSession.AppClass
		summary, exists := summaries[key]

		currentDuration := time.Since(at.currentSession.StartTime)

		if !exists {
			summary = ActivitySummary{
				AppClass:        at.currentSession.AppClass,
				ActivityDetails: at.currentSession.WindowTitle,
				FirstSeen:       at.currentSession.StartTime,
				LastSeen:        time.Now(),
			}
		}

		summary.TotalDuration += currentDuration
		summary.SessionCount++

		// Update activity details to current window title
		summary.ActivityDetails = at.currentSession.WindowTitle
		summary.LastSeen = time.Now()

		summaries[key] = summary
	}

	return summaries
}

// ClearCompletedSessions removes all completed sessions, keeping only the current active session
func (at *ActivityTracker) ClearCompletedSessions() {
	at.mu.Lock()
	defer at.mu.Unlock()

	// Clear all stored sessions but keep the current active one
	at.sessions = make([]ActivitySession, 0)
}

func getActiveWindow() (*HyprlandWindow, error) {
	// Use hyprctl to get active window information in JSON format
	cmd := exec.Command("hyprctl", "activewindow", "-j")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get active window from hyprctl: %v", err)
	}

	var window HyprlandWindow
	err = json.Unmarshal(output, &window)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hyprctl JSON output: %v", err)
	}

	return &window, nil
}

func getActiveWindowName() (string, error) {
	window, err := getActiveWindow()
	if err != nil {
		return "", err
	}
	return window.Title, nil
}

func getActiveWindowClass() (string, error) {
	window, err := getActiveWindow()
	if err != nil {
		return "", err
	}
	return window.Class, nil
}

func formatWindowOutput(windowName, windowClass string) string {
	if windowClass != "" {
		return fmt.Sprintf("Active Window: %s (%s)", windowName, windowClass)
	}
	return fmt.Sprintf("Active Window: %s", windowName)
}

// printActivitySummary prints a summary of tracked activities
func printActivitySummary(tracker *ActivityTracker) {
	fmt.Println("\n=== Activity Summary ===")

	summaries := tracker.GetActivitySummaries()
	if len(summaries) == 0 {
		fmt.Println("No activities tracked.")
		return
	}

	totalTime := time.Duration(0)
	for _, summary := range summaries {
		totalTime += summary.TotalDuration
	}

	fmt.Printf("Total tracking time: %v\n\n", totalTime.Round(time.Second))

	for appClass, summary := range summaries {
		percentage := float64(summary.TotalDuration) / float64(totalTime) * 100
		fmt.Printf("%s: %v (%.1f%%) - %d sessions\n",
			appClass,
			summary.TotalDuration.Round(time.Second),
			percentage,
			summary.SessionCount)
		fmt.Printf("  └─ %s\n\n", summary.ActivityDetails)
	}
}

func getCurrentWindowInfo() (string, error) {
	windowName, err := getActiveWindowName()
	if err != nil {
		return "", err
	}

	windowClass, _ := getActiveWindowClass()
	return formatWindowOutput(windowName, windowClass), nil
}

func monitorWindowChanges(interval time.Duration, submitToAPI bool, apiKey string, submissionInterval time.Duration) {
	var lastAppClass, lastWindowTitle string

	// Create activity tracker
	tracker := NewActivityTracker()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Get initial window info and start the first session
	window, err := getActiveWindow()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting initial window info: %v\n", err)
		return
	}

	// Start the initial session
	tracker.StartSession(window.Class, window.Title)
	lastAppClass = window.Class
	lastWindowTitle = window.Title

	// Print initial window
	currentInfo := formatWindowOutput(window.Title, window.Class)
	fmt.Printf("%s [%s]\n", currentInfo, time.Now().Format("15:04:05"))

	pollTicker := time.NewTicker(interval)
	defer pollTicker.Stop()

	var submitTicker *time.Ticker
	var submitChan <-chan time.Time

	if submitToAPI {
		submitTicker = time.NewTicker(submissionInterval)
		defer submitTicker.Stop()
		submitChan = submitTicker.C
		fmt.Printf("API submission enabled: will submit every %v\n", submissionInterval)
	}

	for {
		select {
		case <-sigChan:
			fmt.Println("\nShutting down window monitor...")

			// End the current session
			tracker.EndCurrentSession()

			// Submit final data if API submission is enabled
			if submitToAPI {
				summaries := tracker.GetActivitySummaries()
				submitActivitiesToRescueTime(apiKey, summaries)
			}

			// Print summary before exit
			printActivitySummary(tracker)
			return

		case <-submitChan:
			// Time to submit data to RescueTime
			summaries := tracker.GetActivitySummaries()
			submitActivitiesToRescueTime(apiKey, summaries)

			// Clear completed sessions after successful submission
			tracker.ClearCompletedSessions()

		case <-pollTicker.C:
			window, err := getActiveWindow()
			if err != nil {
				// Don't spam errors, just skip this iteration
				continue
			}

			// Check if the application or window title changed
			if window.Class != lastAppClass || window.Title != lastWindowTitle {
				// Start a new session for the new window/app
				tracker.StartSession(window.Class, window.Title)

				// Print the change
				currentInfo := formatWindowOutput(window.Title, window.Class)
				fmt.Printf("%s [%s]\n", currentInfo, time.Now().Format("15:04:05"))

				// Update tracking variables
				lastAppClass = window.Class
				lastWindowTitle = window.Title
			}
		}
	}
}

func main() {
	// Command line flags
	monitor := flag.Bool("monitor", false, "Continuously monitor for window changes")
	track := flag.Bool("track", false, "Monitor and track time spent in applications")
	submit := flag.Bool("submit", false, "Submit activity data to RescueTime API")
	interval := flag.Duration("interval", 200*time.Millisecond, "Polling interval for monitoring mode (e.g., 100ms, 1s)")
	submissionInterval := flag.Duration("submission-interval", 15*time.Minute, "Interval for submitting data to RescueTime (e.g., 15m, 1h)")
	flag.Parse()

	// Check if we're running in a graphical environment (Wayland or X11)
	if os.Getenv("WAYLAND_DISPLAY") == "" && os.Getenv("DISPLAY") == "" {
		fmt.Fprintf(os.Stderr, "Error: No graphical display found. Make sure you're running this in a Wayland or X11 environment.\n")
		os.Exit(1)
	}

	// Check if hyprctl is available (required for Wayland/Hyprland)
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		_, err := exec.LookPath("hyprctl")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: hyprctl not found. This script requires Hyprland on Wayland.\n")
			os.Exit(1)
		}
	}

	if *monitor || *track {
		if *track {
			fmt.Printf("Tracking application usage (polling every %v). Press Ctrl+C to stop and see summary.\n", *interval)
		} else {
			fmt.Printf("Monitoring window changes (polling every %v). Press Ctrl+C to stop.\n", *interval)
		}

		// Handle API submission setup
		var apiKey string
		if *submit {
			// Load environment variables from .env file
			err := loadEnvFile(".env")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading .env file: %v\n", err)
				os.Exit(1)
			}

			// Get API key from environment
			apiKey = os.Getenv("RESCUE_TIME_API_KEY")
			if apiKey == "" {
				fmt.Fprintf(os.Stderr, "Error: RESCUE_TIME_API_KEY not found in .env file\n")
				os.Exit(1)
			}

			// Call with API submission enabled
			monitorWindowChanges(*interval, true, apiKey, *submissionInterval)
		} else {
			// Call without API submission
			monitorWindowChanges(*interval, false, "", 0)
		}
	} else {
		// Single execution mode
		currentInfo, err := getCurrentWindowInfo()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting window info: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(currentInfo)
	}
}
