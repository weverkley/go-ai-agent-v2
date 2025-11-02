package tools

import (
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// PullTool implements the Tool interface for pulling latest changes from a Git remote.
type PullTool struct {
	*types.BaseDeclarativeTool
	gitService *services.GitService
}

// NewPullTool creates a new PullTool.
func NewPullTool() *PullTool {
	return &PullTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			"pull",
			"Pull Git Changes",
			"Pulls the latest changes from the remote for the current branch in the given Git repository.",
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

// Execute performs the pull operation.
func (t *PullTool) Execute(args map[string]any) (types.ToolResult, error) {
	dir, ok := args["dir"].(string)
	if !ok || dir == "" {
		return types.ToolResult{}, fmt.Errorf("missing or invalid 'dir' argument")
	}

	err := t.gitService.Pull(dir)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to pull changes in %s: %w", dir, err)
	}

	return types.ToolResult{
		LLMContent:    fmt.Sprintf("Successfully pulled latest changes in %s.", dir),
		ReturnDisplay: fmt.Sprintf("Successfully pulled latest changes in %s.", dir),
	}, nil
}

// Definition returns the tool definition for the Gemini API.
func (t *PullTool) Definition() *genai.Tool {
	return t.BaseDeclarativeTool.Definition()
}
