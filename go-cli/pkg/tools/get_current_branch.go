package tools

import (
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// GetCurrentBranchTool implements the Tool interface for getting the current Git branch.
type GetCurrentBranchTool struct {
	*types.BaseDeclarativeTool
	gitService *services.GitService
}

// NewGetCurrentBranchTool creates a new GetCurrentBranchTool.
func NewGetCurrentBranchTool() *GetCurrentBranchTool {
	return &GetCurrentBranchTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			"get_current_branch",
			"Get Current Git Branch",
			"Returns the name of the current Git branch in the specified directory.",
			types.KindOther,
			types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]types.JsonSchemaProperty{
					"dir": {
						Type:        "string",
						Description: "The absolute path to the Git repository (e.g., '/home/user/project').",
					},
				},
				Required: []string{"dir"},
			},
			false,
			false,
			nil,
		),
		gitService: services.NewGitService(),
	}
}

// Execute performs the get_current_branch operation.
func (t *GetCurrentBranchTool) Execute(args map[string]any) (types.ToolResult, error) {
	dir, ok := args["dir"].(string)
	if !ok || dir == "" {
		return types.ToolResult{}, fmt.Errorf("missing or invalid 'dir' argument")
	}

	branch, err := t.gitService.GetCurrentBranch(dir)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to get current branch for %s: %w", dir, err)
	}

	return types.ToolResult{
		LLMContent:    branch,
		ReturnDisplay: fmt.Sprintf("Current Git branch in %s: %s", dir, branch),
	}, nil
}

// Definition returns the tool definition for the Gemini API.
func (t *GetCurrentBranchTool) Definition() *genai.Tool {
	return t.BaseDeclarativeTool.Definition()
}