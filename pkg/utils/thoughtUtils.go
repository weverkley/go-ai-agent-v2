package utils

import (
	"strings"
)

// ThoughtSummary represents a structured thought object.
type ThoughtSummary struct {
	Subject     string
	Description string
}

const (
	startDelimiter = "**"
	endDelimiter   = "**"
)

// ParseThought parses a raw thought string into a structured ThoughtSummary object.
// Thoughts are expected to have a bold "subject" part enclosed in double
// asterisks (e.g., **Subject**). The rest of the string is considered
// the description. This function only parses the first valid subject found.
func ParseThought(rawText string) ThoughtSummary {
	startIndex := strings.Index(rawText, startDelimiter)
	if startIndex == -1 {
		// No start delimiter found, the whole text is the description.
		return ThoughtSummary{Subject: "", Description: strings.TrimSpace(rawText)}
	}

	endIndex := strings.Index(rawText[startIndex+len(startDelimiter):], endDelimiter)
	if endIndex == -1 {
		// Start delimiter found but no end delimiter, so it's not a valid subject.
		// Treat the entire string as the description.
		return ThoughtSummary{Subject: "", Description: strings.TrimSpace(rawText)}
	}

	endIndex += startIndex + len(startDelimiter) // Adjust endIndex to be absolute

	subject := strings.TrimSpace(rawText[startIndex+len(startDelimiter) : endIndex])

	// The description is everything before the start delimiter and after the end delimiter.
	description := strings.TrimSpace(rawText[0:startIndex] + rawText[endIndex+len(endDelimiter):])

	return ThoughtSummary{Subject: subject, Description: description}
}
