package agents

import (
	"encoding/json"
	"fmt"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/prompts"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// ... (rest of the file)

// CodebaseInvestigatorAgent defines the Codebase Investigator subagent.
var CodebaseInvestigatorAgent = func() AgentDefinition {
	agentPromptsRaw, ok := prompts.GetPrompt("codebase_investigator")
	if !ok {
		panic("could not find codebase_investigator prompts")
	}
	agentPrompts, err := prompts.LoadPrompts(strings.NewReader(agentPromptsRaw))
	if err != nil {
		panic(fmt.Sprintf("failed to parse codebase_investigator prompts: %v", err))
	}

	return AgentDefinition{
		Name:        types.CODEBASE_INVESTIGATOR_TOOL_NAME,
		DisplayName: "Codebase Investigator Agent",
		Description: agentPrompts["Description"],
		InputConfig: InputConfig{
			Inputs: map[string]InputParameter{
				"objective": {
					Description: agentPrompts["Objective Description"],
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
			Tools: []string{types.LS_TOOL_NAME, types.READ_FILE_TOOL_NAME, types.GLOB_TOOL_NAME, types.GREP_TOOL_NAME, types.FIND_UNUSED_CODE_TOOL_NAME},
		},
		PromptConfig: PromptConfig{
			Query:        agentPrompts["Query"],
			SystemPrompt: agentPrompts["System Prompt"],
		},
	}
}()
