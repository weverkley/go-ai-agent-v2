package services

import "go-ai-agent-v2/go-cli/pkg/types"

// SessionStore is the interface for session persistence.
type SessionStore interface {
	Save(sessionID string, history []*types.Content) error
	Load(sessionID string) ([]*types.Content, error)
	List() ([]string, error)
	Delete(sessionID string) error
}
