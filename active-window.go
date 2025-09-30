package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
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

// RescueTimePayload represents the data structure for RescueTime API
type RescueTimePayload struct {
	StartTime       string `json:"start_time"`       // YYYY-MM-DD HH:MM:SS format
	Duration        int    `json:"duration"`         // duration in minutes
	ActivityName    string `json:"activity_name"`    // application class
	ActivityDetails string `json:"activity_details"` // window title/details
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

func monitorWindowChanges(interval time.Duration) {
	var lastAppClass, lastWindowTitle string

	// Create activity tracker
	tracker := NewActivityTracker()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Get initial window info and start first session
	window, err := getActiveWindow()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting initial window info: %v\n", err)
		return
	}

	// Start initial session
	tracker.StartSession(window.Class, window.Title)
	lastAppClass = window.Class
	lastWindowTitle = window.Title

	// Print initial window
	currentInfo := formatWindowOutput(window.Title, window.Class)
	fmt.Printf("%s [%s]\n", currentInfo, time.Now().Format("15:04:05"))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\nShutting down window monitor...")

			// End current session and print summary before exit
			tracker.EndCurrentSession()
			printActivitySummary(tracker)
			return

		case <-ticker.C:
			window, err := getActiveWindow()
			if err != nil {
				// Don't spam errors, just skip this iteration
				continue
			}

			// Check if the application or window title changed
			if window.Class != lastAppClass || window.Title != lastWindowTitle {
				// Start new session for the new window/app
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
	interval := flag.Duration("interval", 200*time.Millisecond, "Polling interval for monitoring mode (e.g., 100ms, 1s)")
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
		monitorWindowChanges(*interval)
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
