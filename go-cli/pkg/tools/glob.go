package tools

import (
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
	fsService *services.FileSystemService
}

// NewGlobTool creates a new instance of GlobTool.
func NewGlobTool() *GlobTool {
	return &GlobTool{
		fsService: services.NewFileSystemService(),
	}
}

// FileInfo represents a file found by glob.
type FileInfo struct {
	Path    string
	ModTime time.Time
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

	var matchedFiles []FileInfo

	err = filepath.Walk(absSearchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories for now, we are looking for files
		if info.IsDir() {
			return nil
		}

		// Get relative path from the search path
		relPath, err := filepath.Rel(absSearchPath, path)
		if err != nil {
			return err
		}

		matchPath := relPath
		if !caseSensitive {
			matchPath = strings.ToLower(relPath)
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
