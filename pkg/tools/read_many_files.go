package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/gobwas/glob"
)

// ReadManyFilesTool represents the read-many-files tool.
type ReadManyFilesTool struct {
	*types.BaseDeclarativeTool
	fileSystemService services.FileSystemService
}

// NewReadManyFilesTool creates a new instance of ReadManyFilesTool.
func NewReadManyFilesTool(fs services.FileSystemService) *ReadManyFilesTool {
	return &ReadManyFilesTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			"read_many_files",
			"read_many_files",
			"Reads content from multiple files specified by paths or glob patterns (e.g., `src/**/*.ts`, `**/*.md`), returning absolute paths sorted by modification time (newest first). Ideal for quickly locating files based on their name or path structure, especially in large codebases.",
			types.KindOther, // Assuming KindOther for now
			&types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]*types.JsonSchemaProperty{
					"paths": &types.JsonSchemaProperty{
						Type:        "array",
						Description: "Required. An array of glob patterns or paths relative to the tool's target directory. Examples: ['src/**/*.ts'], ['README.md', 'docs/']",
						Items:       &types.JsonSchemaObject{Type: "string"},
					},
					"include": &types.JsonSchemaProperty{
						Type:        "array",
						Description: "Optional. Additional glob patterns to include. These are merged with `paths`. Example: \"*.test.ts\" to specifically add test files if they were broadly excluded.",
						Items:       &types.JsonSchemaObject{Type: "string"},
					},
					"exclude": &types.JsonSchemaProperty{
						Type:        "array",
						Description: "Optional. Glob patterns for files/directories to exclude. Added to default excludes if useDefaultExcludes is true. Example: \"**/*.log\", \"temp/\"",
						Items:       &types.JsonSchemaObject{Type: "string"},
					},
					"recursive": &types.JsonSchemaProperty{
						Type:        "boolean",
						Description: "Optional. Whether to search recursively (primarily controlled by `**` in glob patterns). Defaults to true.",
					},
					"useDefaultExcludes": &types.JsonSchemaProperty{
						Type:        "boolean",
						Description: "Optional. Whether to apply a list of default exclusion patterns (e.g., node_modules, .git, binary files). Defaults to true.",
					},
					"file_filtering_options": &types.JsonSchemaProperty{
						Type:        "object",
						Description: "Whether to respect ignore patterns from .gitignore or .goaiagentignore",
						Properties: map[string]*types.JsonSchemaProperty{
							"respect_git_ignore": &types.JsonSchemaProperty{
								Type:        "boolean",
								Description: "Optional: Whether to respect .gitignore patterns when finding files. Defaults to true.",
							},
							"respect_goaiagent_ignore": &types.JsonSchemaProperty{
								Type:        "boolean",
								Description: "Optional: Whether to respect .goaiagentignore patterns when finding files. Defaults to true.",
							},
						},
					},
				},
				Required: []string{"paths"},
			},
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
		fileSystemService: fs,
	}
}

// SkippedFile represents a file that was skipped during processing.
type SkippedFile struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// Execute performs a read-many-files operation.
func (t *ReadManyFilesTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	patterns, ok := args["paths"].([]any)
	if !ok || len(patterns) == 0 {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "Invalid or missing 'paths' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("invalid or missing 'paths' argument")
	}
	pathStrings := make([]string, len(patterns))
	for i, v := range patterns {
		pathStrings[i] = fmt.Sprint(v)
	}

	var includePatterns []string
	if include, ok := args["include"].([]any); ok {
		for _, v := range include {
			includePatterns = append(includePatterns, fmt.Sprint(v))
		}
	}

	var excludePatterns []string
	if exclude, ok := args["exclude"].([]any); ok {
		for _, v := range exclude {
			excludePatterns = append(excludePatterns, fmt.Sprint(v))
		}
	}

	recursive := true
	if r, ok := args["recursive"].(bool); ok {
		recursive = r
	}

	useDefaultExcludes := true
	if ude, ok := args["useDefaultExcludes"].(bool); ok {
		useDefaultExcludes = ude
	}

	respectGitIgnore := true
	respectGoaiagentIgnore := true

	if fileFilteringOptions, ok := args["file_filtering_options"].(map[string]any); ok {
		if val, ok := fileFilteringOptions["respect_git_ignore"].(bool); ok {
			respectGitIgnore = val
		}
		if val, ok := fileFilteringOptions["respect_goaiagent_ignore"].(bool); ok {
			respectGoaiagentIgnore = val
		}
	}

	var allFiles []string
	var processedFiles []string
	var skippedFilesList []SkippedFile
	var contentBuilder strings.Builder

	searchPatterns := append(pathStrings, includePatterns...)

	var compiledExcludeGlobs []glob.Glob

	// Get ignore patterns from files
	fileIgnoreGlobs, err := t.fileSystemService.GetIgnorePatterns(".", respectGitIgnore, respectGoaiagentIgnore)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to get ignore patterns from files: %v", err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to get ignore patterns from files: %w", err)
	}
	compiledExcludeGlobs = append(compiledExcludeGlobs, fileIgnoreGlobs...)

	if useDefaultExcludes {
		defaultExcludes := []string{"node_modules", ".git", ".goaiagent"}
		for _, p := range defaultExcludes {
			g, err := glob.Compile(p)
			if err != nil {
				// Log the error and continue without adding this exclude pattern
				fmt.Printf("Warning: failed to compile default exclude pattern %s: %v\n", p, err)
				continue
			}
			compiledExcludeGlobs = append(compiledExcludeGlobs, g)
		}
	}

	for _, p := range excludePatterns {
		g, err := glob.Compile(p)
		if err != nil {
			// Log the error and continue without adding this exclude pattern
			fmt.Printf("Warning: failed to compile exclude pattern %s: %v\n", p, err)
			continue
		}
		compiledExcludeGlobs = append(compiledExcludeGlobs, g)
	}

	var compiledSearchGlobs []glob.Glob
	for _, p := range searchPatterns {
		g, err := glob.Compile(p)
		if err != nil {
			// Log the error and continue without adding this search pattern
			fmt.Printf("Warning: failed to compile search pattern %s: %v\n", p, err)
			continue
		}
		compiledSearchGlobs = append(compiledSearchGlobs, g)
	}

	absSearchPath, err := filepath.Abs(".")
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to get absolute path: %v", err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to get absolute path: %w", err)
	}

	err = filepath.Walk(absSearchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from the search path
		relPath, err := filepath.Rel(absSearchPath, path)
		if err != nil {
			return err
		}

		// Check against ignore patterns
		for _, excludeG := range compiledExcludeGlobs {
			if excludeG.Match(relPath) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil // Skip this file
			}
		}

		if info.IsDir() {
			if !recursive && path != absSearchPath {
				return filepath.SkipDir
			}
			return nil
		}

		for _, g := range compiledSearchGlobs {
			if g.Match(relPath) {
				allFiles = append(allFiles, path)
				break // Move to next file once matched
			}
		}
		return nil
	})

	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Error walking path: %v", err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("error walking path: %w", err)
	}

	uniqueFiles := make(map[string]bool)
	for _, file := range allFiles {
		if _, exists := uniqueFiles[file]; !exists {
			uniqueFiles[file] = true

			content, err := os.ReadFile(file)
			if err != nil {
				skippedFilesList = append(skippedFilesList, SkippedFile{Path: file, Reason: fmt.Sprintf("failed to read: %v", err)})
				continue
			}

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

	return types.ToolResult{
		LLMContent:    contentBuilder.String() + "\n--- End of content ---\n\n" + displayMessage.String(),
		ReturnDisplay: displayMessage.String(),
	}, nil
}
