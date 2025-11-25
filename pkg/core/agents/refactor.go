package agents

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// RefactorAgent defines the Refactor subagent.
var RefactorAgent = AgentDefinition{
	Name:        types.REFACTOR_AGENT_NAME,
	DisplayName: types.REFACTOR_AGENT_DISPLAY_NAME,
	Description: "", // Will be populated in init()
	InputConfig: InputConfig{
		Inputs: map[string]InputParameter{
			"target_path": {
				Description: "The file path or directory containing the code to be refactored.",
				Type:        "string",
				Required:    true,
			},
			"refactoring_goal": {
				Description: "A clear description of the refactoring goal (e.g., 'Simplify the processOrder function by extracting helper methods').",
				Type:        "string",
				Required:    true,
			},
		},
	},
	OutputConfig: &OutputConfig{
		OutputName: "report",
		// Schema: RefactorReportSchema, // Zod schema equivalent
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
		MaxTimeMinutes: 15, // Refactoring can take longer
		MaxTurns:       30, // Refactoring can take more turns
	},

	ToolConfig: &ToolConfig{
		Tools: []string{
			types.READ_FILE_TOOL_NAME,
			types.WRITE_FILE_TOOL_NAME,
			types.RUN_TESTS_TOOL_NAME,
			types.CODEBASE_INVESTIGATOR_TOOL_NAME, // Assuming this is defined
			types.EXTRACT_FUNCTION_TOOL_NAME,
			types.RENAME_SYMBOL_TOOL_NAME,
			types.SMART_EDIT_TOOL_NAME,
			types.GIT_COMMIT_TOOL_NAME, // For committing atomic changes if desired by agent
		},
	},

	PromptConfig: PromptConfig{
		Query:        "", // Will be populated in init()
		SystemPrompt: "", // Will be populated in init()
	},
}

var refactorAgentPrompts map[string]string

func init() {
	var err error
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path for refactor_agent")
	}
	promptsFilePath := filepath.Join(filepath.Dir(filename), "refactor_prompts.md")
	refactorAgentPrompts, err = loadPromptsFromFile(promptsFilePath)
	if err != nil {
		panic(fmt.Sprintf("failed to load refactor agent prompts from %s: %v", promptsFilePath, err))
	}

	// Update the RefactorAgent's prompts from the loaded map
	RefactorAgent.Description = refactorAgentPrompts["Description"]
	RefactorAgent.PromptConfig.Query = refactorAgentPrompts["Query"]
	RefactorAgent.PromptConfig.SystemPrompt = refactorAgentPrompts["System Prompt"]
}
