package tools

import (
	"context"
	"fmt"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// ExecuteCommandTool implements the Tool interface for executing shell commands.
type ExecuteCommandTool struct {
	*types.BaseDeclarativeTool
	shellService *services.ShellExecutionService
}

// NewExecuteCommandTool creates a new ExecuteCommandTool.
func NewExecuteCommandTool(shellService *services.ShellExecutionService) *ExecuteCommandTool {
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
					"background": {
						Type:        "boolean",
						Description: "Optional: If true, the command will be executed in the background. The tool will return immediately with the process ID. Default is false.",
					},
				},
				Required: []string{"command"},
			},
			false,
			false,
			nil,
		),
		shellService: shellService,
	}
}

// Execute performs the execute_command operation.
func (t *ExecuteCommandTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	command, ok := args["command"].(string)
	if !ok || command == "" {
		return types.ToolResult{}, fmt.Errorf("missing or invalid 'command' argument")
	}

	dir := "."
	if d, ok := args["dir"].(string); ok && d != "" {
		dir = d
	}

	background := false
	if b, ok := args["background"].(bool); ok {
		background = b
	}

	if background {
		pid, err := t.shellService.ExecuteCommandInBackground(command, dir)
		if err != nil {
			return types.ToolResult{}, fmt.Errorf("failed to execute command in background: %w", err)
		}
		return types.ToolResult{
			LLMContent:    fmt.Sprintf("Command '%s' started in background with PID %d", command, pid),
			ReturnDisplay: fmt.Sprintf("Command '%s' started in background with PID %d", command, pid),
		}, nil
	}

	stdout, stderr, err := t.shellService.ExecuteCommand(ctx, command, dir)
	
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
