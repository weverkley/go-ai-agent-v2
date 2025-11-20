package pathutils

import (
	"fmt"
	"os"
	"strings"
)

// ExpandPath expands the tilde (~) in a path to the user's home directory.
func ExpandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	if path == "~" {
		return homeDir, nil
	}

	if strings.HasPrefix(path, "~/") {
		return strings.Replace(path, "~", homeDir, 1), nil
	}

	return path, nil
}
