package services

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gobwas/glob"
)

// FileSystemService interface defines the methods for interacting with the file system.
type FileSystemService interface {
	ListDirectory(dirPath string, ignorePatterns []string, respectGitIgnore, respectGeminiIgnore bool) ([]string, error)
	PathExists(path string) (bool, error)
	IsDirectory(path string) (bool, error)
	ReadFile(filePath string) (string, error)
	WriteFile(filePath string, content string) error
	CreateDirectory(path string) error
	CopyDirectory(src string, dst string) error
	JoinPaths(elements ...string) string
}

// fileSystemService implements the FileSystemService interface.
type fileSystemService struct{}

// NewFileSystemService creates a new instance of FileSystemService.
func NewFileSystemService() FileSystemService {
	return &fileSystemService{}
}

// getIgnorePatterns reads .gitignore and .geminiignore files and returns a list of glob patterns.
func (s *fileSystemService) getIgnorePatterns(searchDir string, respectGitIgnore, respectGeminiIgnore bool) ([]glob.Glob, error) {
	var ignorePatterns []glob.Glob

	// Read .gitignore
	if respectGitIgnore {
		gitIgnorePath := filepath.Join(searchDir, ".gitignore")
		if _, err := os.Stat(gitIgnorePath); err == nil {
			patterns, err := s.readIgnoreFile(gitIgnorePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read .gitignore: %w", err)
			}
			ignorePatterns = append(ignorePatterns, patterns...)
		}
	}

	// Read .geminiignore
	if respectGeminiIgnore {
		geminiIgnorePath := filepath.Join(searchDir, ".geminiignore")
		if _, err := os.Stat(geminiIgnorePath); err == nil {
			patterns, err := s.readIgnoreFile(geminiIgnorePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read .geminiignore: %w", err)
			}
			ignorePatterns = append(ignorePatterns, patterns...)
		}
	}

	return ignorePatterns, nil
}

// readIgnoreFile reads an ignore file and compiles its patterns.
func (s *fileSystemService) readIgnoreFile(filePath string) ([]glob.Glob, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []glob.Glob
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}
		g, err := glob.Compile(line)
		if err != nil {
			return nil, fmt.Errorf("failed to compile ignore pattern %s from %s: %w", line, filePath, err)
		}
		patterns = append(patterns, g)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return patterns, nil
}

// ListDirectory lists the contents of a directory.
func (s *fileSystemService) ListDirectory(dirPath string, ignorePatterns []string, respectGitIgnore, respectGeminiIgnore bool) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	var names []string
	var compiledIgnoreGlobs []glob.Glob

	// Compile ignore patterns from argument
	for _, p := range ignorePatterns {
		g, err := glob.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("failed to compile ignore pattern %s: %w", p, err)
		}
		compiledIgnoreGlobs = append(compiledIgnoreGlobs, g)
	}

	// Get ignore patterns from files
	fileIgnoreGlobs, err := s.getIgnorePatterns(dirPath, respectGitIgnore, respectGeminiIgnore)
	if err != nil {
		return nil, fmt.Errorf("failed to get ignore patterns from files: %w", err)
	}
	compiledIgnoreGlobs = append(compiledIgnoreGlobs, fileIgnoreGlobs...)

	for _, entry := range entries {
		// Check against ignore patterns
		shouldIgnore := false
		for _, ignoreG := range compiledIgnoreGlobs {
			if ignoreG.Match(entry.Name()) {
				shouldIgnore = true
				break
			}
		}
		if shouldIgnore {
			continue
		}
		names = append(names, entry.Name())
	}

	// Sort entries (directories first, then alphabetically)
	sort.Slice(names, func(i, j int) bool {
		pathI := filepath.Join(dirPath, names[i])
		pathJ := filepath.Join(dirPath, names[j])

		infoI, errI := os.Stat(pathI)
		infoJ, errJ := os.Stat(pathJ)

		isDirI := errI == nil && infoI.IsDir()
		isDirJ := errJ == nil && infoJ.IsDir()

		if isDirI && !isDirJ {
			return true
		}
		if !isDirI && isDirJ {
			return false
		}
		return names[i] < names[j]
	})

	return names, nil
}

// PathExists checks if a file or directory exists at the given path.
func (s *fileSystemService) PathExists(path string) (bool, error) {
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
func (s *fileSystemService) IsDirectory(path string) (bool, error) {
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
func (s *fileSystemService) ReadFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return string(content), nil
}

// WriteFile writes the given content to a file at the given path.
func (s *fileSystemService) WriteFile(filePath string, content string) error {
	err := os.WriteFile(filePath, []byte(content), 0644) // 0644 is standard file permissions
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	return nil
}

// JoinPaths joins any number of path elements into a single path, adding a separating slash if necessary.
func (s *fileSystemService) JoinPaths(elements ...string) string {
	return filepath.Join(elements...)
}

// CreateDirectory creates a directory, ensuring it doesn't already exist.
func (s *fileSystemService) CreateDirectory(path string) error {
	exists, err := s.PathExists(path)
	if err != nil {
		return fmt.Errorf("failed to check path existence for %s: %w", path, err)
	}
	if exists {
		return fmt.Errorf("path already exists: %s", path)
	}

	err = os.MkdirAll(path, 0755) // 0755 is standard directory permissions
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// CopyDirectory recursively copies a directory from src to dst.
func (s *fileSystemService) CopyDirectory(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	dirents, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, dirent := range dirents {
		srcPath := filepath.Join(src, dirent.Name())
		dstPath := filepath.Join(dst, dirent.Name())

		if dirent.IsDir() {
			err = s.CopyDirectory(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = s.copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile copies a file from src to dst.
func (s *fileSystemService) copyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}