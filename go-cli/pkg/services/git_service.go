package services

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GitService provides functionality to interact with Git repositories.
type GitService struct {
	// No longer needs shellService
}

// NewGitService creates a new instance of GitService.
func NewGitService() *GitService {
	return &GitService{}
}

// GetCurrentBranch returns the name of the current Git branch using go-git.
func (s *GitService) GetCurrentBranch(dir string) (string, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return "", fmt.Errorf("failed to open git repository at %s: %w", dir, err)
	}

	headRef, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get head reference: %w", err)
	}

	if headRef.Name().IsBranch() {
		return strings.TrimPrefix(headRef.Name().String(), "refs/heads/"), nil
	} else if headRef.Name().IsTag() {
		return strings.TrimPrefix(headRef.Name().String(), "refs/tags/"), nil
	} else if headRef.Type() == plumbing.HashReference {
		// Detached HEAD, return the short hash
		return headRef.Hash().String()[:7], nil
	}

	return "", fmt.Errorf("could not determine current branch/tag/commit")
}
