// Package git implements the DiffSource port by shelling out to the git CLI.
package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CLI reads diffs from the repository containing the working directory.
type CLI struct{}

func New() *CLI {
	return &CLI{}
}

// StagedDiff returns `git diff --cached`; empty output means nothing staged.
func (c *CLI) StagedDiff(ctx context.Context) (string, error) {
	return run(ctx, "diff", "--cached")
}

// BranchDiff returns the merge-base diff between target and HEAD — the same
// changes a pull request against target would show.
func (c *CLI) BranchDiff(ctx context.Context, target string) (string, error) {
	if _, err := run(ctx, "rev-parse", "--verify", target); err != nil {
		return "", fmt.Errorf("branch %q does not exist", target)
	}
	return run(ctx, "diff", target+"...HEAD")
}

// RepoRoot returns the repository top-level directory, or "" when the
// working directory is not inside a git repository.
func (c *CLI) RepoRoot(ctx context.Context) string {
	out, err := run(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.String(), nil
}
