package tools

import (
	"fmt"
	"os"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

// SmartEditTool represents the smart-edit tool.
type SmartEditTool struct{}

// NewSmartEditTool creates a new instance of SmartEditTool.
func NewSmartEditTool() *SmartEditTool {
	return &SmartEditTool{}
}

// Name returns the name of the tool.
func (t *SmartEditTool) Name() string {
	return "smart_edit"
}

// Definition returns the tool's definition for the Gemini API.
func (t *SmartEditTool) Definition() *genai.Tool {
	return &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        t.Name(),
				Description: "Replaces text within a file. Replaces a single occurrence.",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"file_path": {
							Type:        genai.TypeString,
							Description: "The absolute path to the file to modify. Must start with '/'.",
						},
						"instruction": {
							Type:        genai.TypeString,
							Description: "A clear, semantic instruction for the code change.",
						},
						"old_string": {
							Type:        genai.TypeString,
							Description: "The exact literal text to replace.",
						},
						"new_string": {
							Type:        genai.TypeString,
							Description: "The exact literal text to replace old_string with.",
						},
					},
					Required: []string{"file_path", "instruction", "old_string", "new_string"},
				},
			},
		},
	}
}

// Execute performs a smart edit operation.
func (t *SmartEditTool) Execute(args map[string]any) (types.ToolResult, error) {
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return types.ToolResult{}, fmt.Errorf("invalid or missing 'file_path' argument")
	}

	// instruction is mainly for the LLM, not used directly in this simplified Go version yet.
	_, ok = args["instruction"].(string)
	if !ok {
		return types.ToolResult{}, fmt.Errorf("invalid or missing 'instruction' argument")
	}

	oldString, _ := args["old_string"].(string)

	newString, ok := args["new_string"].(string)
	if !ok {
		return types.ToolResult{}, fmt.Errorf("invalid or missing 'new_string' argument")
	}

	// Read the file content
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		// If file doesn't exist and old_string is empty, it's a new file creation
		if os.IsNotExist(err) && oldString == "" {
			err = os.WriteFile(filePath, []byte(newString), 0644)
			if err != nil {
				return types.ToolResult{}, fmt.Errorf("failed to create new file %s: %w", filePath, err)
			}
			successMessage := fmt.Sprintf("Successfully created new file: %s", filePath)
			return types.ToolResult{
				LLMContent:    successMessage,
				ReturnDisplay: successMessage,
			}, nil
		}
		return types.ToolResult{}, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	content := string(contentBytes)

	// If old_string is empty but file exists, it's an error
	if oldString == "" {
		return types.ToolResult{}, fmt.Errorf("file already exists, cannot create: %s. Use a non-empty old_string to modify.", filePath)
	}

	// Perform the replacement
	newContent := strings.Replace(content, oldString, newString, 1) // Replace only the first occurrence

	if newContent == content {
		return types.ToolResult{}, fmt.Errorf("old_string not found or no changes made in file %s", filePath)
	}

	// Write the new content back to the file
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	successMessage := fmt.Sprintf("Successfully modified file: %s", filePath)
	return types.ToolResult{
		LLMContent:    successMessage,
		ReturnDisplay: successMessage,
	}, nil
}
