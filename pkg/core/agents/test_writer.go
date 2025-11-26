package agents

import (
	"encoding/json"
	"fmt"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/prompts"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// TestWriterAgent defines the Test Writer subagent.
var TestWriterAgent = func() AgentDefinition {
	agentPromptsRaw, ok := prompts.GetPrompt("test_writer")
	if !ok {
		panic("could not find test_writer prompts")
	}
	agentPrompts, err := prompts.LoadPrompts(strings.NewReader(agentPromptsRaw))
	if err != nil {
		panic(fmt.Sprintf("failed to parse test_writer prompts: %v", err))
	}

	return AgentDefinition{
		Name:        types.TEST_WRITER_AGENT_NAME,
		DisplayName: types.TEST_WRITER_AGENT_DISPLAY_NAME,
		Description: agentPrompts["Description"],
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
			Query:        agentPrompts["Query"],
			SystemPrompt: agentPrompts["System Prompt"],
		},
	}
}()
