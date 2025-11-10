package services

import (
	"bufio"
	"fmt" // Add fmt import
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/gobwas/glob"
)

// FileFilteringService implements the FileService interface for filtering files based on ignore patterns.
type FileFilteringService struct {
	projectRoot string
	gitIgnorePatterns []glob.Glob
	geminiIgnorePatterns []glob.Glob
	gitIgnoreStringPatterns []string // New field
	geminiIgnoreStringPatterns []string // New field
	mu sync.RWMutex
}

// NewFileFilteringService creates a new instance of FileFilteringService.
func NewFileFilteringService(projectRoot string) (*FileFilteringService, error) {
	fs := &FileFilteringService{
		projectRoot: projectRoot,
	}
	if err := fs.loadIgnorePatterns(); err != nil {
		return nil, err
	}
	return fs, nil
}

// loadIgnorePatterns loads .gitignore and .geminiignore patterns.
func (fs *FileFilteringService) loadIgnorePatterns() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	gitIgnorePath := filepath.Join(fs.projectRoot, ".gitignore")
	geminiIgnorePath := filepath.Join(fs.projectRoot, ".geminiignore")

	var err error
	fs.gitIgnorePatterns, fs.gitIgnoreStringPatterns, err = fs.parseIgnoreFile(gitIgnorePath)
	if err != nil {
		// Log error but continue, .gitignore is optional
		fmt.Printf("Warning: failed to load .gitignore: %v\n", err)
	}

	fs.geminiIgnorePatterns, fs.geminiIgnoreStringPatterns, err = fs.parseIgnoreFile(geminiIgnorePath)
	if err != nil {
		// Log error but continue, .geminiignore is optional
		fmt.Printf("Warning: failed to load .geminiignore: %v\n", err)
	}
	return nil
}

// parseIgnoreFile reads an ignore file and compiles glob patterns.
func (fs *FileFilteringService) parseIgnoreFile(filePath string) ([]glob.Glob, []string, error) { // Modified return signature
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil // File does not exist, no patterns
		}
		return nil, nil, fmt.Errorf("failed to open ignore file %s: %w", filePath, err)
	}
	defer file.Close()

	var globPatterns []glob.Glob
	var stringPatterns []string // New slice to store string patterns
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		// Handle negation patterns (e.g., !file.txt)
		isNegated := false
		if strings.HasPrefix(line, "!") {
			isNegated = true
			line = strings.TrimPrefix(line, "!")
		}

		// Convert gitignore patterns to glob patterns
		// A leading slash means the pattern is relative to the root of the git repo
		// No leading slash means the pattern can match in any directory
		processedLine := line // Store the processed line for string patterns
		if !strings.HasPrefix(line, "/") {
			line = "**/" + line
		} else {
			line = strings.TrimPrefix(line, "/")
		}

		g, err := glob.Compile(line)
		if err != nil {
			fmt.Printf("Warning: failed to compile glob pattern '%s' from %s: %v\n", line, filePath, err)
			continue
		}
		if isNegated {
			// Store negated patterns separately or handle during matching
			// For simplicity, we'll just compile it and handle negation during ShouldIgnoreFile
			// A more robust solution might store negated patterns in a separate list
		}
		globPatterns = append(globPatterns, g)
		stringPatterns = append(stringPatterns, processedLine) // Store the processed string pattern
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to read ignore file %s: %w", filePath, err)
	}
	return globPatterns, stringPatterns, nil // Modified return
}

// ShouldIgnoreFile checks if a file should be ignored based on filtering options.
func (fs *FileFilteringService) ShouldIgnoreFile(filePath string, options types.FileFilteringOptions) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	relativePath, err := filepath.Rel(fs.projectRoot, filePath)
	if err != nil {
		// If path is not relative to project root, can't apply ignore rules
		return false
	}

	// Normalize path for glob matching (use forward slashes)
	normalizedPath := filepath.ToSlash(relativePath)

	// Check .gitignore patterns
	if options.RespectGitIgnore != nil && *options.RespectGitIgnore {
		for _, pattern := range fs.gitIgnorePatterns {
			if pattern.Match(normalizedPath) {
				return true
			}
		}
	}

	// Check .geminiignore patterns
	if options.RespectGeminiIgnore != nil && *options.RespectGeminiIgnore {
		for _, pattern := range fs.geminiIgnorePatterns {
			if pattern.Match(normalizedPath) {
				return true
			}
		}
	}

	return false
}

// GetIgnoredPatterns returns a list of all loaded ignore patterns.
func (fs *FileFilteringService) GetIgnoredPatterns() []string {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var patterns []string
	patterns = append(patterns, fs.gitIgnoreStringPatterns...)
	patterns = append(patterns, fs.geminiIgnoreStringPatterns...)
	return patterns
}
