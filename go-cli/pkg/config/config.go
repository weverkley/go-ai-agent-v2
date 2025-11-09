package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/telemetry"
	"go-ai-agent-v2/go-cli/pkg/types"
)

// SettingScope defines the scope of a setting.
type SettingScope string

const (
	SettingScopeUser      SettingScope = "user"
	SettingScopeWorkspace SettingScope = "workspace"
)

// WorkspaceContext defines an interface for accessing workspace-related information.
type WorkspaceContext interface {
	GetDirectories() []string
}



// OutputSettings represents the output settings.
type OutputSettings struct {
	Format string `json:"format,omitempty"`
}

// ConfigParameters represents the parameters for creating a new Config.
type ConfigParameters struct {
	SessionID      string
	EmbeddingModel string
	TargetDir      string
	DebugMode      bool
	ModelName          string
	McpServers     map[string]types.MCPServerConfig
	ApprovalMode   types.ApprovalMode // Use ApprovalMode from types package
	Telemetry      *types.TelemetrySettings
	Output         *OutputSettings
	CodebaseInvestigator *types.CodebaseInvestigatorSettings
	ToolRegistry *types.ToolRegistry // Changed to exported
	ToolDiscoveryCommand string
	ToolCallCommand      string
}

// Config represents the application configuration.
type Config struct {
	sessionID      string
	embeddingModel string
	targetDir      string
	debugMode      bool
	modelName          string
	mcpServers     map[string]types.MCPServerConfig
	approvalMode   types.ApprovalMode // Use ApprovalMode from types package
	telemetry      *types.TelemetrySettings
	output         *OutputSettings
	codebaseInvestigatorSettings *types.CodebaseInvestigatorSettings
	ToolRegistry *types.ToolRegistry // Changed to exported
	toolDiscoveryCommand string
	toolCallCommand      string
	telemetryLogger telemetry.TelemetryLogger
	fileService types.FileService // Add FileService to Config struct
}

func NewConfig(params *ConfigParameters) *Config {
	cfg := &Config{
		sessionID:      params.SessionID,
		embeddingModel: params.EmbeddingModel,
		targetDir:      params.TargetDir,
		debugMode:      params.DebugMode,
		modelName:      params.ModelName,
		mcpServers:     params.McpServers,
		approvalMode:   params.ApprovalMode,
		telemetry:      params.Telemetry,
		output:         params.Output,
		codebaseInvestigatorSettings: params.CodebaseInvestigator,
		ToolRegistry: params.ToolRegistry,
		toolDiscoveryCommand: params.ToolDiscoveryCommand,
		toolCallCommand:      params.ToolCallCommand,
	}
	cfg.telemetryLogger = telemetry.NewTelemetryLogger(params.Telemetry) // Initialize here
	// fileService will be set by SetConfiguredFileService later

	return cfg
}

// SetConfiguredFileService sets the FileService for the Config.
func (c *Config) SetConfiguredFileService(fs types.FileService) {
	c.fileService = fs
}

// GetToolDiscoveryCommand returns the tool discovery command.
func (c *Config) GetToolDiscoveryCommand() string {
	return c.toolDiscoveryCommand
}

// GetWorkspaceContext returns the workspace context.
func (c *Config) GetWorkspaceContext() WorkspaceContext {
	return &realWorkspaceContext{projectRoot: c.targetDir}
}

// realWorkspaceContext is a real implementation of WorkspaceContext.
type realWorkspaceContext struct {
	projectRoot string
}

func (rwc *realWorkspaceContext) GetDirectories() []string {
	return []string{rwc.projectRoot}
}
// Model returns the configured model name.
func (c *Config) Model() string {
	return c.modelName
}

// GetCodebaseInvestigatorSettings returns the Codebase Investigator settings.
func (c *Config) GetCodebaseInvestigatorSettings() *types.CodebaseInvestigatorSettings {
	return c.codebaseInvestigatorSettings
}

// GetDebugMode returns true if debug mode is enabled.
func (c *Config) GetDebugMode() bool {
	return c.debugMode
}

// GetToolRegistry returns the global tool registry.
func (c *Config) GetToolRegistry() *types.ToolRegistry {
	return c.ToolRegistry
}

