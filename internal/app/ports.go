// Package app holds the application use cases and the ports they depend on.
// Ports are consumer-owned: adapters in internal/llm, internal/git, and
// internal/config implement them; nothing here imports those packages.
package app

import (
	"context"

	"github.com/m7medvision/lazycommit/internal/domain"
)

// Generator produces raw LLM output for a prompt. The commit/PR distinction
// lives in the prompt, not in the interface.
type Generator interface {
	Generate(ctx context.Context, prompt domain.Prompt) (string, error)
}

// DiffSource supplies raw diffs from version control.
type DiffSource interface {
	StagedDiff(ctx context.Context) (string, error)
	BranchDiff(ctx context.Context, target string) (string, error)
}

// PromptSettings is the effective prompt configuration after layering.
type PromptSettings struct {
	SystemMessage   string
	CommitTemplate  domain.PromptTemplate
	PRTitleTemplate domain.PromptTemplate
	Language        domain.Language
	SuggestionCount int
}

// ConfigRepository yields the effective settings the use cases need.
type ConfigRepository interface {
	PromptSettings() (PromptSettings, error)
}
