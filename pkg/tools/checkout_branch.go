package tools

import (
	"context"
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// CheckoutBranchTool implements the Tool interface for checking out a Git branch.
type CheckoutBranchTool struct {
	*types.BaseDeclarativeTool
	gitService services.GitService
}

// NewCheckoutBranchTool creates a new CheckoutBranchTool.
func NewCheckoutBranchTool(gitService services.GitService) *CheckoutBranchTool {
	return &CheckoutBranchTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			types.CHECKOUT_BRANCH_TOOL_NAME,
			"Checkout Branch",
			"Checks out a Git branch.",
			types.KindExecute,
			(&types.JsonSchemaObject{
				Type: "object",
			}).SetProperties(map[string]*types.JsonSchemaProperty{
				"branch_name": {
					Type:        "string",
					Description: "The name of the Git branch to checkout.",
				},
				"dir": {
					Type:        "string",
					Description: "The absolute path to the Git repository (e.g., '/home/user/project').",
				},
			}).SetRequired([]string{"branch_name", "dir"}),
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
		gitService: gitService,
	}
}

// Execute performs the checkout_branch operation.
func (t *CheckoutBranchTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	dir, ok := args["dir"].(string)
	if !ok || dir == "" {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "missing or invalid 'dir' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'dir' argument")
	}
	branchName, ok := args["branch_name"].(string)
	if !ok || branchName == "" {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "missing or invalid 'branch_name' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("missing or invalid 'branch_name' argument")
	}

	err := t.gitService.CheckoutBranch(dir, branchName)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to checkout branch %s in %s: %v", branchName, dir, err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to checkout branch %s in %s: %w", branchName, dir, err)
	}

	return types.ToolResult{
		LLMContent:    fmt.Sprintf("Successfully checked out branch %s in %s.", branchName, dir),
		ReturnDisplay: fmt.Sprintf("Successfully checked out branch %s in %s.", branchName, dir),
	}, nil
}