// GetTelemetryLogger returns the initialized telemetry logger.
func (c *Config) GetTelemetryLogger() telemetry.TelemetryLogger {
	return c.telemetryLogger
}

// GetGeminiDir returns the path to the .gemini directory within the target directory.
func (c *Config) GetGeminiDir() string {
	return filepath.Join(c.targetDir, ".gemini")
}

// Get returns the value of a configuration setting by key.
func (c *Config) Get(key string) (interface{}, bool) {
	switch key {
	case "model":
		return c.modelName, true
	case "debugMode":
		return c.debugMode, true
	case "sessionID":
		return c.sessionID, true
	case "embeddingModel":
		return c.embeddingModel, true
	case "targetDir":
		return c.targetDir, true
	case "approvalMode":
		return c.approvalMode, true
	case "toolDiscoveryCommand":
		return c.toolDiscoveryCommand, true
	case "toolCallCommand":
		return c.toolCallCommand, true
	case "telemetryEnabled":
		if c.telemetry != nil {
			return c.telemetry.Enabled, true
		}
		return false, true // Default to false if telemetry settings are nil
	// Add more cases for other settings as needed
	default:
		return nil, false
	}
}

const (
	MAX_ITEMS            = 200
	TRUNCATION_INDICATOR = "..."
)

var (
	DEFAULT_IGNORED_FOLDERS = map[string]bool{
		"node_modules": true,
		".git":         true,
		"dist":         true,
	}
	DEFAULT_FILE_FILTERING_OPTIONS = types.FileFilteringOptions{
		RespectGitIgnore:  boolPtr(true),
		RespectGeminiIgnore: boolPtr(true),
	}
)

// boolPtr returns a pointer to a boolean.
func boolPtr(b bool) *bool {
	return &b
}

// GetDirectoryContextString generates a string describing the current workspace directories and their structures.
func (c *Config) GetDirectoryContextString() (string, error) {
	workspaceContext := c.GetWorkspaceContext()
	workspaceDirectories := workspaceContext.GetDirectories()

	var folderStructures []string
	for _, dir := range workspaceDirectories {
		structure, err := c._getFolderStructure(dir, &types.FolderStructureOptions{}, c.fileService)
		if err != nil {
			return "", err
		}
		folderStructures = append(folderStructures, structure)
	}

	folderStructure := strings.Join(folderStructures, "\n")

	var workingDirPreamble string
	if len(workspaceDirectories) == 1 {
		workingDirPreamble = fmt.Sprintf("I'm currently working in the directory: %s", workspaceDirectories[0])
	} else {
		var dirList []string
		for _, dir := range workspaceDirectories {
			dirList = append(dirList, fmt.Sprintf("  - %s", dir))
		}
		workingDirPreamble = fmt.Sprintf("I'm currently working in the following directories:\n%s", strings.Join(dirList, "\n"))
	}

	return fmt.Sprintf("%s\n\n%s%c\n%s", workingDirPreamble, folderStructure, filepath.Separator, strings.Join(folderStructures, "\n")),
		nil
}

