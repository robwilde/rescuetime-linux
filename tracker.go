package main

import (
	"sync"
	"time"
)

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

// NewActivityTracker creates a new activity tracker with the given configuration
func NewActivityTracker(mergeThreshold, minDuration time.Duration) *ActivityTracker {
	return &ActivityTracker{
		sessions:       make([]ActivitySession, 0),
		mergeThreshold: mergeThreshold,
		minDuration:    minDuration,
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
