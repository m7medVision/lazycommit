// Package config implements the ConfigRepository port with a two-file split:
// secret backend settings live only in the global config directory, while
// shareable prompt settings layer repo-local over global over defaults.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/m7medvision/lazycommit/v2/internal/app"
	"github.com/m7medvision/lazycommit/v2/internal/domain"
)

const (
	// DefaultBackend is the only backend family lazycommit talks to: any
	// endpoint speaking the OpenAI chat-completions protocol.
	DefaultBackend = "openai-compatible"

	backendsFile    = "config.yaml"
	promptsFile     = "prompts.yaml"
	repoPromptsFile = "lazycommit.prompts.yaml"
	filePermissions = 0o600
	dirPermissions  = 0o755
)

// BackendSettings configures one backend; fields a backend does not use are
// left empty.
type BackendSettings struct {
	Model          string   `yaml:"model,omitempty"`
	FallbackModels []string `yaml:"fallback_models,omitempty"`
	APIKey         string   `yaml:"api_key,omitempty"`
	BaseURL        string   `yaml:"base_url,omitempty"`
}

// Backends is the secret half of the configuration (global only).
type Backends struct {
	Active   string                     `yaml:"active_backend"`
	Backends map[string]BackendSettings `yaml:"backends"`
}

// Prompts is the shareable half; zero values mean "unset, fall through".
type Prompts struct {
	Language              string `yaml:"language,omitempty"`
	SystemMessage         string `yaml:"system_message,omitempty"`
	CommitMessageTemplate string `yaml:"commit_message_template,omitempty"`
	PRTitleTemplate       string `yaml:"pr_title_template,omitempty"`
	NumSuggestions        int    `yaml:"num_suggestions,omitempty"`
}

// DefaultBackends is the effective configuration when no file exists; the
// model and key are unset until `lazycommit config set` provides them.
func DefaultBackends() Backends {
	return Backends{
		Active: DefaultBackend,
		Backends: map[string]BackendSettings{
			DefaultBackend: {},
		},
	}
}

// Repository reads and writes both configuration files. It implements
// app.ConfigRepository.
type Repository struct {
	globalDir string
	repoRoot  string
	env       func(string) string
}

// DefaultGlobalDir is <user config dir>/lazycommit.
func DefaultGlobalDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolving user config dir: %w", err)
	}
	return filepath.Join(base, "lazycommit"), nil
}

// NewRepository creates a repository rooted at globalDir; repoRoot (the
// working repository, may be empty) is searched for prompt overrides.
func NewRepository(globalDir, repoRoot string) *Repository {
	return &Repository{globalDir: globalDir, repoRoot: repoRoot, env: os.Getenv}
}

// LoadBackendsRaw returns the saved backend configuration exactly as
// written (no secret expansion), or defaults when the file does not exist.
// Use it when editing and re-saving so expanded secrets never hit disk.
func (r *Repository) LoadBackendsRaw() (Backends, error) {
	var b Backends
	ok, err := readYAML(filepath.Join(r.globalDir, backendsFile), &b)
	if err != nil {
		return Backends{}, err
	}
	if !ok {
		return DefaultBackends(), nil
	}
	if b.Active == "" {
		b.Active = DefaultBackend
	}
	if b.Backends == nil {
		b.Backends = map[string]BackendSettings{}
	}
	return b, nil
}

// LoadBackends returns the saved backend configuration, or defaults when the
// file does not exist. API keys of the form $VAR are expanded from the
// environment.
func (r *Repository) LoadBackends() (Backends, error) {
	b, err := r.LoadBackendsRaw()
	if err != nil {
		return Backends{}, err
	}
	for name, settings := range b.Backends {
		expanded, err := r.expandSecret(settings.APIKey)
		if err != nil {
			return Backends{}, fmt.Errorf("backend %q: %w", name, err)
		}
		settings.APIKey = expanded
		b.Backends[name] = settings
	}
	return b, nil
}

// SaveBackends writes the global backend configuration with owner-only
// permissions.
func (r *Repository) SaveBackends(b Backends) error {
	return writeYAML(filepath.Join(r.globalDir, backendsFile), b)
}