// _getFolderStructure generates a string representation of a directory's structure.
func (c *Config) _getFolderStructure(directory string, options *types.FolderStructureOptions, fileService types.FileService) (string, error) {
	resolvedPath, err := filepath.Abs(directory)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	mergedOptions := types.FolderStructureOptions{
		MaxItems:           intPtr(MAX_ITEMS),
		IgnoredFolders:     &[]string{}, // Will be merged with DEFAULT_IGNORED_FOLDERS
		FileIncludePattern: nil,
		FileFilteringOptions: &DEFAULT_FILE_FILTERING_OPTIONS,
	}

	if options != nil {
		if options.MaxItems != nil {
			mergedOptions.MaxItems = options.MaxItems
		}
		if options.IgnoredFolders != nil {
			mergedOptions.IgnoredFolders = options.IgnoredFolders
		}
		if options.FileIncludePattern != nil {
			mergedOptions.FileIncludePattern = options.FileIncludePattern
		}
		if options.FileFilteringOptions != nil {
			mergedOptions.FileFilteringOptions = options.FileFilteringOptions
		}
	}

	// Merge default ignored folders
	defaultIgnoredMap := make(map[string]bool)
	for k := range DEFAULT_IGNORED_FOLDERS {
		defaultIgnoredMap[k] = true
	}
	if mergedOptions.IgnoredFolders != nil {
		for _, folder := range *mergedOptions.IgnoredFolders {
			defaultIgnoredMap[folder] = true
		}
	}
	mergedOptions.IgnoredFolders = &[]string{}
	for k := range defaultIgnoredMap {
		*mergedOptions.IgnoredFolders = append(*mergedOptions.IgnoredFolders, k)
	}


	structureRoot, err := c.readFullStructure(resolvedPath, &mergedOptions, fileService)
	if err != nil {
		return "", err
	}
	if structureRoot == nil {
		return fmt.Sprintf("Error: Could not read directory \"%s\". Check path and permissions.", resolvedPath), nil
	}

	structureLines := []string{}
	c.formatStructure(structureRoot, "", true, true, &structureLines)

	summary := fmt.Sprintf("Showing up to %d items (files + folders).", *mergedOptions.MaxItems)

	if c.isTruncated(structureRoot) {
		summary += fmt.Sprintf(" Folders or files indicated with %s contain more items not shown, were ignored, or the display limit (%d items) was reached.", TRUNCATION_INDICATOR, *mergedOptions.MaxItems)
	}

	return fmt.Sprintf("%s\n\n%s%c\n%s", summary, resolvedPath, filepath.Separator, strings.Join(structureLines, "\n")),
		nil
}

// intPtr returns a pointer to an int.
func intPtr(i int) *int {
	return &i
}

// readFullStructure reads the directory structure using BFS, respecting maxItems.
func (c *Config) readFullStructure(rootPath string, options *types.FolderStructureOptions, fileService types.FileService) (*types.FullFolderInfo, error) {
	rootName := filepath.Base(rootPath)
	rootNode := &types.FullFolderInfo{
		Name:       rootName,
		Path:       rootPath,
		Files:      []string{},
		SubFolders: []types.FullFolderInfo{},
	}

	queue := []struct {
		folderInfo  *types.FullFolderInfo
		currentPath string
	}{
		{folderInfo: rootNode, currentPath: rootPath},
	}
	currentItemCount := 0
	processedPaths := make(map[string]bool)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		folderInfo := current.folderInfo
		currentPath := current.currentPath

		if processedPaths[currentPath] {
			continue
		}
		processedPaths[currentPath] = true

		if currentItemCount >= *options.MaxItems {
			continue
		}

		entries, err := os.ReadDir(currentPath)
		if err != nil {
			if os.IsPermission(err) || os.IsNotExist(err) {
				// debugLogger.Warn(fmt.Sprintf("Warning: Could not read directory %s: %v", currentPath, err))
				// If root directory itself not found
				if currentPath == rootPath && os.IsNotExist(err) {
					return nil, nil 
				}
				continue
			}
			return nil, fmt.Errorf("failed to read directory %s: %w", currentPath, err)
		}

		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Name() < entries[j].Name()
		})

		filesInCurrentDir := []string{}
		subFoldersInCurrentDir := []types.FullFolderInfo{}

		// Process files first
		for _, entry := range entries {
			if entry.Type().IsRegular() {
				if currentItemCount >= *options.MaxItems {
					folderInfo.HasMoreFiles = true
					break
				}
				fileName := entry.Name()
				

				isIgnored := false
				if fileService != nil {
					isIgnored = fileService.ShouldIgnoreFile(filepath.Join(currentPath, fileName), *options.FileFilteringOptions)
				}

				if isIgnored {
					continue
				}

				if options.FileIncludePattern == nil || regexp.MustCompile(*options.FileIncludePattern).MatchString(fileName) {
					filesInCurrentDir = append(filesInCurrentDir, fileName)
					currentItemCount++
					folderInfo.TotalFiles++
					folderInfo.TotalChildren++
				}
			}
		}
		folderInfo.Files = filesInCurrentDir

		// Then process directories and queue them
		for _, entry := range entries {
			if entry.IsDir() {
				if currentItemCount >= *options.MaxItems {
					folderInfo.HasMoreSubfolders = true
					break
				}

				subFolderName := entry.Name()
				subFolderPath := filepath.Join(currentPath, subFolderName)

				isIgnored := false
				if containsString(*options.IgnoredFolders, subFolderName) {
					isIgnored = true
				}
				if fileService != nil {
					fileServiceIgnored := fileService.ShouldIgnoreFile(subFolderPath, *options.FileFilteringOptions)
					if fileServiceIgnored {
						isIgnored = true
					}
				}

				if isIgnored {
					ignoredSubFolder := types.FullFolderInfo{
						Name:        subFolderName,
						Path:        subFolderPath,
						IsIgnored:   true,
						Files:       []string{},
						SubFolders:  []types.FullFolderInfo{},
						TotalChildren: 0,
						TotalFiles:  0,
					}
					subFoldersInCurrentDir = append(subFoldersInCurrentDir, ignoredSubFolder)
					currentItemCount++
					folderInfo.TotalChildren++
					continue
				}

				subFolderNode := &types.FullFolderInfo{
					Name:       subFolderName,
					Path:       subFolderPath,
					Files:      []string{},
					SubFolders: []types.FullFolderInfo{},
				}
				subFoldersInCurrentDir = append(subFoldersInCurrentDir, *subFolderNode)
				currentItemCount++
				folderInfo.TotalChildren++

				queue = append(queue, struct {
					folderInfo  *types.FullFolderInfo
					currentPath string
				}{folderInfo: subFolderNode, currentPath: subFolderPath})
			}
		}
		folderInfo.SubFolders = subFoldersInCurrentDir
	}

	return rootNode, nil
}

