package agents

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// TestWriterAgent defines the Test Writer subagent.
var TestWriterAgent = AgentDefinition{
	Name:        "test_writer",
	DisplayName: "Test Writer Agent",
	Description: "", // Will be populated in init()
	InputConfig: InputConfig{
		Inputs: map[string]InputParameter{
			"source_file_path": {
				Description: "The path to the source file containing the symbol to test.",
				Type:        "string",
				Required:    true,
			},
			"symbol_name": {
				Description: "The name of the function or method to write tests for.",
				Type:        "string",
				Required:    true,
			},
		},
	},
	OutputConfig: &OutputConfig{
		OutputName: "report",
		// Schema: TestWriterReportSchema, // Zod schema equivalent
	},

	ProcessOutput: func(output interface{}) string {
		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Sprintf("Error marshaling output: %v", err)
		}
		return string(jsonBytes)
	},

	ModelConfig: ModelConfig{
		Model:          config.DEFAULT_GEMINI_MODEL,
		Temperature:    0.1,
		TopP:           0.95,
		ThinkingBudget: -1,
	},

	RunConfig: RunConfig{
		MaxTimeMinutes: 5,
		MaxTurns:       15,
	},

	ToolConfig: &ToolConfig{
		Tools: []string{
			types.READ_FILE_TOOL_NAME,
			types.WRITE_FILE_TOOL_NAME,
			types.RUN_TESTS_TOOL_NAME,
			types.FIND_REFERENCES_TOOL_NAME, // Might be useful for context
		},
	},

	PromptConfig: PromptConfig{
		Query:        "", // Will be populated in init()
		SystemPrompt: "", // Will be populated in init()
	},
}

var testWriterPrompts map[string]string

func init() {
	var err error
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path for test_writer")
	}
	promptsFilePath := filepath.Join(filepath.Dir(filename), "test_writer_prompts.md")
	testWriterPrompts, err = loadPromptsFromFile(promptsFilePath)
	if err != nil {
		panic(fmt.Sprintf("failed to load test writer prompts from %s: %v", promptsFilePath, err))
	}

	// Update the TestWriterAgent's prompts from the loaded map
	TestWriterAgent.Description = testWriterPrompts["Description"]
	TestWriterAgent.PromptConfig.Query = testWriterPrompts["Query"]
	TestWriterAgent.PromptConfig.SystemPrompt = testWriterPrompts["System Prompt"]
}
