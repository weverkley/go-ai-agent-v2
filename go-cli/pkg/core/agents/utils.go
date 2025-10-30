package agents

import (
	"fmt"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/services"
)

// templateString replaces placeholders in a string with values from inputs.
func templateString(template string, inputs AgentInputs) string {
	result := template
	for key, value := range inputs {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// logAgentStart is a placeholder for telemetry logging.
func logAgentStart(runtimeContext interface{}, event interface{}) {
	// fmt.Printf("Telemetry: Agent Start - %v, %v\n", runtimeContext, event)
}

// logAgentFinish is a placeholder for telemetry logging.
func logAgentFinish(runtimeContext interface{}, event interface{}) {
	// fmt.Printf("Telemetry: Agent Finish - %v, %v\n", runtimeContext, event)
}

// AgentStartEvent is a placeholder for telemetry event.
type AgentStartEvent struct {
	AgentID   string
	AgentName string
}

// AgentFinishEvent is a placeholder for telemetry event.
type AgentFinishEvent struct {
	AgentID       string
	AgentName     string
	DurationMs    int64
	TurnCounter   int
	TerminateReason AgentTerminateMode
}

// getDirectoryContextString generates a string describing the current workspace directories and their structures.
func getDirectoryContextString(cfg *config.Config) (string, error) {
	workspaceContext := cfg.GetWorkspaceContext()
	workspaceDirectories := workspaceContext.GetDirectories()

	folderStructures := make([]string, len(workspaceDirectories))
	for i, dir := range workspaceDirectories {
		fs, err := getFolderStructure(dir, nil, cfg.GetFileService().(*config.dummyFileService)) // Cast to concrete type for now
		if err != nil {
			return "", fmt.Errorf("failed to get folder structure for %s: %w", dir, err)
		}
		folderStructures[i] = fs
	}

	folderStructure := strings.Join(folderStructures, "\n")

	var workingDirPreamble string
	if len(workspaceDirectories) == 1 {
		workingDirPreamble = fmt.Sprintf("I'm currently working in the directory: %s", workspaceDirectories[0])
	} else {
		dirList := make([]string, len(workspaceDirectories))
		for i, dir := range workspaceDirectories {
			dirList[i] = fmt.Sprintf("  - %s", dir)
		}
		workingDirPreamble = fmt.Sprintf("I'm currently working in the following directories:\n%s", strings.Join(dirList, "\n"))
	}

	return fmt.Sprintf("%s\nHere is the folder structure of the current working directories:\n\n%s", workingDirPreamble, folderStructure), nil
}