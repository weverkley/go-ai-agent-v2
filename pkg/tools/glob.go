package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go-ai-agent-v2/go-cli/pkg/services" // Added
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/gobwas/glob"
)

// GlobTool represents the glob tool.
type GlobTool struct {
	*types.BaseDeclarativeTool
	fileSystemService services.FileSystemService
	workspaceService  types.WorkspaceServiceIface
}

// NewGlobTool creates a new instance of GlobTool.
func NewGlobTool(fileSystemService services.FileSystemService, workspaceService types.WorkspaceServiceIface) *GlobTool {
	return &GlobTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			types.GLOB_TOOL_NAME,
			types.GLOB_TOOL_DISPLAY_NAME,
			"Efficiently finds files matching specific glob patterns.",
			types.KindOther, // Assuming KindOther for now
			&types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]*types.JsonSchemaProperty{
					"pattern": &types.JsonSchemaProperty{
						Type:        "string",
						Description: "The glob pattern to match against (e.g., '**/*.py', 'docs/*.md').",
					},
					"path": &types.JsonSchemaProperty{
						Type:        "string",
						Description: "Optional: The path to the directory to search within, relative to the project root. If omitted, searches the current working directory.",
					},
					"case_sensitive": &types.JsonSchemaProperty{
						Type:        "boolean",
						Description: "Optional: Whether the search should be case-sensitive. Defaults to false.",
					},
					"respect_git_ignore": &types.JsonSchemaProperty{
						Type:        "boolean",
						Description: "Optional: Whether to respect .gitignore patterns when finding files. Defaults to true.",
					},
					"respect_goaiagent_ignore": &types.JsonSchemaProperty{
						Type:        "boolean",
						Description: "Optional: Whether to respect .goaiagentignore patterns when finding files. Defaults to true.",
					},
				},
				Required: []string{"pattern"},
			},
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
		fileSystemService: fileSystemService,
		workspaceService:  workspaceService,
	}
}

// FileInfo represents a file found by glob.
type FileInfo struct {
	Path    string
	ModTime time.Time
}

// Execute performs a glob search.
func (t *GlobTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	pattern, ok := args["pattern"].(string)
	if !ok || pattern == "" {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "Invalid or missing 'pattern' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("invalid or missing 'pattern' argument")
	}

	searchPath := "."
	if p, ok := args["path"].(string); ok && p != "" {
		searchPath = p
	}

	projectRoot := t.workspaceService.GetProjectRoot()
	absolutePath := filepath.Join(projectRoot, searchPath)

	caseSensitive := false
	if cs, ok := args["case_sensitive"].(bool); ok {
		caseSensitive = cs
	}

	respectGitIgnore := true
	if rgi, ok := args["respect_git_ignore"].(bool); ok {
		respectGitIgnore = rgi
	}

	respectGoaiagentIgnore := true
	if rgi, ok := args["respect_goaiagent_ignore"].(bool); ok {
		respectGoaiagentIgnore = rgi
	}

	// Resolve the search path
	absSearchPath, err := filepath.Abs(absolutePath)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to resolve absolute path for %s: %v", searchPath, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to resolve absolute path for %s: %w", searchPath, err)
	}

	// Compile the main glob pattern
	compiledPattern := pattern
	if !caseSensitive {
		compiledPattern = strings.ToLower(pattern)
	}

	// If the pattern contains "**", we need to modify the pattern to also match files directly in the current directory.
	// For example, "**/*.go" should match "main.go" and "src/app.go".
	// The glob "**/*.go" typically matches "src/app.go" but not "main.go".
	// So, we transform "**/*.go" to "{*.go,**/*.go}".
	if strings.Contains(pattern, "**") {
		rootPattern := filepath.Base(pattern)
		if !caseSensitive {
			rootPattern = strings.ToLower(rootPattern)
		}
		compiledPattern = fmt.Sprintf("{%s,%s}", rootPattern, compiledPattern)
	}

	g, err := glob.Compile(compiledPattern)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to compile glob pattern %s: %v", pattern, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to compile glob pattern %s: %w", pattern, err)
	}

	// Get ignore patterns using FileSystemService
	ignorePatterns, err := t.fileSystemService.GetIgnorePatterns(absSearchPath, respectGitIgnore, respectGoaiagentIgnore)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to get ignore patterns: %v", err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to get ignore patterns: %w", err)
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
			if ignoreG.Match(relPath) {
				return nil // Skip this file
			}
		}

		// Check against main glob pattern
		isMatched := g.Match(matchPath)

		if isMatched {
			matchedFiles = append(matchedFiles, FileInfo{
				Path:    path,
				ModTime: info.ModTime(),
			})
		}
		return nil
	})
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Error walking the path %s: %v", absSearchPath, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("error walking the path %s: %w", absSearchPath, err)
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
		return types.ToolResult{
			LLMContent:    fmt.Sprintf("No files found matching pattern \"%s\" in path \"%s\"", pattern, searchPath),
			ReturnDisplay: fmt.Sprintf("No files found matching pattern \"%s\" in path \"%s\"", pattern, searchPath),
		}, nil
	}

	resultMessage := fmt.Sprintf("Found %d file(s) matching \"%s\" in path \"%s\":\n%s",
		len(resultPaths), pattern, searchPath, strings.Join(resultPaths, "\n"))

	return types.ToolResult{
		LLMContent:    resultMessage,
		ReturnDisplay: resultMessage,
	}, nil
}
