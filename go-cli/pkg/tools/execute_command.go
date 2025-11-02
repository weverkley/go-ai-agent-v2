package tools

import (
	"fmt"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// ExecuteCommandTool implements the Tool interface for executing shell commands.
type ExecuteCommandTool struct {
	*types.BaseDeclarativeTool
	shellService *services.ShellExecutionService
}

// NewExecuteCommandTool creates a new ExecuteCommandTool.
func NewExecuteCommandTool() *ExecuteCommandTool {
	return &ExecuteCommandTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			"execute_command",
			"Execute Shell Command",
			"Executes a given shell command in the specified directory.",
			types.KindOther,
			types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]types.JsonSchemaProperty{
					"command": {
						Type:        "string",
						Description: "The shell command to execute.",
					},
					"dir": {
						Type:        "string",
						Description: "Optional: The absolute path of the directory to run the command in. If not provided, the current working directory is used.",
					},
				},
				Required: []string{"command"},
			},
			false,
			false,
			nil,
		),
		shellService: services.NewShellExecutionService(),
	}
}

// Execute performs the execute_command operation.
func (t *ExecuteCommandTool) Execute(args map[string]any) (types.ToolResult, error) {
	command, ok := args["command"].(string)
	if !ok || command == "" {
		return types.ToolResult{}, fmt.Errorf("missing or invalid 'command' argument")
	}

	dir := "."
	if d, ok := args["dir"].(string); ok && d != "" {
		dir = d
	}

	stdout, stderr, err := t.shellService.ExecuteCommand(command, dir)
	
	output := strings.Builder{}
	if stdout != "" {
		output.WriteString("Stdout:\n")
		output.WriteString(stdout)
	}
	if stderr != "" {
		output.WriteString("Stderr:\n")
		output.WriteString(stderr)
	}
	if err != nil {
		output.WriteString(fmt.Sprintf("Error: %v\n", err))
	}

	llmContent := output.String()
	if llmContent == "" {
		llmContent = "Command executed successfully with no output."
	}

	return types.ToolResult{
		LLMContent:    llmContent,
		ReturnDisplay: llmContent,
	},
	nil
}

// Definition returns the tool definition for the Gemini API.
func (t *ExecuteCommandTool) Definition() *genai.Tool {
	return t.BaseDeclarativeTool.Definition()
}
