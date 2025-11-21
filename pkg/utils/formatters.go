package utils

import (
	"fmt"
	"regexp"
	"time"
)

// FormatDuration formats a duration into a human-readable string.
func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

// StripAnsi removes ANSI escape codes from a string.
func StripAnsi(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}