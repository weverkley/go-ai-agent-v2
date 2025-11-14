package tools

import (
	"context"
	"fmt"
	"os"
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
}

// NewWriteTodosTool creates a new instance of WriteTodosTool.
func NewWriteTodosTool() *WriteTodosTool {
	return &WriteTodosTool{
		types.NewBaseDeclarativeTool(
			"write_todos",
			"write_todos",
			"This tool can help you list out the current subtasks that are required to be completed for a given user request.",
			types.KindOther, // Assuming KindOther for now
			types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]types.JsonSchemaProperty{
					"todos": {
						Type:        "array",
						Description: "The complete list of todo items. This will replace the existing list.",
						Items: &types.JsonSchemaPropertyItem{
							Type: "object", // This should be object, not string
						},
					},
				},
				Required: []string{"todos"},
			},
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
	}
}

// getTodosFilePath returns the path to the TODOS.md file.
func getTodosFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".gemini", TODOS_FILENAME), nil
}

// Execute writes the todos to a file.
func (t *WriteTodosTool) Execute(ctx context.Context, args map[string]any) (types.ToolResult, error) {
	todosData, ok := args["todos"].([]any)
	if !ok {
		return types.ToolResult{}, fmt.Errorf("invalid or missing 'todos' argument")
	}

	var todos []Todo
	for _, item := range todosData {
		todoMap, ok := item.(map[string]any)
		if !ok {
			return types.ToolResult{}, fmt.Errorf("invalid todo item format")
		}
		desc, _ := todoMap["description"].(string)
		status, _ := todoMap["status"].(string)
		todos = append(todos, Todo{Description: desc, Status: status})
	}

	// Validate todos
	inProgressCount := 0
	for _, todo := range todos {
		if todo.Description == "" {
			return types.ToolResult{}, fmt.Errorf("each todo must have a non-empty description")
		}
		if !isValidTodoStatus(todo.Status) {
			return types.ToolResult{}, fmt.Errorf("invalid todo status: %s", todo.Status)
		}
		if todo.Status == "in_progress" {
			inProgressCount++
		}
	}

	if inProgressCount > 1 {
		return types.ToolResult{}, fmt.Errorf("only one task can be \"in_progress\" at a time")
	}

	todosFilePath, err := getTodosFilePath()
	if err != nil {
		return types.ToolResult{}, err
	}

	err = os.MkdirAll(filepath.Dir(todosFilePath), 0755)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to create todos directory: %w", err)
	}

	// Format todos for writing to file
	var todoListBuilder strings.Builder
	if len(todos) > 0 {
		todoListBuilder.WriteString("# ToDo List\n\n")
		for i, todo := range todos {
			todoListBuilder.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, todo.Status, todo.Description))
		}
	}

	err = os.WriteFile(todosFilePath, []byte(todoListBuilder.String()), 0644)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to write todos file: %w", err)
	}

	llmContent := "Successfully updated the todo list."
	if len(todos) > 0 {
		llmContent += fmt.Sprintf(" The current list is now:\n%s", todoListBuilder.String())
	} else {
		llmContent = "Successfully cleared the todo list."
	}

	return types.ToolResult{
		LLMContent:    llmContent,
		ReturnDisplay: llmContent,
	}, nil
}

func isValidTodoStatus(status string) bool {
	for _, s := range []string{"pending", "in_progress", "completed", "cancelled"} {
		if s == status {
			return true
		}
	}
	return false
}
