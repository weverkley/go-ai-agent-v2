package tools

import (
	"bufio"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/services"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GrepTool represents the grep tool.
type GrepTool struct {
	fsService *services.FileSystemService
}

// NewGrepTool creates a new instance of GrepTool.
func NewGrepTool() *GrepTool {
	return &GrepTool{
		fsService: services.NewFileSystemService(),
	}
}

// GrepMatch represents a single grep match.
type GrepMatch struct {
	FilePath   string
	LineNumber int
	Line       string
}

// Execute performs a grep search.
func (t *GrepTool) Execute(
	pattern string,
	searchPath string,
	include string,
) (string, error) {
	// Compile the regex pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Resolve the search path
	absSearchPath, err := filepath.Abs(searchPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for %s: %w", searchPath, err)
	}

	var allMatches []GrepMatch
	matchesByFile := make(map[string][]GrepMatch)

	err = filepath.Walk(absSearchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Filter by include glob pattern if provided
		if include != "" {
			match, err := filepath.Match(include, info.Name())
			if err != nil {
				return fmt.Errorf("invalid include pattern: %w", err)
			}
			if !match {
				return nil // Skip file if it doesn't match include pattern
			}
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNumber := 1
		for scanner.Scan() {
			line := scanner.Text()
			if re.MatchString(line) {
				match := GrepMatch{
					FilePath:   path,
					LineNumber: lineNumber,
					Line:       line,
				}
				allMatches = append(allMatches, match)
				matchesByFile[path] = append(matchesByFile[path], match)
			}
			lineNumber++
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error walking the path %s: %w", absSearchPath, err)
	}

	if len(allMatches) == 0 {
		return fmt.Sprintf("No matches found for pattern \"%s\" in path \"%s\"", pattern, searchPath), nil
	}

	var llmContent strings.Builder
	llmContent.WriteString(fmt.Sprintf("Found %d matches for pattern \"%s\" in path \"%s\":\n---\n", len(allMatches), pattern, searchPath))

	for filePath, matches := range matchesByFile {
		llmContent.WriteString(fmt.Sprintf("File: %s\n", filePath))
		for _, match := range matches {
			llmContent.WriteString(fmt.Sprintf("L%d: %s\n", match.LineNumber, strings.TrimSpace(match.Line)))
		}
		llmContent.WriteString("---\n\n")
	}

	return llmContent.String(), nil
}
