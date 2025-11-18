package utils

import (
	"strings"
)

// DiffStat represents statistics about a diff.
type DiffStat struct {
	ModelAddedLines    int
	ModelRemovedLines  int
	ModelAddedChars    int
	ModelRemovedChars  int
	UserAddedLines     int
	UserRemovedLines   int
	UserAddedChars     int
	UserRemovedChars   int
}

// GetDiffStat calculates statistics about changes between different versions of a file.
func GetDiffStat(
	oldStr string,
	aiStr string,
	userStr string,
) DiffStat {
	// For now, a very basic implementation.
	// A full diffing library would be needed for accurate line/char counts.

	oldLines := strings.Split(oldStr, "\n")
	aiLines := strings.Split(aiStr, "\n")
	userLines := strings.Split(userStr, "\n")

	modelAddedLines := len(aiLines) - len(oldLines)
	modelRemovedLines := len(oldLines) - len(aiLines)
	modelAddedChars := len(aiStr) - len(oldStr)
	modelRemovedChars := len(oldStr) - len(aiStr)

	userAddedLines := len(userLines) - len(aiLines)
	userRemovedLines := len(aiLines) - len(userLines)
	userAddedChars := len(userStr) - len(aiStr)
	userRemovedChars := len(aiStr) - len(userStr)

	return DiffStat{
		ModelAddedLines:    max(0, modelAddedLines),
		ModelRemovedLines:  max(0, modelRemovedLines),
		ModelAddedChars:    max(0, modelAddedChars),
		ModelRemovedChars:  max(0, modelRemovedChars),
		UserAddedLines:     max(0, userAddedLines),
		UserRemovedLines:   max(0, userRemovedLines),
		UserAddedChars:     max(0, userAddedChars),
		UserRemovedChars:   max(0, userRemovedChars),
	}
}

// max returns the greater of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
