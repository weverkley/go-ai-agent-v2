package services

import (
	"fmt"
	"strings"
	"time" // New import

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object" // New import
)

// GitService interface defines the methods for interacting with Git repositories.
type GitService interface {
	GetCurrentBranch(dir string) (string, error)
	GetRemoteURL(dir string) (string, error)
	CheckoutBranch(dir string, branchName string) error
	Pull(dir string, ref string) error
	Clone(url string, directory string, ref string) error
	DeleteBranch(dir string, branchName string) error
	StageFiles(dir string, files []string) error
	Commit(dir, message string) error
}

// gitService implements the GitService interface.
type gitService struct {
	// No longer needs shellService
}

// NewGitService creates a new instance of GitService.
func NewGitService() GitService {
	return &gitService{}
}

// GetCurrentBranch returns the name of the current Git branch using go-git.
func (s *gitService) GetCurrentBranch(dir string) (string, error) {
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
func (s *gitService) GetRemoteURL(dir string) (string, error) {
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
func (s *gitService) CheckoutBranch(dir string, branchName string) error {
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
func (s *gitService) Pull(dir string, ref string) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("failed to open git repository at %s: %w", dir, err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree for %s: %w", dir, err)
	}

	pullOptions := &git.PullOptions{}
	if ref != "" {
		pullOptions.ReferenceName = plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", ref))
	}

	err = w.Pull(pullOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull latest changes: %w", err)
	}
	return nil
}

// Clone clones a git repository from a URL into a specified directory and checks out a given reference.
func (s *gitService) Clone(url string, directory string, ref string) error {
	cloneOptions := &git.CloneOptions{
		URL: url,
	}

	if ref != "" {
		cloneOptions.ReferenceName = plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", ref))
		cloneOptions.SingleBranch = true
	}

	_, err := git.PlainClone(directory, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to clone repository %s: %w", url, err)
	}
	return nil
}

// DeleteBranch deletes the specified branch locally.
func (s *gitService) DeleteBranch(dir string, branchName string) error {
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

// StageFiles stages the specified files. If files is empty, it stages all changes.
func (s *gitService) StageFiles(dir string, files []string) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("failed to open git repository at %s: %w", dir, err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree for %s: %w", dir, err)
	}

	if len(files) == 0 {
		// Stage all changes
		_, err = w.Add(".")
		if err != nil {
			return fmt.Errorf("failed to add all files: %w", err)
		}
	} else {
		// Stage specific files
		for _, file := range files {
			_, err = w.Add(file)
			if err != nil {
				return fmt.Errorf("failed to add file %s: %w", file, err)
			}
		}
	}
	return nil
}

// Commit creates a new commit with the given message.
func (s *gitService) Commit(dir, message string) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return fmt.Errorf("failed to open git repository at %s: %w", dir, err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree for %s: %w", dir, err)
	}

	// This is a simplified approach, typically you'd configure the author/committer
	// from git config or environment variables. For now, a placeholder.
	commit, err := w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "AI Agent",
			Email: "ai-agent@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	_, err = repo.CommitObject(commit)
	if err != nil {
		return fmt.Errorf("failed to get commit object: %w", err)
	}

	return nil
}
