package services

import (
	"time"

	"go-ai-agent-v2/go-cli/pkg/types"
)

// SessionService handles the persistence of chat sessions.
type SessionService struct {
	store SessionStore
}

// NewSessionService creates a new SessionService.
func NewSessionService(store SessionStore) (*SessionService, error) {
	return &SessionService{
		store: store,
	}, nil
}

// SaveHistory saves the chat history for a given session ID.
func (s *SessionService) SaveHistory(sessionID string, history []*types.Content) error {
	return s.store.Save(sessionID, history)
}

// LoadHistory loads the chat history for a given session ID.
func (s *SessionService) LoadHistory(sessionID string) ([]*types.Content, error) {
	return s.store.Load(sessionID)
}

// ListSessions returns a sorted list of all saved session IDs, newest first.
func (s *SessionService) ListSessions() ([]string, error) {
	return s.store.List()
}

// DeleteSession deletes the session file for a given session ID.
func (s *SessionService) DeleteSession(sessionID string) error {
	return s.store.Delete(sessionID)
}

// GenerateSessionID creates a new session ID based on the current timestamp.
func (s *SessionService) GenerateSessionID() string {
	return time.Now().Format("20060102-150405")
}
