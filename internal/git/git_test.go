package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func runGit(t *testing.T, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
}

func TestHasChanges(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	runGit(t, "init")
	runGit(t, "config", "user.email", "test@example.com")
	runGit(t, "config", "user.name", "Test User")

	clean, err := HasChanges()
	if err != nil {
		t.Fatalf("HasChanges returned error: %v", err)
	}
	if clean {
		t.Fatalf("expected clean repo to report no changes")
	}

	if err := os.WriteFile(filepath.Join(tmp, "file.txt"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	dirty, err := HasChanges()
	if err != nil {
		t.Fatalf("HasChanges returned error: %v", err)
	}
	if !dirty {
		t.Fatalf("expected repo with untracked file to report changes")
	}
}

func TestStageAll(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	runGit(t, "init")
	runGit(t, "config", "user.email", "test@example.com")
	runGit(t, "config", "user.name", "Test User")

	if err := os.WriteFile(filepath.Join(tmp, "new.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatalf("write new file: %v", err)
	}

	if err := StageAll(); err != nil {
		t.Fatalf("StageAll returned error: %v", err)
	}

	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git diff --cached --name-only failed: %v\n%s", err, string(out))
	}
	if string(out) != "new.txt\n" {
		t.Fatalf("expected staged file new.txt, got %q", string(out))
	}
}
