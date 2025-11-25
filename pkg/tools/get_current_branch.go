package tools

import (
	"context"
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// GetCurrentBranchTool implements the Tool interface for getting the current Git branch.
type GetCurrentBranchTool struct {
	*types.BaseDeclarativeTool
	gitService services.GitService
}

// NewGetCurrentBranchTool creates a new GetCurrentBranchTool.
func NewGetCurrentBranchTool(gitService services.GitService) *GetCurrentBranchTool {
	return &GetCurrentBranchTool{
	BaseDeclarativeTool: types.NewBaseDeclarativeTool(
		types.GET_CURRENT_BRANCH_TOOL_NAME,
		"Get Current Branch",
		"Retrieves the name of the current Git branch.",
		types.KindOther,
		(&types.JsonSchemaObject{
			Type: "object",
		}).SetProperties(map[string]*types.JsonSchemaProperty{
			"dir": {
				Type:        "string",
				Description: "The absolute path to the Git repository (e.g., '/home/user/project').",
			},
		}).SetRequired([]string{"dir"}),
		false, // isOutputMarkdown
		false, // canUpdateOutput
		nil,   // MessageBus
	),
		gitService: gitService,
	}
}

// Execute performs the get_current_branch operation.
func (t *GetCurrentBranchTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	dir, ok := args["dir"].(string)
	if !ok || dir == "" {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "missing or invalid 'dir' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'dir' argument")
	}

	branch, err := t.gitService.GetCurrentBranch(dir)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to get current branch for %s: %v", dir, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to get current branch for %s: %w", dir, err)
	}

	return types.ToolResult{
		LLMContent:    branch,
		ReturnDisplay: fmt.Sprintf("Current Git branch in %s: %s", dir, branch),
	}, nil
}