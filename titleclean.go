package main

import (
	"regexp"
	"strings"
)

// TitleCleaner applies rule-based cleaning to window titles before storage
type TitleCleaner struct {
	rules []cleanRule
}

type cleanRule struct {
	name    string
	pattern *regexp.Regexp
	replace string
}

// NewTitleCleaner creates a new title cleaner with default rules
func NewTitleCleaner() *TitleCleaner {
	tc := &TitleCleaner{}

	// Strip unread/notification counts: "(1,064)", "(3)", "(12)"
	tc.rules = append(tc.rules, cleanRule{
		name:    "unread_counts",
		pattern: regexp.MustCompile(`\(\d{1,3}(?:,\d{3})*\)\s*`),
		replace: "",
	})

	// Strip trailing browser app names
	tc.rules = append(tc.rules, cleanRule{
		name:    "trailing_browser",
		pattern: regexp.MustCompile(`\s*[-–—]\s*(?:Mozilla Firefox|Firefox|Google Chrome|Chromium|Wavebox|Brave|Microsoft Edge|Opera|Vivaldi|Safari)\s*$`),
		replace: "",
	})

	// Normalize whitespace: multiple spaces to single
	tc.rules = append(tc.rules, cleanRule{
		name:    "normalize_whitespace",
		pattern: regexp.MustCompile(`\s{2,}`),
		replace: " ",
	})

	return tc
}

// Clean applies all rules to a window title and returns the cleaned result
func (tc *TitleCleaner) Clean(appClass, title string) string {
	cleaned := title

	for _, rule := range tc.rules {
		cleaned = rule.pattern.ReplaceAllString(cleaned, rule.replace)
	}

	// Trim leading/trailing whitespace
	cleaned = strings.TrimSpace(cleaned)

	// Truncate to 255 chars (RescueTime API limit)
	if len(cleaned) > 255 {
		cleaned = cleaned[:255]
	}

	return cleaned
}
