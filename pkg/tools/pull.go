package tools

import (
	"context"
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// PullTool implements the Tool interface for pulling latest changes from a Git remote.
type PullTool struct {
	*types.BaseDeclarativeTool
	gitService services.GitService
}

// NewPullTool creates a new PullTool.
func NewPullTool(gitService services.GitService) *PullTool {
	return &PullTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			types.PULL_TOOL_NAME,
			"Pull",
			"Pulls changes from a remote Git repository.",
			types.KindOther,
			(&types.JsonSchemaObject{
				Type: "object",
			}).SetProperties(map[string]*types.JsonSchemaProperty{
				"dir": {
					Type:        "string",
					Description: "The absolute path to the Git repository (e.g., '/home/user/project').",
				},
				"remote_name": {
					Type:        "string",
					Description: "Optional: The name of the Git remote (e.g., 'origin'). Defaults to 'origin'.",
				},
				"branch_name": {
					Type:        "string",
					Description: "Optional: The name of the branch to pull. Defaults to the current branch of the remote.",
				},
			}).SetRequired([]string{"dir"}),
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
		gitService: gitService,
	}
}

// Execute performs the pull operation.
func (t *PullTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	dir, ok := args["dir"].(string)
	if !ok || dir == "" {
		return types.ToolResult{}, fmt.Errorf("missing or invalid 'dir' argument")
	}

	remoteName, _ := args["remote_name"].(string)
	branchName, _ := args["branch_name"].(string)

	ref := ""
	if remoteName != "" && branchName != "" {
		ref = fmt.Sprintf("%s/%s", remoteName, branchName)
	} else if branchName != "" {
		ref = branchName
	}
	// If remoteName is provided but branchName is not, the tool will pull the default branch of that remote.
	// If neither are provided, gitService.Pull will likely default to the current branch's upstream.

	err := t.gitService.Pull(dir, ref)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to pull changes in %s (ref: %s): %v", dir, ref, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to pull changes in %s (ref: %s): %w", dir, ref, err)
	}

	successMessage := fmt.Sprintf("Successfully pulled latest changes in %s", dir)
	if ref != "" {
		successMessage += fmt.Sprintf(" for ref '%s'", ref)
	}
	successMessage += "."

	return types.ToolResult{
		LLMContent:    successMessage,
		ReturnDisplay: successMessage,
	}, nil
}
