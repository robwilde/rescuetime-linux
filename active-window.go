package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
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

func monitorWindowChanges(cfg Config, submitToAPI bool, apiKey string) {
	var lastAppClass, lastWindowTitle string

	// Create activity tracker with config values
	tracker := NewActivityTracker(cfg.MergeThreshold.Duration, cfg.MinDuration.Duration)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Get initial window info and start the first session
	window, err := getActiveWindow()
	if err != nil {
		slog.Error("failed to get initial window info", "error", err)
		return
	}

	// Start the initial session
	tracker.StartSession(window.Class, window.Title)
	lastAppClass = window.Class
	lastWindowTitle = window.Title

	// Print initial window (user-facing)
	currentInfo := formatWindowOutput(window.Title, window.Class)
	fmt.Printf("%s [%s]\n", currentInfo, time.Now().Format("15:04:05"))

	pollTicker := time.NewTicker(cfg.PollingInterval.Duration)
	defer pollTicker.Stop()

	var submitTicker *time.Ticker
	var submitChan <-chan time.Time

	if submitToAPI {
		submitTicker = time.NewTicker(cfg.SubmissionInterval.Duration)
		defer submitTicker.Stop()
		submitChan = submitTicker.C
		slog.Info("API submission enabled", "interval", cfg.SubmissionInterval.Duration)
	}

	for {
		select {
		case <-sigChan:
			slog.Info("shutting down window monitor")

			// End the current session
			tracker.EndCurrentSession()

			// Submit final data if API submission is enabled
			if submitToAPI {
				summaries := tracker.GetActivitySummaries()
				submitActivitiesToRescueTime(apiKey, summaries)
			}

			// Print summary before exit (user-facing)
			printActivitySummary(tracker)
			return

		case <-submitChan:
			slog.Debug("submission tick triggered")
			summaries := tracker.GetActivitySummaries()
			submitActivitiesToRescueTime(apiKey, summaries)

			// Clear completed sessions after successful submission
			tracker.ClearCompletedSessions()

		case <-pollTicker.C:
			window, err := getActiveWindow()
			if err != nil {
				slog.Debug("failed to get active window", "error", err)
				continue
			}

			// Check if the application or window title changed
			if window.Class != lastAppClass || window.Title != lastWindowTitle {
				slog.Debug("window changed",
					"from_app", lastAppClass, "to_app", window.Class,
					"title", window.Title)

				// Start a new session for the new window/app
				tracker.StartSession(window.Class, window.Title)

				// Print the change (user-facing)
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
	configPath := flag.String("config", "", "Config file path (default: ~/.config/rescuetime-linux/config.json)")

	// These flags override config file values
	interval := flag.String("interval", "", "Polling interval (e.g., 100ms, 1s)")
	submissionInterval := flag.String("submission-interval", "", "Submission interval (e.g., 15m, 1h)")
	logLevel := flag.String("log-level", "", "Log level: debug, info, warn, error")
	logFile := flag.String("log-file", "", "Log file path")
	flag.Parse()

	// Load configuration: defaults < config file < CLI flags
	cfg, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Apply CLI flag overrides (only flags that were explicitly set)
	cliFlags := make(map[string]string)
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "interval":
			cliFlags["interval"] = *interval
		case "submission-interval":
			cliFlags["submission-interval"] = *submissionInterval
		case "log-level":
			cliFlags["log-level"] = *logLevel
		case "log-file":
			cliFlags["log-file"] = *logFile
		}
	})
	cfg.ApplyCLIFlags(cliFlags)

	// Set up structured logging (using final resolved config)
	level := parseLogLevel(cfg.LogLevel)
	cleanupLogger := setupLogger(level, cfg.LogFile)
	defer cleanupLogger()

	slog.Debug("configuration loaded",
		"polling_interval", cfg.PollingInterval.Duration,
		"submission_interval", cfg.SubmissionInterval.Duration,
		"merge_threshold", cfg.MergeThreshold.Duration,
		"min_duration", cfg.MinDuration.Duration,
		"log_level", cfg.LogLevel)

	// Check if we're running in a graphical environment (Wayland or X11)
	if os.Getenv("WAYLAND_DISPLAY") == "" && os.Getenv("DISPLAY") == "" {
		slog.Error("no graphical display found",
			"hint", "Make sure you're running this in a Wayland or X11 environment")
		os.Exit(1)
	}

	// Check if hyprctl is available (required for Wayland/Hyprland)
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		_, err := exec.LookPath("hyprctl")
		if err != nil {
			slog.Error("hyprctl not found",
				"hint", "This application requires Hyprland on Wayland")
			os.Exit(1)
		}
	}

	if *monitor || *track {
		if *track {
			fmt.Printf("Tracking application usage (polling every %v). Press Ctrl+C to stop and see summary.\n",
				cfg.PollingInterval.Duration)
		} else {
			fmt.Printf("Monitoring window changes (polling every %v). Press Ctrl+C to stop.\n",
				cfg.PollingInterval.Duration)
		}

		// Handle API submission setup
		var apiKey string
		if *submit {
			// Load environment variables from .env file
			err := loadEnvFile(cfg.EnvFilePath)
			if err != nil {
				slog.Error("failed to load .env file", "path", cfg.EnvFilePath, "error", err)
				os.Exit(1)
			}

			// Get API key from environment
			apiKey = os.Getenv("RESCUE_TIME_API_KEY")
			if apiKey == "" {
				slog.Error("RESCUE_TIME_API_KEY not found in environment")
				os.Exit(1)
			}

			monitorWindowChanges(cfg, true, apiKey)
		} else {
			monitorWindowChanges(cfg, false, "")
		}
	} else {
		// Single execution mode
		currentInfo, err := getCurrentWindowInfo()
		if err != nil {
			slog.Error("failed to get window info", "error", err)
			os.Exit(1)
		}
		fmt.Println(currentInfo)
	}
}
