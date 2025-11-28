package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/types"
)

const (
	fileSuffix = "_session.json"
)

// FileSessionStore is a file-based implementation of the SessionStore interface.
type FileSessionStore struct {
	sessionsPath string
}

// NewFileSessionStore creates a new FileSessionStore.
func NewFileSessionStore(sessionsPath string) (*FileSessionStore, error) {
	if err := os.MkdirAll(sessionsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
	}
	return &FileSessionStore{
		sessionsPath: sessionsPath,
	}, nil
}

// getSessionFilePath returns the full path for a given session ID.
func (s *FileSessionStore) getSessionFilePath(sessionID string) string {
	return filepath.Join(s.sessionsPath, fmt.Sprintf("%s%s", sessionID, fileSuffix))
}

// Save saves the chat history for a given session ID to a JSON file.
func (s *FileSessionStore) Save(sessionID string, history []*types.Content) error {
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

// Load loads the chat history for a given session ID from a JSON file.
func (s *FileSessionStore) Load(sessionID string) ([]*types.Content, error) {
	filePath := s.getSessionFilePath(sessionID)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []*types.Content{}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	if len(data) == 0 {
		return []*types.Content{}, nil
	}

	var history []*types.Content
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return history, nil
}

// List returns a sorted list of all saved session IDs, newest first.
func (s *FileSessionStore) List() ([]string, error) {
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

	sort.SliceStable(sessions, func(i, j int) bool {
		infoI, errI := os.Stat(s.getSessionFilePath(sessions[i]))
		infoJ, errJ := os.Stat(s.getSessionFilePath(sessions[j]))
		if errI != nil || errJ != nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	return sessions, nil
}

// Delete deletes the session file for a given session ID.
func (s *FileSessionStore) Delete(sessionID string) error {
	filePath := s.getSessionFilePath(sessionID)
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete session file: %w", err)
	}
	return nil
}
