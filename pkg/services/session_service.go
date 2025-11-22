package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go-ai-agent-v2/go-cli/pkg/types"
)

const (
	sessionsDir = ".go-ai-agent/sessions"
	fileSuffix  = "_session.json"
)

// SessionService handles the persistence of chat sessions.
type SessionService struct {
	sessionsPath string
}

// NewSessionService creates a new SessionService.
// It ensures the session storage directory exists within the provided .goaiagent directory.
func NewSessionService(goaiagentDir string) (*SessionService, error) {
	if goaiagentDir == "" {
		return nil, fmt.Errorf("goaiagent directory path cannot be empty")
	}

	sessionsPath := filepath.Join(goaiagentDir, "sessions")

	if err := os.MkdirAll(sessionsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
	}

	return &SessionService{
		sessionsPath: sessionsPath,
	}, nil
}

// getSessionFilePath returns the full path for a given session ID.
func (s *SessionService) getSessionFilePath(sessionID string) string {
	return filepath.Join(s.sessionsPath, fmt.Sprintf("%s%s", sessionID, fileSuffix))
}

// SaveHistory saves the chat history for a given session ID to a JSON file.
func (s *SessionService) SaveHistory(sessionID string, history []*types.Content) error {
	filePath := s.getSessionFilePath(sessionID)
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}
	return nil
}

// LoadHistory loads the chat history for a given session ID from a JSON file.
// If the file does not exist, it returns an empty history and no error.
func (s *SessionService) LoadHistory(sessionID string) ([]*types.Content, error) {
	filePath := s.getSessionFilePath(sessionID)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Return empty history if file doesn't exist, which is normal for a new session.
		return []*types.Content{}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	if len(data) == 0 {
		// Return empty history if file is empty.
		return []*types.Content{}, nil
	}

	var history []*types.Content
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return history, nil
}

// ListSessions returns a sorted list of all saved session IDs, newest first.
func (s *SessionService) ListSessions() ([]string, error) {
	files, err := os.ReadDir(s.sessionsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var sessions []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), fileSuffix) {
			sessionID := strings.TrimSuffix(file.Name(), fileSuffix)
			sessions = append(sessions, sessionID)
		}
	}

	// Sort sessions by modification time, newest first.
	sort.SliceStable(sessions, func(i, j int) bool {
		infoI, errI := os.Stat(s.getSessionFilePath(sessions[i]))
		infoJ, errJ := os.Stat(s.getSessionFilePath(sessions[j]))
		if errI != nil || errJ != nil {
			return false // Cannot compare, keep original order
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	return sessions, nil
}

// DeleteSession deletes the session file for a given session ID.
func (s *SessionService) DeleteSession(sessionID string) error {
	filePath := s.getSessionFilePath(sessionID)
	if err := os.Remove(filePath); err != nil {
		// If the file doesn't exist, it's not an error in this context.
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete session file: %w", err)
	}
	return nil
}

// GenerateSessionID creates a new session ID based on the current timestamp.
func (s *SessionService) GenerateSessionID() string {
	return time.Now().Format("20060102-150405")
}
