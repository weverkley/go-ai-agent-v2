package services

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileSystemService provides functionality to interact with the file system.
type FileSystemService struct{}

// NewFileSystemService creates a new instance of FileSystemService.
func NewFileSystemService() *FileSystemService {
	return &FileSystemService{}
}

// ListDirectory lists the contents of a directory.
func (s *FileSystemService) ListDirectory(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names, nil
}

// PathExists checks if a file or directory exists at the given path.
func (s *FileSystemService) PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check path existence for %s: %w", path, err)
}

// IsDirectory checks if the given path is a directory.
func (s *FileSystemService) IsDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Path does not exist, so it's not a directory
		}
		return false, fmt.Errorf("failed to get file info for %s: %w", path, err)
	}
	return info.IsDir(), nil
}

// ReadFile reads the content of a file at the given path.
func (s *FileSystemService) ReadFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return string(content), nil
}

// WriteFile writes the given content to a file at the given path.
func (s *FileSystemService) WriteFile(filePath string, content string) error {
	err := os.WriteFile(filePath, []byte(content), 0644) // 0644 is standard file permissions
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	return nil
}

// JoinPaths joins any number of path elements into a single path, adding a separating slash if necessary.
func (s *FileSystemService) JoinPaths(elements ...string) string {
	return filepath.Join(elements...)
}

