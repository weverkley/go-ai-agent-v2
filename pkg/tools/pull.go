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
			"remote_name": &types.JsonSchemaProperty{
				Type:        "string",
				Description: "The name of the Git remote (e.g., 'origin'). Defaults to 'origin'.",
			},
			"branch_name": &types.JsonSchemaProperty{
				Type:        "string",
				Description: "The name of the branch to pull. Defaults to the current branch.",
			},
		}),
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

	err := t.gitService.Pull(dir, "")
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to pull changes in %s: %v", dir, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to pull changes in %s: %w", dir, err)
	}

	return types.ToolResult{
		LLMContent:    fmt.Sprintf("Successfully pulled latest changes in %s.", dir),
		ReturnDisplay: fmt.Sprintf("Successfully pulled latest changes in %s.", dir),
	}, nil
}
