package agents

import (
	"encoding/json"
	"fmt"
	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/tools"
	"os"
	"strings"
)

// CodebaseInvestigationReportSchema represents the output schema for the Codebase Investigator Agent.
type CodebaseInvestigationReportSchema struct {
	SummaryOfFindings string `json:"SummaryOfFindings"`
	ExplorationTrace  []string `json:"ExplorationTrace"`
	RelevantLocations []struct {
		FilePath   string   `json:"FilePath"`
		Reasoning  string   `json:"Reasoning"`
		KeySymbols []string `json:"KeySymbols"`
	} `json:"RelevantLocations"`
}

// AgentDefinition represents the definition of an AI agent.
type AgentDefinition struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	InputConfig struct {
		Inputs map[string]struct {
			Description string `json:"description"`
			Type        string `json:"type"`
			Required    bool   `json:"required"`
		} `json:"inputs"`
	} `json:"inputConfig"`
	OutputConfig struct {
		OutputName  string `json:"outputName"`
		Description string `json:"description"`
		Schema      interface{} `json:"schema"` // This will hold CodebaseInvestigationReportSchema
	} `json:"outputConfig"`
	ModelConfig struct {
		Model        string  `json:"model"`
		Temp         float64 `json:"temp"`
		TopP         float64 `json:"top_p"`
		ThinkingBudget int     `json:"thinkingBudget"`
	} `json:"modelConfig"`
	RunConfig struct {
		MaxTimeMinutes int `json:"max_time_minutes"`
		MaxTurns       int `json:"max_turns"`
	} `json:"runConfig"`
	ToolConfig struct {
		Tools []string `json:"tools"`
	} `json:"toolConfig"`
	PromptConfig struct {
		Query        string `json:"query"`
		SystemPrompt string `json:"systemPrompt"`
	} `json:"promptConfig"`
	ProcessOutput func(output interface{}) (string, error) `json:"-"` // Not marshaled to JSON
}

// loadPromptsFromMarkdown reads a Markdown file and extracts content under H2 headings.
func loadPromptsFromMarkdown(filePath string) (map[string]string, error) {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read markdown file: %w", err)
	}
	content := string(contentBytes)

	prompts := make(map[string]string)
	lines := strings.Split(content, "\n")

	currentHeading := ""
	currentContent := strings.Builder{}

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			if currentHeading != "" {
				prompts[currentHeading] = strings.TrimSpace(currentContent.String())
			}
			currentHeading = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			currentContent.Reset()
		} else if currentHeading != "" {
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
		}
	}

	// Add the last section
	if currentHeading != "" {
		prompts[currentHeading] = strings.TrimSpace(currentContent.String())
	}

	return prompts, nil
}

// CodebaseInvestigatorAgent defines the Codebase Investigator Agent.
var CodebaseInvestigatorAgent = AgentDefinition{
	Name:        "codebase_investigator",
	DisplayName: "Codebase Investigator Agent",
	Description: ``, // This will be manually filled
	InputConfig: struct {
		Inputs map[string]struct {
			Description string `json:"description"`
			Type        string `json:"type"`
			Required    bool   `json:"required"`
		} `json:"inputs"`
	}{
		Inputs: map[string]struct {
			Description string `json:"description"`
			Type        string `json:"type"`
			Required    bool   `json:"required"`
		}{
			"objective": {
				Description: ``, // This will be manually filled
				Type:     "string",
				Required: true,
			},
		},
	},
	OutputConfig: struct {
		OutputName  string `json:"outputName"`
		Description string `json:"description"`
		Schema      interface{} `json:"schema"`
	}{
		OutputName:  "report",
		Description: "The final investigation report as a JSON object.",
		Schema:      CodebaseInvestigationReportSchema{}, // Assign the schema struct
	},
	ModelConfig: struct {
		Model        string  `json:"model"`
		Temp         float64 `json:"temp"`
		TopP         float64 `json:"top_p"`
		ThinkingBudget int     `json:"thinkingBudget"`
	}{
		Model:        config.DEFAULT_GEMINI_MODEL,
		Temp:         0.1,
		TopP:         0.95,
		ThinkingBudget: -1,
	},
	RunConfig: struct {
		MaxTimeMinutes int `json:"max_time_minutes"`
		MaxTurns       int `json:"max_turns"`
	}{
		MaxTimeMinutes: 5,
		MaxTurns:       15,
	},
	ToolConfig: struct {
		Tools []string `json:"tools"`
	}{
		Tools: []string{tools.LS_TOOL_NAME, tools.READ_FILE_TOOL_NAME, tools.GLOB_TOOL_NAME, tools.GREP_TOOL_NAME},
	},
	PromptConfig: struct {
		Query        string `json:"query"`
		SystemPrompt string `json:"systemPrompt"`
	}{
		Query: ``, // This will be manually filled
		SystemPrompt: ``, // This will be manually filled
	},
	ProcessOutput: func(output interface{}) (string, error) {
		bytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal output: %w", err)
		}
		return string(bytes), nil
	},
}

func init() {
	prompts, err := loadPromptsFromMarkdown("go-cli/pkg/core/agents/codebase_investigator_prompts.md")
	if err != nil {
		panic(err) // Or handle error more gracefully
	}

	CodebaseInvestigatorAgent.Description = prompts["Description"]
	CodebaseInvestigatorAgent.InputConfig.Inputs["objective"].Description = prompts["Objective Description"]
	CodebaseInvestigatorAgent.PromptConfig.Query = prompts["Query"]
	CodebaseInvestigatorAgent.PromptConfig.SystemPrompt = prompts["System Prompt"]
}

