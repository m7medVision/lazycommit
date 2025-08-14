package git

import (
	"bytes"
	"fmt"

	"github.com/go-git/go-git/v5"
)

// GetStagedDiff returns the diff of the staged files.
func GetStagedDiff() (string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("error opening repository: %w", err)
	}

	// Get the worktree to access the staging area
	w, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("error getting worktree: %w", err)
	}

	// Get the status to see what's staged
	status, err := w.Status()
	if err != nil {
		return "", fmt.Errorf("error getting status: %w", err)
	}

	// Create a buffer to store the diff
	var diff bytes.Buffer

	// For each staged file, get its diff
	for path, change := range status {
		if change.Staging != git.Unmodified {
			// For a more complete implementation, we would compare the staged version
			// with the HEAD version, but for now we'll just note that files are staged
			diff.WriteString(fmt.Sprintf("Staged file: %s\n", path))
		}
	}

	return diff.String(), nil
}

// GetWorkingTreeDiff returns the diff of the working tree.
func GetWorkingTreeDiff() (string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("error opening repository: %w", err)
	}

	// Get the worktree
	w, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("error getting worktree: %w", err)
	}

	// Get the status
	status, err := w.Status()
	if err != nil {
		return "", fmt.Errorf("error getting status: %w", err)
	}

	// Create a buffer for the diff
	var diff bytes.Buffer

	// For each modified file, get its diff
	for path, change := range status {
		if change.Worktree != git.Unmodified {
			// Note that files are modified
			diff.WriteString(fmt.Sprintf("Modified file: %s\n", path))
		}
	}

	return diff.String(), nil
}
