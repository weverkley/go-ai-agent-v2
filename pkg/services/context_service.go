package services

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const (
	defaultContextFileName = "GOAIAGENT.md"
)

// ContextService manages the hierarchical context from GOAIAGENT.md files.
type ContextService struct {
	mu          sync.RWMutex
	context     string
	baseDir     string
	fileNames   []string
	fileContent map[string]string // Cache for file content
}

// NewContextService creates a new ContextService instance.
func NewContextService(baseDir string) *ContextService {
	cs := &ContextService{
		baseDir:     baseDir,
		fileNames:   []string{defaultContextFileName},
		fileContent: make(map[string]string),
	}
	// Initial load, ignore error as files might not exist.
	_ = cs.Load()
	return cs
}

// GetContext returns the concatenated context from all discovered files.
func (cs *ContextService) GetContext() string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.context
}

// Load discovers, reads, and processes all context files.
func (cs *ContextService) Load() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Reset content
	cs.context = ""
	cs.fileContent = make(map[string]string)

	// 1. Global context file
	homeDir, err := os.UserHomeDir()
	if err == nil {
		for _, fileName := range cs.fileNames {
			globalFile := filepath.Join(homeDir, ".goaiagent", fileName)
			if content, err := cs.readFileAndProcessImports(globalFile); err == nil {
				cs.context += content + "\n"
			}
		}
	}

	// 2. Project root and ancestor context files
	ancestorFiles, err := cs.findAncestorFiles(cs.baseDir)
	if err == nil {
		// Reverse to process from root down to current
		for i := len(ancestorFiles) - 1; i >= 0; i-- {
			if content, err := cs.readFileAndProcessImports(ancestorFiles[i]); err == nil {
				cs.context += content + "\n"
			}
		}
	}

	// 3. Sub-directory context files (not implemented for simplicity, but can be added)

	cs.context = strings.TrimSpace(cs.context)
	return nil
}

// readFileAndProcessImports reads a file and processes any @file.md imports.
func (cs *ContextService) readFileAndProcessImports(filePath string) (string, error) {
	if content, found := cs.fileContent[filePath]; found {
		return content, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	content := string(data)
	// Simple import processing
	re := regexp.MustCompile(`@(\S+\.md)`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		importPath := filepath.Join(filepath.Dir(filePath), match[1])
		importedContent, err := cs.readFileAndProcessImports(importPath)
		if err == nil {
			content = strings.Replace(content, match[0], importedContent, 1)
		}
	}

	cs.fileContent[filePath] = content
	return content, nil
}

// findAncestorFiles finds all context files from the given directory up to the project root (.git).
func (cs *ContextService) findAncestorFiles(startDir string) ([]string, error) {
	var files []string
	currentDir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, err
	}

	for {
		for _, fileName := range cs.fileNames {
			filePath := filepath.Join(currentDir, fileName)
			if _, err := os.Stat(filePath); err == nil {
				files = append(files, filePath)
			}
		}

		// Check for .git to stop at project root
		if _, err := os.Stat(filepath.Join(currentDir, ".git")); err == nil {
			break
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir { // Reached the root of the filesystem
			break
		}
		currentDir = parent
	}

	return files, nil
}

// AddToGlobalMemory appends text to the global GOAIAGENT.md file.
func (cs *ContextService) AddToGlobalMemory(text string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	globalDir := filepath.Join(homeDir, ".goaiagent")
	if err := os.MkdirAll(globalDir, 0755); err != nil {
		return err
	}
	globalFile := filepath.Join(globalDir, defaultContextFileName)

	f, err := os.OpenFile(globalFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString("\n" + text); err != nil {
		return err
	}
	// Reload context after adding
	return cs.Load()
}

// ShowMemory returns the full concatenated context.
func (cs *ContextService) ShowMemory() string {
	return cs.GetContext()
}

// ListMemoryFiles returns the paths of the loaded context files.
func (cs *ContextService) ListMemoryFiles() []string {
	// This is a simplified version. A more robust implementation would
	// track the files found during the Load() process.
	files, _ := cs.findAncestorFiles(cs.baseDir)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		for _, fileName := range cs.fileNames {
			globalFile := filepath.Join(homeDir, ".goaiagent", fileName)
			if _, err := os.Stat(globalFile); err == nil {
				files = append([]string{globalFile}, files...)
			}
		}
	}
	return files
}
