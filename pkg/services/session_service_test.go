package services

import (
	"os"
	"testing"
	"time"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/stretchr/testify/assert"
)

// setupTestSessionService creates a temporary directory for session files.
func setupTestSessionService(t *testing.T) (*SessionService, func()) {
	sessionsPath, err := os.MkdirTemp("", "sessions_*")
	assert.NoError(t, err)

	store, err := NewFileSessionStore(sessionsPath)
	assert.NoError(t, err)

	ss, err := NewSessionService(store)
	assert.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(sessionsPath)
	}

	return ss, cleanup
}

func TestNewSessionService(t *testing.T) {
	_, cleanup := setupTestSessionService(t)
	defer cleanup()
	// The setup function already tests the creation of the service
}

func TestSaveAndLoadHistory(t *testing.T) {
	ss, cleanup := setupTestSessionService(t)
	defer cleanup()

	sessionID := "test_session_1"
	history := []*types.Content{
		{Role: "user", Parts: []types.Part{{Text: "Hello"}}},
		{Role: "model", Parts: []types.Part{{Text: "Hi there!"}}},
	}

	// 1. Save history
	err := ss.SaveHistory(sessionID, history)
	assert.NoError(t, err)

	// 2. Load history
	loadedHistory, err := ss.LoadHistory(sessionID)
	assert.NoError(t, err)
	assert.Equal(t, len(history), len(loadedHistory))
	assert.Equal(t, "Hello", loadedHistory[0].Parts[0].Text)
	assert.Equal(t, "Hi there!", loadedHistory[1].Parts[0].Text)
}

func TestLoadHistory_NonExistent(t *testing.T) {
	ss, cleanup := setupTestSessionService(t)
	defer cleanup()

	// Try to load a session that doesn't exist
	history, err := ss.LoadHistory("non_existent_session")
	assert.NoError(t, err)
	assert.Empty(t, history, "history should be empty for a non-existent session")
}

func TestListSessions(t *testing.T) {
	ss, cleanup := setupTestSessionService(t)
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

	sessions, err := ss.ListSessions()
	assert.NoError(t, err)

	assert.Equal(t, 3, len(sessions), "should only list session files")
	// Verify they are sorted by mod time, newest first
	assert.Equal(t, id3, sessions[0])
	assert.Equal(t, id2, sessions[1])
	assert.Equal(t, id1, sessions[2])
}

func TestDeleteSession(t *testing.T) {
	ss, cleanup := setupTestSessionService(t)
	defer cleanup()

	sessionID := "session_to_delete"
	err := ss.SaveHistory(sessionID, []*types.Content{})
	assert.NoError(t, err)

	// Delete it
	err = ss.DeleteSession(sessionID)
	assert.NoError(t, err)

	// Verify it's gone
	history, err := ss.LoadHistory(sessionID)
	assert.NoError(t, err)
	assert.Empty(t, history)

	// Deleting a non-existent session should not error
	err = ss.DeleteSession("non_existent_session_id")
	assert.NoError(t, err)
}

func TestGenerateSessionID(t *testing.T) {
	ss, cleanup := setupTestSessionService(t)
	defer cleanup()

	sessionID := ss.GenerateSessionID()
	// Example: 20240115-150405
	_, err := time.Parse("20060102-150405", sessionID)
	assert.NoError(t, err, "session ID should be in the expected time format")
}