// formatStructure formats the FullFolderInfo into a string.
func (c *Config) formatStructure(
	node *types.FullFolderInfo,
	currentIndent string,
	isLastChildOfParent bool,
	isProcessingRootNode bool,
	builder *[]string,
) {
	connector := "├───"
	if isLastChildOfParent {
		connector = "└───"
	}

	if !isProcessingRootNode || node.IsIgnored {
		indicator := ""
		if node.IsIgnored {
			indicator = TRUNCATION_INDICATOR
		}
		*builder = append(*builder, fmt.Sprintf("%s%s%s%c%s", currentIndent, connector, node.Name, filepath.Separator, indicator))
	}

	indentForChildren := currentIndent
	if !isProcessingRootNode {
		if isLastChildOfParent {
			indentForChildren += "    "
		} else {
			indentForChildren += "│   "
		}
	}

	// Render files
	fileCount := len(node.Files)
	for i, file := range node.Files {
		isLastFileAmongSiblings := i == fileCount-1 && len(node.SubFolders) == 0 && !node.HasMoreSubfolders
		fileConnector := "├───"
		if isLastFileAmongSiblings {
			fileConnector = "└───"
		}
		*builder = append(*builder, fmt.Sprintf("%s%s%s", indentForChildren, fileConnector, file))
	}
	if node.HasMoreFiles {
		isLastIndicatorAmongSiblings := len(node.SubFolders) == 0 && !node.HasMoreSubfolders
		fileConnector := "├───"
		if isLastIndicatorAmongSiblings {
			fileConnector = "└───"
		}
		*builder = append(*builder, fmt.Sprintf("%s%s%s", indentForChildren, fileConnector, TRUNCATION_INDICATOR))
	}

	// Render subfolders
	subFolderCount := len(node.SubFolders)
	for i := range node.SubFolders {
		isLastSubfolderAmongSiblings := i == subFolderCount-1 && !node.HasMoreSubfolders
		c.formatStructure(&node.SubFolders[i], indentForChildren, isLastSubfolderAmongSiblings, false, builder)
	}
	if node.HasMoreSubfolders {
		*builder = append(*builder, fmt.Sprintf("%s└───%s", indentForChildren, TRUNCATION_INDICATOR))
	}
}

// isTruncated checks if any part of the folder structure was truncated or ignored.
func (c *Config) isTruncated(node *types.FullFolderInfo) bool {
	if node.HasMoreFiles || node.HasMoreSubfolders || node.IsIgnored {
		return true
	}
	for i := range node.SubFolders {
		if c.isTruncated(&node.SubFolders[i]) {
			return true
		}
	}
	return false
}

// containsString checks if a string is present in a slice of strings.
func containsString(slice []string, str string) bool {
	if slice == nil {
		return false
	}
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}