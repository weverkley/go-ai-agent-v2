package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-ai-agent-v2/go-cli/pkg/types"
	"github.com/stretchr/testify/assert"
)

// setupTestSessionService creates a temporary directory for session files.
func setupTestSessionService(t *testing.T) (*SessionService, string, func()) {
	// Create a temp directory to simulate the project root
	projectRoot, err := os.MkdirTemp("", "project_root_*")
	assert.NoError(t, err)

	// Create the .goaiagent directory within the temp project root
	goaiagentDir := filepath.Join(projectRoot, ".goaiagent")
	err = os.Mkdir(goaiagentDir, 0755)
	assert.NoError(t, err)

	ss, err := NewSessionService(goaiagentDir)
	assert.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(projectRoot) // Clean up the entire temp project root
	}

	return ss, goaiagentDir, cleanup
}

func TestNewSessionService(t *testing.T) {
	ss, goaiagentDir, cleanup := setupTestSessionService(t)
	defer cleanup()

	expectedPath := filepath.Join(goaiagentDir, "sessions")
	assert.Equal(t, expectedPath, ss.sessionsPath)
	_, err := os.Stat(expectedPath)
	assert.False(t, os.IsNotExist(err), "sessions directory should exist")
}

func TestSaveAndLoadHistory(t *testing.T) {
	ss, _, cleanup := setupTestSessionService(t)
	defer cleanup()

	sessionID := "test_session_1"
	history := []*types.Content{
		{Role: "user", Parts: []types.Part{{Text: "Hello"}}},
		{Role: "model", Parts: []types.Part{{Text: "Hi there!"}}},
	}

	// 1. Save history
	err := ss.SaveHistory(sessionID, history)
	assert.NoError(t, err)

	// 2. Verify file exists
	filePath := ss.getSessionFilePath(sessionID)
	_, err = os.Stat(filePath)
	assert.False(t, os.IsNotExist(err), "session file should have been created")

	// 3. Load history
	loadedHistory, err := ss.LoadHistory(sessionID)
	assert.NoError(t, err)
	assert.Equal(t, len(history), len(loadedHistory))
	assert.Equal(t, "Hello", loadedHistory[0].Parts[0].Text)
	assert.Equal(t, "Hi there!", loadedHistory[1].Parts[0].Text)
}

func TestLoadHistory_NonExistent(t *testing.T) {
	ss, _, cleanup := setupTestSessionService(t)
	defer cleanup()

	// Try to load a session that doesn't exist
	history, err := ss.LoadHistory("non_existent_session")
	assert.NoError(t, err)
	assert.Empty(t, history, "history should be empty for a non-existent session")
}

func TestListSessions(t *testing.T) {
	ss, _, cleanup := setupTestSessionService(t)
	defer cleanup()

	// Create some session files with different modification times
	id1 := "session_aaa"
	id2 := "session_bbb"
	id3 := "session_ccc"

	err := ss.SaveHistory(id1, []*types.Content{})
	assert.NoError(t, err)
	time.Sleep(10 * time.Millisecond) // Ensure different mod times

	err = ss.SaveHistory(id2, []*types.Content{})
	assert.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	err = ss.SaveHistory(id3, []*types.Content{})
	assert.NoError(t, err)

	// Create a non-session file that should be ignored
	nonSessionFile := filepath.Join(ss.sessionsPath, "ignore_this.txt")
	os.WriteFile(nonSessionFile, []byte("ignore"), 0644)

	sessions, err := ss.ListSessions()
	assert.NoError(t, err)

	assert.Equal(t, 3, len(sessions), "should only list session files")
	// Verify they are sorted by mod time, newest first
	assert.Equal(t, id3, sessions[0])
	assert.Equal(t, id2, sessions[1])
	assert.Equal(t, id1, sessions[2])
}

func TestDeleteSession(t *testing.T) {
	ss, _, cleanup := setupTestSessionService(t)
	defer cleanup()

	sessionID := "session_to_delete"
	err := ss.SaveHistory(sessionID, []*types.Content{})
	assert.NoError(t, err)

	// Verify it exists
	filePath := ss.getSessionFilePath(sessionID)
	_, err = os.Stat(filePath)
	assert.False(t, os.IsNotExist(err))

	// Delete it
	err = ss.DeleteSession(sessionID)
	assert.NoError(t, err)

	// Verify it's gone
	_, err = os.Stat(filePath)
	assert.True(t, os.IsNotExist(err))

	// Deleting a non-existent session should not error
	err = ss.DeleteSession("non_existent_session_id")
	assert.NoError(t, err)
}

func TestGenerateSessionID(t *testing.T) {
	ss, _, cleanup := setupTestSessionService(t)
	defer cleanup()

	sessionID := ss.GenerateSessionID()
	// Example: 20240115-150405
	_, err := time.Parse("20060102-150405", sessionID)
	assert.NoError(t, err, "session ID should be in the expected time format")
}

