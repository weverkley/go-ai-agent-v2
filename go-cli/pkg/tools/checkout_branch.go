package tools

import (
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// CheckoutBranchTool implements the Tool interface for checking out a Git branch.
type CheckoutBranchTool struct {
	*types.BaseDeclarativeTool
	gitService *services.GitService
}

// NewCheckoutBranchTool creates a new CheckoutBranchTool.
func NewCheckoutBranchTool() *CheckoutBranchTool {
	return &CheckoutBranchTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			"checkout_branch",
			"Checkout Git Branch",
			"Checks out the specified branch in the given Git repository.",
			types.KindOther,
			types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]types.JsonSchemaProperty{
					"dir": {
						Type:        "string",
						Description: "The absolute path to the Git repository (e.g., '/home/user/project').",
					},
					"branch_name": {
						Type:        "string",
						Description: "The name of the branch to checkout.",
					},
				},
				Required: []string{"dir", "branch_name"},
			},
			false,
			false,
			nil,
		),
		gitService: services.NewGitService(),
	}
}

// Execute performs the checkout_branch operation.
func (t *CheckoutBranchTool) Execute(args map[string]any) (types.ToolResult, error) {
	dir, ok := args["dir"].(string)
	if !ok || dir == "" {
		return types.ToolResult{}, fmt.Errorf("missing or invalid 'dir' argument")
	}
	branchName, ok := args["branch_name"].(string)
	if !ok || branchName == "" {
		return types.ToolResult{}, fmt.Errorf("missing or invalid 'branch_name' argument")
	}

	err := t.gitService.CheckoutBranch(dir, branchName)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to checkout branch %s in %s: %w", branchName, dir, err)
	}

	return types.ToolResult{
		LLMContent:    fmt.Sprintf("Successfully checked out branch %s in %s.", branchName, dir),
		ReturnDisplay: fmt.Sprintf("Successfully checked out branch %s in %s.", branchName, dir),
	}, nil
}

// Definition returns the tool definition for the Gemini API.
func (t *CheckoutBranchTool) Definition() *genai.Tool {
	return t.BaseDeclarativeTool.Definition()
}
