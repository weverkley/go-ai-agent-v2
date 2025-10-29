package tools

import (
	"bufio"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/services"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gobwas/glob"
)

// GlobTool represents the glob tool.
type GlobTool struct {
	fsService  *services.FileSystemService
	gitService *services.GitService // Add GitService
}

// NewGlobTool creates a new instance of GlobTool.
func NewGlobTool() *GlobTool {
	return &GlobTool{
		fsService:  services.NewFileSystemService(),
		gitService: services.NewGitService(), // Initialize GitService
	}
}

// FileInfo represents a file found by glob.
type FileInfo struct {
	Path    string
	ModTime time.Time
}

// getIgnorePatterns reads .gitignore and .geminiignore files and returns a list of glob patterns.
func (t *GlobTool) getIgnorePatterns(searchDir string, respectGitIgnore, respectGeminiIgnore bool) ([]glob.Glob, error) {
	var ignorePatterns []glob.Glob

	// Read .gitignore
	if respectGitIgnore {
		gitIgnorePath := filepath.Join(searchDir, ".gitignore")
		if _, err := os.Stat(gitIgnorePath); err == nil {
			patterns, err := t.readIgnoreFile(gitIgnorePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read .gitignore: %w", err)
			}
			ignorePatterns = append(ignorePatterns, patterns...)
		}
	}

	// Read .geminiignore
	if respectGeminiIgnore {
		geminiIgnorePath := filepath.Join(searchDir, ".geminiignore")
		if _, err := os.Stat(geminiIgnorePath); err == nil {
			patterns, err := t.readIgnoreFile(geminiIgnorePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read .geminiignore: %w", err)
			}
			ignorePatterns = append(ignorePatterns, patterns...)
		}
	}

	return ignorePatterns, nil
}

// readIgnoreFile reads an ignore file and compiles its patterns.
func (t *GlobTool) readIgnoreFile(filePath string) ([]glob.Glob, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []glob.Glob
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}
		g, err := glob.Compile(line)
		if err != nil {
			return nil, fmt.Errorf("failed to compile ignore pattern %s from %s: %w", line, filePath, err)
		}
		patterns = append(patterns, g)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return patterns, nil
}

// Execute performs a glob search.
func (t *GlobTool) Execute(
	pattern string,
	searchPath string,
	caseSensitive bool,
	respectGitIgnore bool,
	respectGeminiIgnore bool,
) (string, error) {
	// Resolve the search path
	absSearchPath, err := filepath.Abs(searchPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for %s: %w", searchPath, err)
	}

	// Compile the glob pattern
	compiledPattern := pattern
	if !caseSensitive {
		compiledPattern = strings.ToLower(pattern)
	}
	g, err := glob.Compile(compiledPattern)
	if err != nil {
		return "", fmt.Errorf("failed to compile glob pattern %s: %w", pattern, err)
	}

	// Get ignore patterns
	ignorePatterns, err := t.getIgnorePatterns(absSearchPath, respectGitIgnore, respectGeminiIgnore)
	if err != nil {
		return "", fmt.Errorf("failed to get ignore patterns: %w", err)
	}

	var matchedFiles []FileInfo

	err = filepath.Walk(absSearchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path from the search path
		relPath, err := filepath.Rel(absSearchPath, path)
		if err != nil {
			return err
		}

		// Apply case sensitivity for matching
		matchPath := relPath
		if !caseSensitive {
			matchPath = strings.ToLower(relPath)
		}

		// Check against ignore patterns
		for _, ignoreG := range ignorePatterns {
			if ignoreG.Match(relPath) { // Ignore patterns are usually case-sensitive by default, but glob library handles it.
				return nil // Skip this file
			}
		}

		if g.Match(matchPath) {
			matchedFiles = append(matchedFiles, FileInfo{
				Path:    path,
				ModTime: info.ModTime(),
			})
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error walking the path %s: %w", absSearchPath, err)
	}

	// Sort files by modification time (newest first)
	sort.Slice(matchedFiles, func(i, j int) bool {
		return matchedFiles[i].ModTime.After(matchedFiles[j].ModTime)
	})

	var resultPaths []string
	for _, file := range matchedFiles {
		resultPaths = append(resultPaths, file.Path)
	}

	if len(resultPaths) == 0 {
		return fmt.Sprintf("No files found matching pattern \"%s\" in path \"%s\"", pattern, searchPath), nil
	}

	resultMessage := fmt.Sprintf("Found %d file(s) matching \"%s\" in path \"%s\":\n%s",
		len(resultPaths), pattern, searchPath, strings.Join(resultPaths, "\n"))

	return resultMessage, nil
}
