package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

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

// printLiveStats prints a compact top-5 apps by duration
func printLiveStats(tracker *ActivityTracker, startTime time.Time) {
	summaries := tracker.GetActivitySummaries()
	if len(summaries) == 0 {
		return
	}

	// Sort by duration (top 5)
	type entry struct {
		app      string
		duration time.Duration
	}
	var sorted []entry
	for app, s := range summaries {
		sorted = append(sorted, entry{app, s.TotalDuration})
	}
	// Simple insertion sort — at most ~20 apps
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j].duration > sorted[j-1].duration; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	if len(sorted) > 5 {
		sorted = sorted[:5]
	}

	elapsed := time.Since(startTime).Round(time.Second)
	fmt.Printf("\n--- Stats (uptime %v) ---\n", elapsed)
	for _, e := range sorted {
		fmt.Printf("  %s: %v\n", e.app, e.duration.Round(time.Second))
	}
	fmt.Println("---")
}

func getCurrentWindowInfo(backend WindowBackend) (string, error) {
	window, err := backend.GetActiveWindow()
	if err != nil {
		return "", err
	}
	return formatWindowOutput(window.Title, window.Class), nil
}

func monitorWindowChanges(backend WindowBackend, cfg Config, submitToAPI bool, apiKey string, showStats bool, statsInterval time.Duration) {
	var lastAppClass, lastWindowTitle string

	// Create activity tracker and title cleaner
	tracker := NewActivityTracker(cfg.MergeThreshold.Duration, cfg.MinDuration.Duration)
	cleaner := NewTitleCleaner()

	// Set up persistence
	store, err := NewPersistenceStore(cfg.PersistencePath)
	if err != nil {
		slog.Warn("persistence disabled", "error", err)
	} else {
		// Load saved sessions from previous run
		saved, err := store.LoadSessions()
		if err != nil {
			slog.Warn("failed to load saved sessions", "error", err)
		} else if len(saved) > 0 {
			tracker.LoadSessions(saved)
		}
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Get initial window info and start the first session
	window, err := backend.GetActiveWindow()
	if err != nil {
		slog.Error("failed to get initial window info", "error", err)
		return
	}

	// Start the initial session with cleaned title
	cleanedTitle := cleaner.Clean(window.Class, window.Title)
	tracker.StartSession(window.Class, cleanedTitle)
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

	var statsTicker *time.Ticker
	var statsChan <-chan time.Time
	startTime := time.Now()

	if showStats {
		statsTicker = time.NewTicker(statsInterval)
		defer statsTicker.Stop()
		statsChan = statsTicker.C
		slog.Info("live stats enabled", "interval", statsInterval)
	}

	for {
		select {
		case <-sigChan:
			slog.Info("shutting down window monitor")

			// End the current session
			tracker.EndCurrentSession()

			if submitToAPI {
				// Submit final data then clear persistence
				summaries := tracker.GetActivitySummaries()
				submitActivitiesToRescueTime(apiKey, summaries)
				if store != nil {
					store.Clear()
				}
			} else if store != nil {
				// Save sessions for next run
				if err := store.SaveSessions(tracker.GetSessions()); err != nil {
					slog.Error("failed to save sessions on shutdown", "error", err)
				}
			}

			// Print summary before exit (user-facing)
			printActivitySummary(tracker)
			return

		case <-submitChan:
			slog.Debug("submission tick triggered")
			summaries := tracker.GetActivitySummaries()
			submitActivitiesToRescueTime(apiKey, summaries)

			// Clear completed sessions after submission
			tracker.ClearCompletedSessions()

			// Save any remaining sessions to disk
			if store != nil {
				remaining := tracker.GetSessions()
				if len(remaining) > 0 {
					if err := store.SaveSessions(remaining); err != nil {
						slog.Warn("failed to save remaining sessions", "error", err)
					}
				} else {
					store.Clear()
				}
			}

		case <-statsChan:
			printLiveStats(tracker, startTime)

		case <-pollTicker.C:
			window, err := backend.GetActiveWindow()
			if err != nil {
				slog.Debug("failed to get active window", "error", err)
				continue
			}

			// Check if the application or window title changed
			if window.Class != lastAppClass || window.Title != lastWindowTitle {
				slog.Debug("window changed",
					"from_app", lastAppClass, "to_app", window.Class,
					"title", window.Title)

				// Start a new session with cleaned title
				cleanedTitle := cleaner.Clean(window.Class, window.Title)
				tracker.StartSession(window.Class, cleanedTitle)

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

	// Stats flags
	stats := flag.Bool("stats", false, "Show periodic activity statistics")
	statsInterval := flag.Duration("stats-interval", 60*time.Second, "Interval for live stats display (e.g., 10s, 60s)")

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

	// Auto-detect window backend (hyprctl or xdotool)
	backend, err := DetectBackend()
	if err != nil {
		slog.Error("no window backend available", "error", err)
		os.Exit(1)
	}
	slog.Info("detected window backend", "backend", backend.Name())

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

			monitorWindowChanges(backend, cfg, true, apiKey, *stats, *statsInterval)
		} else {
			monitorWindowChanges(backend, cfg, false, "", *stats, *statsInterval)
		}
	} else {
		// Single execution mode
		currentInfo, err := getCurrentWindowInfo(backend)
		if err != nil {
			slog.Error("failed to get window info", "error", err)
			os.Exit(1)
		}
		fmt.Println(currentInfo)
	}
}
