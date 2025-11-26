package prompts

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/types"
)

//go:embed core/*.md agents/*.md
var coreFS embed.FS

var corePrompts = make(map[string]string)

// LoadPromptsFromFile parses a markdown file from a given path.
func LoadPromptsFromFile(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open prompt file %s: %w", filePath, err)
	}
	defer file.Close()
	return LoadPrompts(file)
}

// LoadPrompts parses a markdown stream with "## Header" sections into a map.
func LoadPrompts(r io.Reader) (map[string]string, error) {
	scanner := bufio.NewScanner(r)
	prompts := make(map[string]string)
	var currentHeader string
	var currentContent strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") {
			// If we have a current header, save the content we've collected
			if currentHeader != "" {
				prompts[currentHeader] = strings.TrimSpace(currentContent.String())
			}
			// Start a new prompt section
			currentHeader = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			currentContent.Reset()
		} else if currentHeader != "" {
			// Append line to the current prompt's content
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
		}
	}

	// Save the last prompt section
	if currentHeader != "" {
		prompts[currentHeader] = strings.TrimSpace(currentContent.String())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return prompts, nil
}

func init() {
	// Load core prompts
	loadPromptsFromDir("core")
	// Load agent prompts
	loadPromptsFromDir("agents")
}

func loadPromptsFromDir(dir string) {
	files, err := coreFS.ReadDir(dir)
	if err != nil {
		panic(fmt.Sprintf("failed to read embedded prompts directory %s: %v", dir, err))
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			filePath := filepath.Join(dir, file.Name())
			content, err := coreFS.ReadFile(filePath)
			if err != nil {
				panic(fmt.Sprintf("failed to read embedded prompt file %s: %v", filePath, err))
			}
			promptName := strings.TrimSuffix(file.Name(), ".md")
			corePrompts[promptName] = string(content)
		}
	}
}

// GetCoreSystemPrompt constructs the system prompt from the loaded markdown files.
func GetCoreSystemPrompt(toolRegistry types.ToolRegistryInterface, config types.Config, contextContent string) (string, error) {
	var sb strings.Builder

	// Prepend context from GOAIAGENT.md files
	if contextContent != "" {
		sb.WriteString(contextContent)
		sb.WriteString("\n\n---\n\n") // Separator
	}

	// Add preamble and core mandates
	sb.WriteString(corePrompts["preamble"])
	sb.WriteString("\n\n")
	sb.WriteString(corePrompts["core_mandates"])
	sb.WriteString("\n\n")

	// Conditionally add primary workflows
	toolNames := toolRegistry.GetAllToolNames()
	hasCI := contains(toolNames, types.CODEBASE_INVESTIGATOR_TOOL_NAME)
	hasTodos := contains(toolNames, types.WRITE_TODOS_TOOL_NAME)

	if hasCI && hasTodos {
		sb.WriteString(corePrompts["primary_workflows_prefix_ci_todo"])
	} else if hasCI {
		sb.WriteString(corePrompts["primary_workflows_prefix_ci"])
	} else if hasTodos {
		sb.WriteString(corePrompts["primary_workflows_todo"])
	} else {
		sb.WriteString(corePrompts["primary_workflows_prefix"])
	}
	sb.WriteString("\n\n")
	sb.WriteString(corePrompts["primary_workflows_suffix"])
	sb.WriteString("\n\n")

	// Add other sections
	sb.WriteString(corePrompts["operational_guidelines"])
	sb.WriteString("\n\n")
	sb.WriteString(corePrompts["sandbox"])
	sb.WriteString("\n\n")
	sb.WriteString(corePrompts["git"])
	sb.WriteString("\n\n")
	sb.WriteString(corePrompts["final_reminder"])

	// Replace placeholders
	prompt := sb.String()
	prompt = strings.ReplaceAll(prompt, "${GREP_TOOL_NAME}", types.GREP_TOOL_NAME)
	prompt = strings.ReplaceAll(prompt, "${GLOB_TOOL_NAME}", types.GLOB_TOOL_NAME)
	prompt = strings.ReplaceAll(prompt, "${READ_FILE_TOOL_NAME}", types.READ_FILE_TOOL_NAME)
	prompt = strings.ReplaceAll(prompt, "${codebase_investigator}", types.CODEBASE_INVESTIGATOR_TOOL_NAME)
	prompt = strings.ReplaceAll(prompt, "${write_todos}", types.WRITE_TODOS_TOOL_NAME)
	prompt = strings.ReplaceAll(prompt, "${smart_edit}", types.SMART_EDIT_TOOL_NAME)
	prompt = strings.ReplaceAll(prompt, "${write_file}", types.WRITE_FILE_TOOL_NAME)
	prompt = strings.ReplaceAll(prompt, "${execute_command}", types.EXECUTE_COMMAND_TOOL_NAME)
	prompt = strings.ReplaceAll(prompt, "${memory}", types.MEMORY_TOOL_NAME)
	prompt = strings.ReplaceAll(prompt, "${read_file}", types.READ_FILE_TOOL_NAME)

	return prompt, nil
}

// GetPrompt retrieves a specific prompt by its file name (without extension).
func GetPrompt(name string) (string, bool) {
	prompt, exists := corePrompts[name]
	return prompt, exists
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}