package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go-ai-agent-v2/go-cli/pkg/core" // Import core package
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/google/generative-ai-go/genai"
)

const (
	checkpointFilePrefix = "checkpoint-"
	checkpointFileSuffix = ".json"
)

// ChatService provides methods for managing chat checkpoints.
type ChatService struct {
	config   types.GoaiagentDirProvider
	executor core.Executor // Add executor field
}

// NewChatService creates a new ChatService instance.
func NewChatService(cfg types.GoaiagentDirProvider, executor core.Executor) *ChatService {
	return &ChatService{config: cfg, executor: executor}
}

// ChatDetail represents details of a saved chat checkpoint.
type ChatDetail struct {
	Name  string    `json:"name"`
	Mtime time.Time `json:"mtime"`
}

// SerializablePart represents a serializable part of a genai.Content.
type SerializablePart struct {
	Text           string                 `json:"text,omitempty"`
	FunctionCall   *genai.FunctionCall    `json:"functionCall,omitempty"`
	FunctionResponse *genai.FunctionResponse `json:"functionResponse,omitempty"`
	// Add other part types as needed
}

// SerializableContent represents a serializable genai.Content.
type SerializableContent struct {
	Parts []SerializablePart `json:"parts,omitempty"`
	Role  string             `json:"role,omitempty"`
}

// getProjectTempDir returns the project's temporary directory.
func (cs *ChatService) getProjectTempDir() (string, error) {
	geminiDir := cs.config.GetGoaiagentDir()
	return filepath.Join(geminiDir, "checkpoints"), nil
}

// GetSavedChatTags retrieves details of all saved chat checkpoints.
func (cs *ChatService) GetSavedChatTags(mtSortDesc bool) ([]ChatDetail, error) {
	geminiDir, err := cs.getProjectTempDir()
	if err != nil {
		return nil, fmt.Errorf("could not access temporary directory for chat checkpoints: %w", err)
	}

	// Ensure the directory exists
	if _, err := os.Stat(geminiDir); os.IsNotExist(err) {
		return []ChatDetail{}, nil // No directory, no checkpoints
	}

	files, err := os.ReadDir(geminiDir)
	if err != nil {
		return nil, fmt.Errorf("could not read chat checkpoint directory '%s': %w", geminiDir, err)
	}

	chatDetails := []ChatDetail{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		if !strings.HasPrefix(fileName, checkpointFilePrefix) || !strings.HasSuffix(fileName, checkpointFileSuffix) {
			continue
		}

		filePath := filepath.Join(geminiDir, fileName)
		info, err := file.Info()
		if err != nil {
			// Log error but continue with other files
			fmt.Printf("Warning: failed to get file info for %s: %v\n", filePath, err)
			continue
		}

		tagName := fileName[len(checkpointFilePrefix) : len(fileName)-len(checkpointFileSuffix)]
		chatDetails = append(chatDetails, ChatDetail{
			Name:  tagName,
			Mtime: info.ModTime(),
		})
	}

	sort.Slice(chatDetails, func(i, j int) bool {
		if mtSortDesc {
			return chatDetails[i].Mtime.After(chatDetails[j].Mtime)
		}
		return chatDetails[i].Mtime.Before(chatDetails[j].Mtime)
	})

	return chatDetails, nil
}

// SaveCheckpoint saves the given history as a checkpoint with the specified tag.
func (cs *ChatService) SaveCheckpoint(history []*genai.Content, tag string) error {
	geminiDir, err := cs.getProjectTempDir()
	if err != nil {
		return fmt.Errorf("could not access temporary directory for chat checkpoints: %w", err)
	}

	// Ensure the directory exists
	if err := os.MkdirAll(geminiDir, 0755); err != nil {
		return fmt.Errorf("could not create checkpoint directory '%s': %w", geminiDir, err)
	}

	filePath := filepath.Join(geminiDir, fmt.Sprintf("%s%s%s", checkpointFilePrefix, tag, checkpointFileSuffix))

	// Convert []*genai.Content to []*SerializableContent
	serializableHistory := make([]*SerializableContent, len(history))
	for i, content := range history {
		serializableParts := make([]SerializablePart, len(content.Parts))
		for j, part := range content.Parts {
			if text, ok := part.(genai.Text); ok {
				serializableParts[j].Text = string(text)
			} else if fc, ok := part.(*genai.FunctionCall); ok {
				serializableParts[j].FunctionCall = fc
			} else if fr, ok := part.(*genai.FunctionResponse); ok {
				serializableParts[j].FunctionResponse = fr
			}
		}
		serializableHistory[i] = &SerializableContent{
			Parts: serializableParts,
			Role:  content.Role,
		}
	}

	data, err := json.MarshalIndent(serializableHistory, "", "  ")
	if err != nil {
		return fmt.Errorf("could not prepare chat history for saving: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("could not write checkpoint file '%s': %w", filePath, err)
	}

	return nil
}

// LoadCheckpoint loads a checkpoint with the specified tag.
func (cs *ChatService) LoadCheckpoint(tag string) ([]*genai.Content, error) {
	geminiDir, err := cs.getProjectTempDir()
	if err != nil {
		return nil, fmt.Errorf("could not access temporary directory for chat checkpoints: %w", err)
	}

	filePath := filepath.Join(geminiDir, fmt.Sprintf("%s%s%s", checkpointFilePrefix, tag, checkpointFileSuffix))

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read checkpoint file '%s': %w", filePath, err)
	}

	var serializableHistory []*SerializableContent
	if err := json.Unmarshal(data, &serializableHistory); err != nil {
		return nil, fmt.Errorf("could not parse chat history from '%s': %w", filePath, err)
	}

	// Convert []*SerializableContent back to []*genai.Content
	history := make([]*genai.Content, len(serializableHistory))
	for i, sContent := range serializableHistory {
		genaiParts := make([]genai.Part, len(sContent.Parts))
		for j, sPart := range sContent.Parts {
			if sPart.Text != "" {
				genaiParts[j] = genai.Text(sPart.Text)
			} else if sPart.FunctionCall != nil {
				genaiParts[j] = sPart.FunctionCall
			} else if sPart.FunctionResponse != nil {
				genaiParts[j] = sPart.FunctionResponse
			}
		}
		history[i] = &genai.Content{
			Parts: genaiParts,
			Role:  sContent.Role,
		}
	}

	return history, nil
}

// CheckpointExists checks if a checkpoint with the given tag already exists.
func (cs *ChatService) CheckpointExists(tag string) (bool, error) {
	geminiDir, err := cs.getProjectTempDir()
	if err != nil {
		return false, fmt.Errorf("could not access temporary directory for chat checkpoints: %w", err)
	}

	filePath := filepath.Join(geminiDir, fmt.Sprintf("%s%s%s", checkpointFilePrefix, tag, checkpointFileSuffix))

	_, err = os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("could not determine if checkpoint '%s' exists: %w", tag, err)
}

// DeleteCheckpoint deletes a checkpoint with the specified tag.
func (cs *ChatService) DeleteCheckpoint(tag string) (bool, error) {
	geminiDir, err := cs.getProjectTempDir()
	if err != nil {
		return false, fmt.Errorf("could not access temporary directory for chat checkpoints: %w", err)
	}

	filePath := filepath.Join(geminiDir, fmt.Sprintf("%s%s%s", checkpointFilePrefix, tag, checkpointFileSuffix))

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false, nil // Checkpoint does not exist
	}

	if err := os.Remove(filePath); err != nil {
		return false, fmt.Errorf("could not delete checkpoint file '%s': %w", filePath, err)
	}

	return true, nil
}
