package tools

import (
	"fmt"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// GetRemoteURLTool implements the Tool interface for getting the Git remote URL.
type GetRemoteURLTool struct {
	*types.BaseDeclarativeTool
	gitService *services.GitService
}

// NewGetRemoteURLTool creates a new GetRemoteURLTool.
func NewGetRemoteURLTool() *GetRemoteURLTool {
	return &GetRemoteURLTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			"get_remote_url",
			"Get Git Remote URL",
			"Returns the URL of the 'origin' remote for the given Git repository.",
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

// Execute performs the get_remote_url operation.
func (t *GetRemoteURLTool) Execute(args map[string]any) (types.ToolResult, error) {
	dir, ok := args["dir"].(string)
	if !ok || dir == "" {
		return types.ToolResult{}, fmt.Errorf("missing or invalid 'dir' argument")
	}

	remoteURL, err := t.gitService.GetRemoteURL(dir)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to get remote URL for %s: %w", dir, err)
	}

	return types.ToolResult{
		LLMContent:    remoteURL,
		ReturnDisplay: fmt.Sprintf("Git remote URL for %s: %s", dir, remoteURL),
	}, nil
}

// Definition returns the tool definition for the Gemini API.
func (t *GetRemoteURLTool) Definition() *genai.Tool {
	return t.BaseDeclarativeTool.Definition()
}