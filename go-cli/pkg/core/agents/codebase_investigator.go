package agents

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/types"
)

func loadPromptsFromFile(filePath string) (map[string]string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	prompts := make(map[string]string)
	sections := strings.Split(string(content), "## ")

	for _, section := range sections {
		if strings.TrimSpace(section) == "" {
			continue
		}
		parts := strings.SplitN(section, "\n", 2)
		if len(parts) < 2 {
			continue
		}
		header := strings.TrimSpace(parts[0])
		body := strings.TrimSpace(parts[1])
		prompts[header] = body
	}

	return prompts, nil
}

var prompts, _ = loadPromptsFromFile("pkg/core/agents/codebase_investigator_prompts.md")

// CodebaseInvestigatorAgent defines the Codebase Investigator subagent.
var CodebaseInvestigatorAgent = AgentDefinition{
	Name:        "codebase_investigator",
	DisplayName: "Codebase Investigator Agent",
	Description: prompts["Description"],
	InputConfig: InputConfig{
		Inputs: map[string]InputParameter{
			"objective": {
				Description: prompts["Objective Description"],
				Type:     "string",
				Required: true,
			},
		},
	},
	OutputConfig: &OutputConfig{
		OutputName: "report",
		// Schema: CodebaseInvestigationReportSchema, // Zod schema equivalent
	},

	// The 'output' parameter is now strongly typed as CodebaseInvestigationReportSchema
	ProcessOutput: func(output interface{}) string {
		// In Go, we'll assume the output is already a struct matching the schema
		// and we'll marshal it to JSON.
		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Sprintf("Error marshaling output: %v", err)
		}
		return string(jsonBytes)
	},

	ModelConfig: ModelConfig{
		Model:        config.DEFAULT_GEMINI_MODEL, // Assuming DEFAULT_GEMINI_MODEL is defined in config
		Temperature:  0.1,
		TopP:         0.95,
		ThinkingBudget: -1,
	},

	RunConfig: RunConfig{
		MaxTimeMinutes: 5,
		MaxTurns:       15,
	},

	ToolConfig: &ToolConfig{
		// Grant access only to read-only tools.
		Tools: []string{types.LS_TOOL_NAME, types.READ_FILE_TOOL_NAME, types.GLOB_TOOL_NAME, types.GREP_TOOL_NAME},
	},

	PromptConfig: PromptConfig{
		Query: prompts["Query"],
		SystemPrompt: prompts["System Prompt"],
	},
}