// SavePrompts writes the global prompt configuration.
func (r *Repository) SavePrompts(p Prompts) error {
	return writeYAML(filepath.Join(r.globalDir, promptsFile), p)
}

// LoadGlobalPrompts returns only the global prompt file, for editing via
// `config set` without folding repo-local overrides into it.
func (r *Repository) LoadGlobalPrompts() (Prompts, error) {
	var global Prompts
	if _, err := readYAML(filepath.Join(r.globalDir, promptsFile), &global); err != nil {
		return Prompts{}, err
	}
	return global, nil
}

// LoadPrompts returns the raw layered prompt files (repo over global),
// without applying defaults.
func (r *Repository) LoadPrompts() (Prompts, error) {
	global, err := r.LoadGlobalPrompts()
	if err != nil {
		return Prompts{}, err
	}
	if r.repoRoot == "" {
		return global, nil
	}
	var local Prompts
	if _, err := readYAML(filepath.Join(r.repoRoot, repoPromptsFile), &local); err != nil {
		return Prompts{}, err
	}
	return mergePrompts(local, global), nil
}

// PromptSettings implements app.ConfigRepository: the fully layered,
// validated effective settings.
func (r *Repository) PromptSettings() (app.PromptSettings, error) {
	p, err := r.LoadPrompts()
	if err != nil {
		return app.PromptSettings{}, err
	}

	commitText := p.CommitMessageTemplate
	if commitText == "" {
		commitText = domain.DefaultCommitTemplate
	}
	commit, err := domain.NewPromptTemplate(commitText)
	if err != nil {
		return app.PromptSettings{}, fmt.Errorf("commit_message_template: %w", err)
	}

	prText := p.PRTitleTemplate
	if prText == "" {
		prText = domain.DefaultPRTitleTemplate
	}
	pr, err := domain.NewPromptTemplate(prText)
	if err != nil {
		return app.PromptSettings{}, fmt.Errorf("pr_title_template: %w", err)
	}

	system := p.SystemMessage
	if system == "" {
		system = domain.DefaultSystemMessage
	}
	count := p.NumSuggestions
	if count <= 0 {
		count = domain.DefaultSuggestionCount
	}

	return app.PromptSettings{
		SystemMessage:   system,
		CommitTemplate:  commit,
		PRTitleTemplate: pr,
		Language:        domain.NewLanguage(p.Language),
		SuggestionCount: count,
	}, nil
}

// V1ConfigPath returns the path of a leftover v1 configuration file so the
// CLI can hint that v2 uses a new format, or "" when none exists.
func (r *Repository) V1ConfigPath() string {
	base := filepath.Dir(r.globalDir)
	for _, name := range []string{".lazycommit.yaml", ".lazycommit.prompts.yaml"} {
		p := filepath.Join(base, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// HasBackendsFile reports whether a v2 backend configuration was saved.
func (r *Repository) HasBackendsFile() bool {
	_, err := os.Stat(filepath.Join(r.globalDir, backendsFile))
	return err == nil
}

func (r *Repository) expandSecret(value string) (string, error) {
	if !strings.HasPrefix(value, "$") {
		return value, nil
	}
	name := strings.TrimPrefix(value, "$")
	v := r.env(name)
	if v == "" {
		return "", fmt.Errorf("environment variable %s is not set", name)
	}
	return v, nil
}

// mergePrompts overlays every set field of top onto bottom.
func mergePrompts(top, bottom Prompts) Prompts {
	out := bottom
	if top.Language != "" {
		out.Language = top.Language
	}
	if top.SystemMessage != "" {
		out.SystemMessage = top.SystemMessage
	}
	if top.CommitMessageTemplate != "" {
		out.CommitMessageTemplate = top.CommitMessageTemplate
	}
	if top.PRTitleTemplate != "" {
		out.PRTitleTemplate = top.PRTitleTemplate
	}
	if top.NumSuggestions > 0 {
		out.NumSuggestions = top.NumSuggestions
	}
	return out
}

// readYAML reports whether the file existed.
func readYAML(path string, v any) (bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reading %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, v); err != nil {
		return false, fmt.Errorf("parsing %s: %w", path, err)
	}
	return true, nil
}

func writeYAML(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), dirPermissions); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, filePermissions); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}
