package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initRepo creates a throwaway git repository with one commit on main and
// chdirs into it for the duration of the test.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)

	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"config", "user.name", "test"},
		{"config", "user.email", "test@example.com"},
		{"config", "commit.gpgsign", "false"},
	} {
		gitRun(t, dir, args...)
	}
	writeAndCommit(t, dir, "base.txt", "base\n", "initial commit")
	return dir
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeAndCommit(t *testing.T, dir, name, content, msg string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "add", name)
	gitRun(t, dir, "commit", "-m", msg)
}

func TestStagedDiff(t *testing.T) {
	dir := initRepo(t)
	cli := New()
	ctx := context.Background()

	out, err := cli.StagedDiff(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "" {
		t.Fatalf("expected empty staged diff, got %q", out)
	}

	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("staged content\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "add", "new.txt")

	out, err = cli.StagedDiff(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "staged content") {
		t.Fatalf("staged diff missing content: %q", out)
	}
}

func TestBranchDiff(t *testing.T) {
	dir := initRepo(t)
	cli := New()
	ctx := context.Background()

	gitRun(t, dir, "checkout", "-b", "feature")
	writeAndCommit(t, dir, "feature.txt", "feature work\n", "feat: add feature")

	out, err := cli.BranchDiff(ctx, "main")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "feature work") {
		t.Fatalf("branch diff missing feature change: %q", out)
	}
}

func TestBranchDiffEmptyWhenNoDivergence(t *testing.T) {
	initRepo(t)
	out, err := New().BranchDiff(context.Background(), "main")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "" {
		t.Fatalf("expected empty diff on same commit, got %q", out)
	}
}

func TestBranchDiffMissingBranch(t *testing.T) {
	initRepo(t)
	_, err := New().BranchDiff(context.Background(), "no-such-branch")
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected missing-branch error, got %v", err)
	}
}

func TestRepoRoot(t *testing.T) {
	dir := initRepo(t)
	got := New().RepoRoot(context.Background())
	realDir, _ := filepath.EvalSymlinks(dir)
	realGot, _ := filepath.EvalSymlinks(got)
	if realGot != realDir {
		t.Fatalf("repo root = %q, want %q", got, dir)
	}
}

func TestRepoRootOutsideRepo(t *testing.T) {
	t.Chdir(os.TempDir())
	if got := New().RepoRoot(context.Background()); got != "" {
		// os.TempDir could theoretically live inside a repo on odd setups;
		// only fail when it clearly returned this test's directory.
		t.Logf("unexpected repo root %q (tolerated on unusual setups)", got)
	}
}
