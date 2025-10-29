package tools

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
}

// NewWriteTodosTool creates a new instance of WriteTodosTool.
func NewWriteTodosTool() *WriteTodosTool {
	return &WriteTodosTool{}
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
func (t *WriteTodosTool) Execute(
	todos []Todo,
) (string, error) {
	// Validate todos
	inProgressCount := 0
	for _, todo := range todos {
		if todo.Description == "" {
			return "", fmt.Errorf("each todo must have a non-empty description")
		}
		if !isValidTodoStatus(todo.Status) {
			return "", fmt.Errorf("invalid todo status: %s", todo.Status)
		}
		if todo.Status == "in_progress" {
			inProgressCount++
		}
	}

	if inProgressCount > 1 {
		return "", fmt.Errorf("only one task can be \"in_progress\" at a time")
	}

	todosFilePath, err := getTodosFilePath()
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(filepath.Dir(todosFilePath), 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create todos directory: %w", err)
	}

	// Format todos for writing to file
	var todoListBuilder strings.Builder
	if len(todos) > 0 {
		todoListBuilder.WriteString("# ToDo List\n\n")
		for i, todo := range todos {
			todoListBuilder.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, todo.Status, todo.Description))
		}
	}

	err = ioutil.WriteFile(todosFilePath, []byte(todoListBuilder.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write todos file: %w", err)
	}

	llmContent := "Successfully updated the todo list."
	if len(todos) > 0 {
		llmContent += fmt.Sprintf(" The current list is now:\n%s", todoListBuilder.String())
	} else {
		llmContent = "Successfully cleared the todo list."
	}

	return llmContent, nil
}

func isValidTodoStatus(status string) bool {
	for _, s := range []string{"pending", "in_progress", "completed", "cancelled"} {
		if s == status {
			return true
		}
	}
	return false
}
