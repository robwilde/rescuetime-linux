package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
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
