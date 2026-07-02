package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/m7medvision/lazycommit/internal/domain"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestLoadBackendsMissingFileGivesDefaults(t *testing.T) {
	r := NewRepository(filepath.Join(t.TempDir(), "lazycommit"), "")
	b, err := r.LoadBackends()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Active != "openai-compatible" {
		t.Fatalf("default backend = %q, want openai-compatible", b.Active)
	}
	if _, ok := b.Backends["openai-compatible"]; !ok {
		t.Fatal("default backend entry missing")
	}
}

func TestBackendsRoundTrip(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "lazycommit")
	r := NewRepository(dir, "")
	in := Backends{
		Active: "openai-compatible",
		Backends: map[string]BackendSettings{
			"openai-compatible": {Model: "gpt-4o", APIKey: "sk-plain", BaseURL: "http://localhost:11434/v1"},
		},
	}
	if err := r.SaveBackends(in); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("secret file permissions = %v, want 0600", info.Mode().Perm())
	}

	out, err := r.LoadBackends()
	if err != nil {
		t.Fatal(err)
	}
	if out.Active != "openai-compatible" {
		t.Fatalf("active = %q", out.Active)
	}
	got := out.Backends["openai-compatible"]
	if got.Model != "gpt-4o" || got.APIKey != "sk-plain" || got.BaseURL != "http://localhost:11434/v1" {
		t.Fatalf("round trip mangled settings: %+v", got)
	}
}

func TestLoadBackendsExpandsEnvKey(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "lazycommit")
	writeFile(t, filepath.Join(dir, "config.yaml"), `
active_backend: openai-compatible
backends:
  openai-compatible:
    model: gpt-4o
    api_key: "$TEST_LAZYCOMMIT_KEY"
`)
	r := NewRepository(dir, "")
	r.env = func(name string) string {
		if name == "TEST_LAZYCOMMIT_KEY" {
			return "expanded-secret"
		}
		return ""
	}

	b, err := r.LoadBackends()
	if err != nil {
		t.Fatal(err)
	}
	if b.Backends["openai-compatible"].APIKey != "expanded-secret" {
		t.Fatalf("api key not expanded: %+v", b.Backends["openai-compatible"])
	}
}

func TestLoadBackendsEmptyEnvKeyFails(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "lazycommit")
	writeFile(t, filepath.Join(dir, "config.yaml"), `
active_backend: openai-compatible
backends:
  openai-compatible:
    api_key: "$MISSING_VAR_FOR_TEST"
`)
	r := NewRepository(dir, "")
	r.env = func(string) string { return "" }

	_, err := r.LoadBackends()
	if err == nil || !strings.Contains(err.Error(), "MISSING_VAR_FOR_TEST") {
		t.Fatalf("expected error naming the variable, got %v", err)
	}
}

func TestPromptSettingsDefaultsWhenNoFiles(t *testing.T) {
	r := NewRepository(filepath.Join(t.TempDir(), "lazycommit"), t.TempDir())
	s, err := r.PromptSettings()
	if err != nil {
		t.Fatal(err)
	}
	if s.Language.String() != "English" {
		t.Fatalf("language = %q", s.Language)
	}
	if s.SuggestionCount != domain.DefaultSuggestionCount {
		t.Fatalf("count = %d", s.SuggestionCount)
	}
	if s.SystemMessage != domain.DefaultSystemMessage {
		t.Fatalf("system = %q", s.SystemMessage)
	}
}

func TestPromptSettingsLayering(t *testing.T) {
	globalDir := filepath.Join(t.TempDir(), "lazycommit")
	repoRoot := t.TempDir()
	writeFile(t, filepath.Join(globalDir, "prompts.yaml"), `
language: Spanish
system_message: global system
num_suggestions: 7
`)
	writeFile(t, filepath.Join(repoRoot, "lazycommit.prompts.yaml"), `
language: Korean
commit_message_template: "REPO %s"
`)

	s, err := NewRepository(globalDir, repoRoot).PromptSettings()
	if err != nil {
		t.Fatal(err)
	}
	if s.Language.String() != "Korean" {
		t.Fatalf("repo-local language should win, got %q", s.Language)
	}
	if s.SystemMessage != "global system" {
		t.Fatalf("global field should fall through, got %q", s.SystemMessage)
	}
	if s.SuggestionCount != 7 {
		t.Fatalf("global count should fall through, got %d", s.SuggestionCount)
	}
	if s.CommitTemplate.String() != "REPO %s" {
		t.Fatalf("repo-local template should win, got %q", s.CommitTemplate)
	}
	if s.PRTitleTemplate.String() != domain.DefaultPRTitleTemplate {
		t.Fatalf("unset field should use default, got %q", s.PRTitleTemplate)
	}
}

func TestPromptSettingsInvalidTemplate(t *testing.T) {
	globalDir := filepath.Join(t.TempDir(), "lazycommit")
	writeFile(t, filepath.Join(globalDir, "prompts.yaml"), `
commit_message_template: "no placeholder here"
`)
	_, err := NewRepository(globalDir, "").PromptSettings()
	if err == nil || !strings.Contains(err.Error(), "commit_message_template") {
		t.Fatalf("expected template validation error, got %v", err)
	}
}

func TestV1ConfigDetection(t *testing.T) {
	base := t.TempDir()
	globalDir := filepath.Join(base, "lazycommit")
	r := NewRepository(globalDir, "")
	if got := r.V1ConfigPath(); got != "" {
		t.Fatalf("expected no v1 config, got %q", got)
	}

	writeFile(t, filepath.Join(base, ".lazycommit.yaml"), "active_provider: opencode\n")
	if got := r.V1ConfigPath(); got == "" {
		t.Fatal("expected v1 config to be detected")
	}
}

func TestPromptsRoundTrip(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "lazycommit")
	r := NewRepository(dir, "")
	if err := r.SavePrompts(Prompts{Language: "Arabic", NumSuggestions: 5}); err != nil {
		t.Fatal(err)
	}
	p, err := r.LoadPrompts()
	if err != nil {
		t.Fatal(err)
	}
	if p.Language != "Arabic" || p.NumSuggestions != 5 {
		t.Fatalf("round trip mangled prompts: %+v", p)
	}
}
