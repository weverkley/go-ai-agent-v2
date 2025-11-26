package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/types"
)

const (
	TODOS_FILENAME = "TODOS.md"
)

// Todo represents a single todo item.
type Todo struct {
	Description string `json:"description"`
	Status      string `json:"status"`
}

// WriteTodosTool represents the write-todos tool.
type WriteTodosTool struct {
	*types.BaseDeclarativeTool
	settingsService types.SettingsServiceIface
}

// NewWriteTodosTool creates a new instance of WriteTodosTool.
func NewWriteTodosTool(settingsService types.SettingsServiceIface) *WriteTodosTool {
	return &WriteTodosTool{
		BaseDeclarativeTool: types.NewBaseDeclarativeTool(
			types.WRITE_TODOS_TOOL_NAME,
			"Write Todos",
			"Creates or replaces the full list of subtasks required to complete a user request.",
			types.KindEdit,
			(&types.JsonSchemaObject{
				Type: "object",
			}).SetProperties(map[string]*types.JsonSchemaProperty{
				"todos": &types.JsonSchemaProperty{
					Type:        "array",
					Description: "The complete list of todo items. This will replace the existing list.",
					Items: (&types.JsonSchemaObject{
						Type: "object",
					}).SetProperties(map[string]*types.JsonSchemaProperty{
						"description": &types.JsonSchemaProperty{
							Type:        "string",
							Description: "The description of the task.",
						},
						"status": &types.JsonSchemaProperty{
							Type:        "string",
							Description: "The current status of the task.",
							Enum:        []string{"pending", "in_progress", "completed", "cancelled"},
						},
					}).SetRequired([]string{"description", "status"}),
				},
			}).SetRequired([]string{"todos"}),
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
		settingsService: settingsService,
	}
}

// getTodosFilePath returns the path to the TODOS.md file within the workspace.
func (t *WriteTodosTool) getTodosFilePath() (string, error) {
	workspaceDir := t.settingsService.GetWorkspaceDir()
	if workspaceDir == "" {
		// Fallback or error if workspace directory is not available
		homeDir, err := osUserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		return filepath.Join(homeDir, ".goaiagent", TODOS_FILENAME), nil
	}
	return filepath.Join(workspaceDir, ".goaiagent", TODOS_FILENAME), nil
}

// Execute writes the todos to a file.
func (t *WriteTodosTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	todosData, ok := args["todos"].([]any)
	if !ok {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "Invalid or missing 'todos' argument",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("invalid or missing 'todos' argument")
	}

	var todos []Todo
	for _, item := range todosData {
		todoMap, ok := item.(map[string]any)
		if !ok {
			return types.ToolResult{
				Error: &types.ToolError{
					Message: "Invalid todo item format",
					Type:    types.ToolErrorTypeExecutionFailed,
				},
			}, fmt.Errorf("invalid todo item format")
		}
		desc, _ := todoMap["description"].(string)
		status, _ := todoMap["status"].(string)
		todos = append(todos, Todo{Description: desc, Status: status})
	}

	// Validate todos
	inProgressCount := 0
	for _, todo := range todos {
		if todo.Description == "" {
			return types.ToolResult{
				Error: &types.ToolError{
					Message: "Each todo must have a non-empty description",
					Type:    types.ToolErrorTypeExecutionFailed,
				},
			}, fmt.Errorf("each todo must have a non-empty description")
		}
		if !t.isValidTodoStatus(todo.Status) { // Call isValidTodoStatus as a method
			return types.ToolResult{
				Error: &types.ToolError{
					Message: fmt.Sprintf("Invalid todo status: %s", todo.Status),
					Type:    types.ToolErrorTypeExecutionFailed,
				},
			}, fmt.Errorf("invalid todo status: %s", todo.Status)
		}
		if todo.Status == "in_progress" {
			inProgressCount++
		}
	}

	if inProgressCount > 1 {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: "Only one task can be \"in_progress\" at a time",
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("only one task can be \"in_progress\" at a time")
	}

	todosFilePath, err := t.getTodosFilePath()
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to get todos file path: %v", err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, err
	}

	err = osMkdirAll(filepath.Dir(todosFilePath), 0755)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to create todos directory: %v", err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to create todos directory: %w", err)
	}

	// Format todos for writing to file
	var todoListBuilder strings.Builder
	if len(todos) > 0 {
		todoListBuilder.WriteString("# ToDo List\n\n")
		for i, todo := range todos {
			todoListBuilder.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, todo.Status, todo.Description))
		}
	}

	err = osWriteFile(todosFilePath, []byte(todoListBuilder.String()), 0644)
	if err != nil {
		return types.ToolResult{
			Error: &types.ToolError{
				Message: fmt.Sprintf("Failed to write todos file: %v", err),
				Type:    types.ToolErrorTypeExecutionFailed,
			},
		}, fmt.Errorf("failed to write todos file: %w", err)
	}

	displayContent := "Successfully updated the todo list."
	if len(todos) > 0 {
		displayContent += fmt.Sprintf(" The current list is now:\n%s", todoListBuilder.String())
	} else {
		displayContent = "Successfully cleared the todo list."
	}

	// Provide a simple, structured response for the model.
	llmContent := map[string]interface{}{
		"success": true,
		"message": "Successfully updated the todo list.",
	}

	return types.ToolResult{
		LLMContent:    llmContent,
		ReturnDisplay: displayContent,
	}, nil
}

// isValidTodoStatus checks if the given status is a valid todo status.
func (t *WriteTodosTool) isValidTodoStatus(status string) bool {
	for _, s := range []string{"pending", "in_progress", "completed", "cancelled"} {
		if s == status {
			return true
		}
	}
	return false
}
