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

// GetRemoteURL returns the URL of the "origin" remote for the given Git repository.
func (s *GitService) GetRemoteURL(dir string) (string, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return "", fmt.Errorf("failed to open git repository at %s: %w", dir, err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return "", fmt.Errorf("failed to get remote 'origin': %w", err)
	}

	if len(remote.Config().URLs) == 0 {
		return "", fmt.Errorf("no URLs found for remote 'origin'")
	}

	return remote.Config().URLs[0], nil
}

// CheckoutBranch checks out the specified branch in the given repository.
func (s *GitService) CheckoutBranch(dir string, branchName string) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("failed to open git repository at %s: %w", dir, err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree for %s: %w", dir, err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
	}
	return nil
}

// Pull pulls the latest changes from the remote for the current branch.
func (s *GitService) Pull(dir string) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("failed to open git repository at %s: %w", dir, err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree for %s: %w", dir, err)
	}

	err = w.Pull(&git.PullOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull latest changes: %w", err)
	}
	return nil
}

// DeleteBranch deletes the specified branch locally.
func (s *GitService) DeleteBranch(dir string, branchName string) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("failed to open git repository at %s: %w", dir, err)
	}

	err = repo.Storer.RemoveReference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)))
	if err != nil {
		return fmt.Errorf("failed to delete branch %s: %w", branchName, err)
	}
	return nil
}
