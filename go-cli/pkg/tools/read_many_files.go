package tools

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"go-ai-agent-v2/go-cli/pkg/services"
)

// ReadManyFilesTool represents the read-many-files tool.
type ReadManyFilesTool struct {
	fsService *services.FileSystemService
	gitService *services.GitService
}

// NewReadManyFilesTool creates a new instance of ReadManyFilesTool.
func NewReadManyFilesTool() *ReadManyFilesTool {
	return &ReadManyFilesTool{
		fsService: services.NewFileSystemService(),
				gitService: services.NewGitService(),
			}
		}

// SkippedFile represents a file that was skipped during processing.
type SkippedFile struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// ReadManyFilesMetadata represents metadata about the files processed.
type ReadManyFilesMetadata struct {
	ProcessedFiles []string      `json:"processedFiles"`
	SkippedFiles   []SkippedFile `json:"skippedFiles"`
}

// ReadManyFilesResult represents the structure of the read_many_files tool output.
type ReadManyFilesResult struct {
	Content  string                `json:"content"`
	Metadata ReadManyFilesMetadata `json:"metadata"`
}

// Execute performs a read-many-files operation.
func (t *ReadManyFilesTool) Execute(
	patterns []string,
	includePatterns []string,
	excludePatterns []string,
	recursive bool,
	useDefaultExcludes bool,
	respectGitIgnore bool,
	respectGeminiIgnore bool,
) (string, error) {
	var allFiles []string
	var processedFiles []string
	var skippedFilesList []SkippedFile // Renamed to avoid conflict with local var
	var contentBuilder strings.Builder

	// Combine all patterns for glob search
	searchPatterns := append(patterns, includePatterns...)

	// Default excludes (simplified for now)
	if useDefaultExcludes {
		excludePatterns = append(excludePatterns, "node_modules", ".git", ".gemini")
	}

	// Compile exclude patterns
	var compiledExcludeGlobs []glob.Glob
	for _, p := range excludePatterns {
		g, err := glob.Compile(p)
		if err != nil {
			return "", fmt.Errorf("failed to compile exclude pattern %s: %w", p, err)
		}
		compiledExcludeGlobs = append(compiledExcludeGlobs, g)
	}

	// Walk through each search pattern
	for _, searchPattern := range searchPatterns {
		absSearchPath, err := filepath.Abs(".") // Start from current directory
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path: %w", err)
		}

		err = filepath.Walk(absSearchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				// If not recursive and it's a directory, skip
				if !recursive && path != absSearchPath {
					return filepath.SkipDir
				}
				// Check if directory is excluded
				for _, excludeG := range compiledExcludeGlobs {
					relPath, _ := filepath.Rel(absSearchPath, path)
					if excludeG.Match(relPath) {
						return filepath.SkipDir
					}
				}
				return nil
			}

			relPath, err := filepath.Rel(absSearchPath, path)
			if err != nil {
				return err
			}

			// Check against exclude patterns
			for _, excludeG := range compiledExcludeGlobs {
				if excludeG.Match(relPath) {
					return nil // Skip this file
				}
			}

			// Match against the current search pattern
			g, err := glob.Compile(searchPattern)
			if err != nil {
				return fmt.Errorf("failed to compile search pattern %s: %w", searchPattern, err)
			}

			if g.Match(relPath) {
				allFiles = append(allFiles, path)
			}
			return nil
		})

		if err != nil {
			return "", fmt.Errorf("error walking path for pattern %s: %w", searchPattern, err)
		}
	}

	// Process unique files
	uniqueFiles := make(map[string]bool)
	for _, file := range allFiles {
		if _, exists := uniqueFiles[file]; !exists {
			uniqueFiles[file] = true

			// Read file content
			content, err := ioutil.ReadFile(file)
			if err != nil {
				skippedFilesList = append(skippedFilesList, SkippedFile{Path: file, Reason: fmt.Sprintf("failed to read: %v", err)})
				continue
			}

			// For now, assume all are text files.
			// TODO: Implement image/pdf handling.
			contentBuilder.WriteString(fmt.Sprintf("--- %s ---\n", file))
			contentBuilder.WriteString(string(content))
			contentBuilder.WriteString("\n")
			processedFiles = append(processedFiles, file)
		}
	}

	var displayMessage strings.Builder
	displayMessage.WriteString("### ReadManyFiles Result\n\n")

	if len(processedFiles) > 0 {
		displayMessage.WriteString(fmt.Sprintf("Successfully read and concatenated content from **%d file(s)**.\n", len(processedFiles)))
		displayMessage.WriteString("\n**Processed Files:**\n")
		for _, p := range processedFiles {
			displayMessage.WriteString(fmt.Sprintf("- `%s`\n", p))
		}
	}

	if len(skippedFilesList) > 0 {
		displayMessage.WriteString(fmt.Sprintf("\n**Skipped %d item(s):**\n", len(skippedFilesList)))
		for _, f := range skippedFilesList {
			displayMessage.WriteString(fmt.Sprintf("- `%s` (Reason: %s)\n", f.Path, f.Reason))
		}
	}

	if len(processedFiles) == 0 && len(skippedFilesList) == 0 {
		displayMessage.WriteString("No files were read and concatenated based on the criteria.\n")
	}

	return contentBuilder.String() + "\n\n" + displayMessage.String(), nil
}
