package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/config"
	"go-ai-agent-v2/go-cli/pkg/types"
)

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
func GetDirectoryContextString(config *config.Config) (string, error) {
	workspaceContext := config.GetWorkspaceContext()
	workspaceDirectories := workspaceContext.GetDirectories()

	var folderStructures []string
	for _, dir := range workspaceDirectories {
		structure, err := _getFolderStructure(dir, &types.FolderStructureOptions{}, config.GetFileService())
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

	return fmt.Sprintf("%s\nHere is the folder structure of the current working directories:\n\n%s", workingDirPreamble, folderStructure), nil
}

// _getFolderStructure generates a string representation of a directory's structure.
func _getFolderStructure(directory string, options *types.FolderStructureOptions, fileService config.FileService) (string, error) {
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


	structureRoot, err := readFullStructure(resolvedPath, &mergedOptions, fileService)
	if err != nil {
		return "", err
	}
	if structureRoot == nil {
		return fmt.Sprintf("Error: Could not read directory \"%s\". Check path and permissions.", resolvedPath), nil
	}

	structureLines := []string{}
	formatStructure(structureRoot, "", true, true, &structureLines)

	summary := fmt.Sprintf("Showing up to %d items (files + folders).", *mergedOptions.MaxItems)

	if isTruncated(structureRoot) {
		summary += fmt.Sprintf(" Folders or files indicated with %s contain more items not shown, were ignored, or the display limit (%d items) was reached.", TRUNCATION_INDICATOR, *mergedOptions.MaxItems)
	}

	return fmt.Sprintf("%s\n\n%s%c\n%s", summary, resolvedPath, filepath.Separator, strings.Join(structureLines, "\n")), nil
}

// intPtr returns a pointer to an int.
func intPtr(i int) *int {
	return &i
}

// readFullStructure reads the directory structure using BFS, respecting maxItems.
func readFullStructure(rootPath string, options *types.FolderStructureOptions, fileService config.FileService) (*types.FullFolderInfo, error) {
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
				if currentPath == rootPath && os.IsNotExist(err) {
					return nil, nil // Root directory itself not found
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
				

				// TODO: Implement shouldIgnoreFile using fileService
				// For now, a dummy check
				isIgnored := false
				if fileService != nil {
					// Not implemented yet
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
				// TODO: Implement shouldIgnoreFile using fileService
				if fileService != nil {
					// Not implemented yet
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
func formatStructure(
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
		formatStructure(&node.SubFolders[i], indentForChildren, isLastSubfolderAmongSiblings, false, builder)
	}
	if node.HasMoreSubfolders {
		*builder = append(*builder, fmt.Sprintf("%s└───%s", indentForChildren, TRUNCATION_INDICATOR))
	}
}

// isTruncated checks if any part of the folder structure was truncated or ignored.
func isTruncated(node *types.FullFolderInfo) bool {
	if node.HasMoreFiles || node.HasMoreSubfolders || node.IsIgnored {
		return true
	}
	for i := range node.SubFolders {
		if isTruncated(&node.SubFolders[i]) {
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