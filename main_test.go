package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupEnv gives the test its own config home and a git repo as working
// directory.
func setupEnv(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	repo := t.TempDir()
	t.Chdir(repo)
	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"config", "user.name", "test"},
		{"config", "user.email", "test@example.com"},
		{"config", "commit.gpgsign", "false"},
	} {
		c := exec.Command("git", args...)
		c.Dir = repo
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

// fakeLLMServer returns an OpenAI-compatible endpoint answering every chat
// completion with content.
func fakeLLMServer(t *testing.T, content string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := `{"choices":[{"message":{"role":"assistant","content":` + jsonString(content) + `}}]}`
		_, _ = w.Write([]byte(resp))
	}))
	t.Cleanup(server.Close)
	return server
}

func jsonString(s string) string {
	var b bytes.Buffer
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func writeBackendConfig(t *testing.T, baseURL string) {
	t.Helper()
	dir := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "lazycommit")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "active_backend: openai-compatible\n" +
		"backends:\n" +
		"  openai-compatible:\n" +
		"    model: test-model\n" +
		"    api_key: test-key\n" +
		"    base_url: " + baseURL + "\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
}

func stage(t *testing.T, name, content string) {
	t.Helper()
	if err := os.WriteFile(name, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "add", name).CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
}

func TestCommitEndToEnd(t *testing.T) {
	setupEnv(t)
	server := fakeLLMServer(t, "# Suggestions\n1. feat: add login flow\n2. fix: handle empty diff\n- chore: tidy deps")
	writeBackendConfig(t, server.URL)
	stage(t, "file.txt", "hello\n")

	var stdout, stderr bytes.Buffer
	code := run([]string{"commit"}, &stdout, &stderr, strings.NewReader(""))
	if code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, stderr.String())
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	want := []string{"Suggestions", "feat: add login flow", "fix: handle empty diff", "chore: tidy deps"}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines %q, want %d", len(lines), lines, len(want))
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("line %d = %q, want %q (cleaning failed)", i, lines[i], want[i])
		}
	}
}

func TestCommitNoStagedChanges(t *testing.T) {
	setupEnv(t)

	var stdout, stderr bytes.Buffer
	code := run([]string{"commit"}, &stdout, &stderr, strings.NewReader(""))
	if code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "No staged changes to commit.") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestCommitWithoutModelHintsConfigSet(t *testing.T) {
	setupEnv(t)
	stage(t, "file.txt", "hello\n")

	var stdout, stderr bytes.Buffer
	code := run([]string{"commit"}, &stdout, &stderr, strings.NewReader(""))
	if code == 0 {
		t.Fatal("expected non-zero exit without configured model")
	}
	if !strings.Contains(stderr.String(), "lazycommit config set") {
		t.Fatalf("stderr should hint at config set: %q", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout must stay clean on errors, got %q", stdout.String())
	}
}

func TestCommitUnreachableEndpointExitsNonZero(t *testing.T) {
	setupEnv(t)
	server := fakeLLMServer(t, "unused")
	url := server.URL
	server.Close()
	writeBackendConfig(t, url)
	stage(t, "file.txt", "hello\n")

	var stdout, stderr bytes.Buffer
	code := run([]string{"commit"}, &stdout, &stderr, strings.NewReader(""))
	if code == 0 {
		t.Fatal("expected non-zero exit for unreachable endpoint")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout must stay clean on errors, got %q", stdout.String())
	}
}

func TestPREndToEnd(t *testing.T) {
	setupEnv(t)
	server := fakeLLMServer(t, "improve login flow\nrefactor auth module")
	writeBackendConfig(t, server.URL)

	stage(t, "base.txt", "base\n")
	if out, err := exec.Command("git", "commit", "-m", "initial").CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
	if out, err := exec.Command("git", "checkout", "-b", "feature").CombinedOutput(); err != nil {
		t.Fatalf("git checkout: %v\n%s", err, out)
	}
	stage(t, "feature.txt", "feature\n")
	if out, err := exec.Command("git", "commit", "-m", "feature work").CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}

	var stdout, stderr bytes.Buffer
	code := run([]string{"pr", "main"}, &stdout, &stderr, strings.NewReader(""))
	if code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "improve login flow") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestPRRequiresTargetBranch(t *testing.T) {
	setupEnv(t)

	var stdout, stderr bytes.Buffer
	code := run([]string{"pr"}, &stdout, &stderr, strings.NewReader(""))
	if code == 0 {
		t.Fatal("expected non-zero exit for missing target branch")
	}
	if !strings.Contains(stderr.String(), "arg") && !strings.Contains(stderr.String(), "Usage") {
		t.Fatalf("stderr should explain usage: %q", stderr.String())
	}
}

func TestConfigSetThenGet(t *testing.T) {
	setupEnv(t)

	input := "1\ngpt-4o\nhttp://localhost:11434/v1\nsk-secret-1234\nKorean\n"
	var stdout, stderr bytes.Buffer
	code := run([]string{"config", "set"}, &stdout, &stderr, strings.NewReader(input))
	if code != 0 {
		t.Fatalf("config set failed: %d, stderr: %s", code, stderr.String())
	}

	stdout.Reset()
	code = run([]string{"config", "get"}, &stdout, &stderr, strings.NewReader(""))
	if code != 0 {
		t.Fatalf("config get failed: %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"openai-compatible", "gpt-4o", "http://localhost:11434/v1", "Korean"} {
		if !strings.Contains(out, want) {
			t.Fatalf("config get output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "sk-secret-1234") {
		t.Fatalf("api key printed unmasked:\n%s", out)
	}
	if !strings.Contains(out, "****1234") {
		t.Fatalf("expected masked key ****1234:\n%s", out)
	}
}
