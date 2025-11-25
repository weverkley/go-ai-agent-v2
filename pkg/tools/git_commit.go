package tools

import (
	"context"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"
)

const GIT_COMMIT_TOOL_NAME = "git_commit"

// GitCommitTool implements the Tool interface for staging files and creating a Git commit.
type GitCommitTool struct {
	*types.BaseDeclarativeTool
	gitService services.GitService
}

// NewGitCommitTool creates a new GitCommitTool.
func NewGitCommitTool(gitService services.GitService) *GitCommitTool {
	return &GitCommitTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			GIT_COMMIT_TOOL_NAME,
			"Git Commit",
			"Stages files and creates a Git commit.",
			types.KindOther,
			(&types.JsonSchemaObject{
				Type: "object",
			}).SetProperties(map[string]*types.JsonSchemaProperty{
				"message": {
					Type:        "string",
					Description: "The commit message.",
				},
				"files_to_stage": {
					Type:        "array",
					Description: "Optional: A list of specific files to stage. If not provided, all tracked and modified files will be staged.",
					Items:       &types.JsonSchemaObject{Type: "string"},
				},
				"dir": {
					Type:        "string",
					Description: "Optional: The absolute path to the Git repository. Defaults to the current working directory.",
				},
			}).SetRequired([]string{"message"}),
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
		gitService: gitService,
	}
}

// Execute performs the git_commit operation.
func (t *GitCommitTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	message, ok := args["message"].(string)
	if !ok || message == "" {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "missing or invalid 'message' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'message' argument")
	}

	var filesToStage []string
	if files, ok := args["files_to_stage"].([]any); ok {
		for _, f := range files {
			if fileName, isString := f.(string); isString {
				filesToStage = append(filesToStage, fileName)
			}
		}
	}

	dir := "."
	if d, ok := args["dir"].(string); ok && d != "" {
		dir = d
	}

	err := t.gitService.StageFiles(dir, filesToStage)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("failed to stage files in %s: %v", dir, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to stage files in %s: %w", dir, err)
	}

	err = t.gitService.Commit(dir, message)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("failed to commit changes in %s: %v", dir, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to commit changes in %s: %w", dir, err)
	}

	return types.ToolResult{
		LLMContent:    fmt.Sprintf("Successfully committed changes in %s with message: \"%s\"", dir, message),
		ReturnDisplay: fmt.Sprintf("Successfully committed changes in %s with message: \"%s\"", dir, message),
	}, nil
}
