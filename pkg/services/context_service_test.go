package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextService_NewContextService(t *testing.T) {
	tempDir := t.TempDir()
	cs := NewContextService(tempDir)
	assert.NotNil(t, cs)
	assert.Equal(t, tempDir, cs.baseDir)
}

func TestContextService_Load_GlobalFile(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	// Create a global GOAIAGENT.md file
	globalDir := filepath.Join(homeDir, ".goaiagent")
	os.MkdirAll(globalDir, 0755)
	globalFile := filepath.Join(globalDir, "GOAIAGENT.md")
	os.WriteFile(globalFile, []byte("Global context"), 0644)

	cs := NewContextService(t.TempDir()) // Base dir doesn't matter for this test
	err := cs.Load()
	assert.NoError(t, err)
	assert.Equal(t, "Global context", cs.GetContext())
}

func TestContextService_Load_AncestorFiles(t *testing.T) {
	projectRoot := t.TempDir()
	os.Mkdir(filepath.Join(projectRoot, ".git"), 0755)
	os.WriteFile(filepath.Join(projectRoot, "GOAIAGENT.md"), []byte("Root context"), 0644)

	subDir := filepath.Join(projectRoot, "sub")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "GOAIAGENT.md"), []byte("Sub context"), 0644)

	cs := NewContextService(subDir)
	err := cs.Load()
	assert.NoError(t, err)

	// Note: The order is root then sub, because we reverse the ancestor files.
	expected := "Root context\nSub context"
	assert.Equal(t, expected, cs.GetContext())
}

func TestContextService_Load_WithImports(t *testing.T) {
	projectRoot := t.TempDir()
	os.WriteFile(filepath.Join(projectRoot, "imported.md"), []byte("Imported content"), 0644)
	os.WriteFile(filepath.Join(projectRoot, "GOAIAGENT.md"), []byte("Main context\n@imported.md"), 0644)

	cs := NewContextService(projectRoot)
	err := cs.Load()
	assert.NoError(t, err)

	expected := "Main context\nImported content"
	assert.Equal(t, expected, cs.GetContext())
}

func TestContextService_AddToGlobalMemory(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	cs := NewContextService(t.TempDir())
	err := cs.AddToGlobalMemory("New memory")
	assert.NoError(t, err)

	globalFile := filepath.Join(homeDir, ".goaiagent", "GOAIAGENT.md")
	content, err := os.ReadFile(globalFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "New memory")

	// Check that the context is reloaded
	assert.Contains(t, cs.GetContext(), "New memory")
}

func TestContextService_ShowMemory(t *testing.T) {
	cs := NewContextService(t.TempDir())
	cs.context = "Test context"
	assert.Equal(t, "Test context", cs.ShowMemory())
}

func TestContextService_ListMemoryFiles(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	globalDir := filepath.Join(homeDir, ".goaiagent")
	os.MkdirAll(globalDir, 0755)
	globalFile := filepath.Join(globalDir, "GOAIAGENT.md")
	os.WriteFile(globalFile, []byte("Global context"), 0644)

	projectRoot := t.TempDir()
	os.Mkdir(filepath.Join(projectRoot, ".git"), 0755)
	projectFile := filepath.Join(projectRoot, "GOAIAGENT.md")
	os.WriteFile(projectFile, []byte("Project context"), 0644)

	cs := NewContextService(projectRoot)
	err := cs.Load()
	assert.NoError(t, err)

	files := cs.ListMemoryFiles()
	assert.Len(t, files, 2)

	// Normalize paths for comparison
	expectedFiles := []string{
		filepath.Clean(globalFile),
		filepath.Clean(projectFile),
	}
	actualFiles := []string{}
	for _, f := range files {
		actualFiles = append(actualFiles, filepath.Clean(f))
	}
	assert.ElementsMatch(t, expectedFiles, actualFiles)
}

func TestContextService_Load_Empty(t *testing.T) {
	cs := NewContextService(t.TempDir())
	err := cs.Load()
	assert.NoError(t, err)
	assert.Equal(t, "", cs.GetContext())
}

func TestContextService_Load_WithSubdirectories(t *testing.T) {
	// This test is to show that sub-directory scanning is not yet implemented.
	// When it is, this test should be updated.
	projectRoot := t.TempDir()
	os.WriteFile(filepath.Join(projectRoot, "GOAIAGENT.md"), []byte("Root context"), 0644)

	subDir := filepath.Join(projectRoot, "sub")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "GOAIAGENT.md"), []byte("Sub context"), 0644)

	cs := NewContextService(projectRoot) // Searching from root
	err := cs.Load()
	assert.NoError(t, err)

	// Currently, only the root file is loaded.
	// When sub-directory scanning is added, this should also contain "Sub context".
	assert.Equal(t, "Root context", cs.GetContext())
}
