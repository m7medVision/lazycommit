package git

import (
	"bytes"
	"fmt"

	"os/exec"
)

// GetStagedDiff returns the diff of the staged files.
func GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error running git diff --cached: %w", err)
	}
	return out.String(), nil
}

// GetWorkingTreeDiff returns the diff of the working tree.
func GetWorkingTreeDiff() (string, error) {
	cmd := exec.Command("git", "diff")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error running git diff: %w", err)
	}
	return out.String(), nil
}
