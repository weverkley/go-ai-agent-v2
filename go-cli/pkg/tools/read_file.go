package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFileTool represents the read-file tool.
type ReadFileTool struct {
}

// NewReadFileTool creates a new instance of ReadFileTool.
func NewReadFileTool() *ReadFileTool {
	return &ReadFileTool{}
}

// Execute performs a read-file operation.
func (t *ReadFileTool) Execute(
	absolutePath string,
	offset int,
	limit int,
) (string, error) {
	// Check if file exists
	info, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", absolutePath)
		}
		return "", fmt.Errorf("failed to get file info for %s: %w", absolutePath, err)
	}

	// Check if it's a directory
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file: %s", absolutePath)
	}

	// For now, assume all are text files.
	// TODO: Implement image/pdf handling.
	// If it's not a text file, return a placeholder message. 
	// For simplicity, we'll just check extension for now.
	ext := strings.ToLower(filepath.Ext(absolutePath))
	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".bmp" || ext == ".pdf" {
		return fmt.Sprintf("Content of %s (binary file, not displayed)", absolutePath), nil
	}

	file, err := os.Open(absolutePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", absolutePath, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file %s: %w", absolutePath, err)
	}

	originalLineCount := len(lines)
	isTruncated := false
	linesShownStart := 0
	linesShownEnd := originalLineCount

	if offset > 0 {
		linesShownStart = offset
	}
	if limit > 0 {
		linesShownEnd = linesShownStart + limit
	}

	if linesShownStart >= originalLineCount {
		return fmt.Sprintf("Offset %d is beyond the end of the file (total lines: %d)", offset, originalLineCount), nil
	}

	if linesShownEnd > originalLineCount {
		linesShownEnd = originalLineCount
	}

	if linesShownStart > 0 || linesShownEnd < originalLineCount {
		isTruncated = true
	}

	var contentBuilder strings.Builder
	for i := linesShownStart; i < linesShownEnd; i++ {
		contentBuilder.WriteString(lines[i])
		contentBuilder.WriteString("\n")
	}

	var llmContent strings.Builder
	if isTruncated {
		llmContent.WriteString("\nIMPORTANT: The file content has been truncated.\n")
		llmContent.WriteString(fmt.Sprintf("Status: Showing lines %d-%d of %d total lines.\n", linesShownStart+1, linesShownEnd, originalLineCount))
		llmContent.WriteString(fmt.Sprintf("Action: To read more of the file, you can use the 'offset' and 'limit' parameters in a subsequent 'read_file' call. For example, to read the next section of the file, use offset: %d.\n", linesShownEnd+1))
		llmContent.WriteString("\n--- FILE CONTENT (truncated) ---\n")
	}
	llmContent.WriteString(contentBuilder.String())

	return llmContent.String(), nil
}
