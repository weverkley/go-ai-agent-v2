package agents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath" // Added
	"runtime"       // Added
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// ... (rest of the file)

func loadPromptsFromFile(filePath string) (map[string]string, error) {
	content, err := os.ReadFile(filePath)
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

var codebaseInvestigatorPrompts map[string]string

func init() {
	var err error
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path")
	}
	// The path to the prompts file is relative to the current file.
	promptsFilePath := filepath.Join(filepath.Dir(filename), "codebase_investigator_prompts.md")
	codebaseInvestigatorPrompts, err = loadPromptsFromFile(promptsFilePath)
	if err != nil {
		panic(fmt.Sprintf("failed to load codebase investigator prompts from %s: %v", promptsFilePath, err))
	}
}

// CodebaseInvestigatorAgent defines the Codebase Investigator subagent.
var CodebaseInvestigatorAgent = AgentDefinition{
	Name:        types.CODEBASE_INVESTIGATOR_TOOL_NAME,
	DisplayName: "Codebase Investigator Agent",
	Description: codebaseInvestigatorPrompts["Description"],
	InputConfig: InputConfig{
		Inputs: map[string]InputParameter{
			"objective": {
				Description: codebaseInvestigatorPrompts["Objective Description"],
				Type:        "string",
				Required:    true,
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
		Model:          config.DEFAULT_GEMINI_MODEL, // Assuming DEFAULT_GEMINI_MODEL is defined in config
		Temperature:    0.1,
		TopP:           0.95,
		ThinkingBudget: -1,
	},

	RunConfig: RunConfig{
		MaxTimeMinutes: 5,
		MaxTurns:       15,
	},

	ToolConfig: &ToolConfig{
		// Grant access only to read-only tools.
		Tools: []string{types.LS_TOOL_NAME, types.READ_FILE_TOOL_NAME, types.GLOB_TOOL_NAME, types.GREP_TOOL_NAME, types.FIND_UNUSED_CODE_TOOL_NAME},
	},

	PromptConfig: PromptConfig{
		Query:        codebaseInvestigatorPrompts["Query"],
		SystemPrompt: codebaseInvestigatorPrompts["System Prompt"],
	},
}
