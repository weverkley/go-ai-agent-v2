package tools

import (
	"context"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/services"
	"go-ai-agent-v2/go-cli/pkg/types"
	"path/filepath"
)

const RUN_TESTS_TOOL_NAME = "run_tests"

// RunTestsTool is a dedicated tool to execute project-specific tests.
type RunTestsTool struct {
	*types.BaseDeclarativeTool
	shellService     services.ShellExecutionService
	fsService        services.FileSystemService
	workspaceService types.WorkspaceServiceIface
}

// NewRunTestsTool creates a new RunTestsTool.
func NewRunTestsTool(shellService services.ShellExecutionService, fsService services.FileSystemService, workspaceService types.WorkspaceServiceIface) *RunTestsTool {
	return &RunTestsTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			RUN_TESTS_TOOL_NAME,
			"Run Tests",
			"A dedicated tool to execute project-specific tests. It intelligently discovers and uses the correct test command for the project (e.g., `go test`, `npm test`, `pytest`).",
			types.KindExecute,
			&types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]*types.JsonSchemaProperty{
					"target": {
						Type:        "string",
						Description: "Optional: The specific test file, directory, or test name/pattern to run. Defaults to all tests.",
					},
					"coverage": {
						Type:        "boolean",
						Description: "Optional: If true, generates and includes a code coverage report in the output. Defaults to false.",
					},
					"dir": {
						Type:        "string",
						Description: "Optional: The directory to run the tests in, relative to the project root. Defaults to the current working directory.",
					},
				},
			},
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
		shellService:     shellService,
		fsService:        fsService,
		workspaceService: workspaceService,
	}
}

// Execute performs the run_tests operation.
func (t *RunTestsTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	target, _ := args["target"].(string)
	coverage, _ := args["coverage"].(bool)
	dir := "."
	if d, ok := args["dir"].(string); ok && d != "" {
		dir = d
	}

	projectRoot := t.workspaceService.GetProjectRoot()
	absolutePath := filepath.Join(projectRoot, dir)

	// Detect project type and build command
	var command string
	if exists, _ := t.fsService.PathExists(filepath.Join(absolutePath, "go.mod")); exists {
		// Go project
		command = "go test"
		if target != "" {
			command += " " + target
		} else {
			command += " ./..."
		}
		if coverage {
			command += " -cover"
		}
	} else if exists, _ := t.fsService.PathExists(filepath.Join(absolutePath, "package.json")); exists {
		// Node.js project
		command = "npm test"
		if target != "" {
			command += " -- " + target
		}
		if coverage {
			command += " -- --coverage"
		}
	} else if exists, _ := t.fsService.PathExists(filepath.Join(absolutePath, "requirements.txt")); exists {
		// Python project
		command = "pytest"
		if target != "" {
			command += " " + target
		}
		if coverage {
			command += " --cov"
		}
	} else {
		return types.ToolResult{
				Error: &types.ToolError{
					Message: "could not determine project type to run tests",
					Type:    types.ToolErrorTypeExecutionFailed,
				},
			},
			fmt.Errorf("could not determine project type to run tests in directory %s", absolutePath)
	}

	stdout, stderr, err := t.shellService.ExecuteCommand(ctx, command, absolutePath)

	output := "Stdout:\n" + stdout + "\nStderr:\n" + stderr
	if err != nil {
		output += "\nError: " + err.Error()
		return types.ToolResult{
				LLMContent:    output,
				ReturnDisplay: output,
				Error: &types.ToolError{
					Message: fmt.Sprintf("test command failed: %v", err),
					Type:    types.ToolErrorTypeExecutionFailed,
				},
			},
			err
	}

	return types.ToolResult{
			LLMContent:    output,
			ReturnDisplay: output,
		},
		nil
}
