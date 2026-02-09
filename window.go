package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ActiveWindow represents the active window's essential properties.
type ActiveWindow struct {
	Class string
	Title string
}

// WindowBackend abstracts active window detection across compositors.
type WindowBackend interface {
	GetActiveWindow() (*ActiveWindow, error)
	Name() string
}

// HyprlandBackend gets the active window via hyprctl (Hyprland compositor).
type HyprlandBackend struct{}

type hyprctlWindow struct {
	Class string `json:"class"`
	Title string `json:"title"`
}

func (b *HyprlandBackend) GetActiveWindow() (*ActiveWindow, error) {
	cmd := exec.Command("hyprctl", "activewindow", "-j")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get active window from hyprctl: %v", err)
	}

	var window hyprctlWindow
	err = json.Unmarshal(output, &window)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hyprctl JSON output: %v", err)
	}

	return &ActiveWindow{Class: window.Class, Title: window.Title}, nil
}

func (b *HyprlandBackend) Name() string { return "hyprland" }

// XdotoolBackend gets the active window via xdotool (X11/XWayland).
type XdotoolBackend struct{}

func (b *XdotoolBackend) GetActiveWindow() (*ActiveWindow, error) {
	// Get window ID first to avoid race between class/title lookups
	idOut, err := exec.Command("xdotool", "getactivewindow").Output()
	if err != nil {
		return nil, fmt.Errorf("xdotool getactivewindow failed: %v", err)
	}
	windowID := strings.TrimSpace(string(idOut))

	classOut, err := exec.Command("xdotool", "getwindowclassname", windowID).Output()
	if err != nil {
		return nil, fmt.Errorf("xdotool getwindowclassname failed: %v", err)
	}

	titleOut, err := exec.Command("xdotool", "getwindowname", windowID).Output()
	if err != nil {
		return nil, fmt.Errorf("xdotool getwindowname failed: %v", err)
	}

	return &ActiveWindow{
		Class: strings.TrimSpace(string(classOut)),
		Title: strings.TrimSpace(string(titleOut)),
	}, nil
}

func (b *XdotoolBackend) Name() string { return "xdotool" }

// DetectBackend auto-detects the best available window backend.
// Preference order: hyprctl (Hyprland) > xdotool (X11/XWayland).
func DetectBackend() (WindowBackend, error) {
	if _, err := exec.LookPath("hyprctl"); err == nil {
		return &HyprlandBackend{}, nil
	}
	if _, err := exec.LookPath("xdotool"); err == nil {
		return &XdotoolBackend{}, nil
	}
	return nil, fmt.Errorf("no window backend found: install hyprctl (Hyprland) or xdotool (X11/XWayland)")
}
