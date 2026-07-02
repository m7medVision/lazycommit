package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/m7medvision/lazycommit/v2/internal/domain"
)

// SuggestionsResult carries generated suggestions, or NoChanges when the
// relevant diff was empty and no backend was invoked.
type SuggestionsResult struct {
	Suggestions []domain.Suggestion
	NoChanges   bool
}

// GenerateCommitSuggestions produces commit message suggestions from the
// staged diff.
type GenerateCommitSuggestions struct {
	pipeline suggestionPipeline
}

func NewGenerateCommitSuggestions(gen Generator, diffs DiffSource, cfg ConfigRepository) *GenerateCommitSuggestions {
	return &GenerateCommitSuggestions{pipeline: suggestionPipeline{gen: gen, diffs: diffs, cfg: cfg}}
}

func (uc *GenerateCommitSuggestions) Execute(ctx context.Context) (SuggestionsResult, error) {
	return uc.pipeline.run(ctx, func(ctx context.Context, diffs DiffSource) (string, error) {
		return diffs.StagedDiff(ctx)
	}, func(s PromptSettings) domain.PromptTemplate {
		return s.CommitTemplate
	})
}

// GeneratePRTitles produces pull request title suggestions from the diff
// between the current branch and a target branch.
type GeneratePRTitles struct {
	pipeline suggestionPipeline
}

func NewGeneratePRTitles(gen Generator, diffs DiffSource, cfg ConfigRepository) *GeneratePRTitles {
	return &GeneratePRTitles{pipeline: suggestionPipeline{gen: gen, diffs: diffs, cfg: cfg}}
}

func (uc *GeneratePRTitles) Execute(ctx context.Context, target string) (SuggestionsResult, error) {
	if target == "" {
		return SuggestionsResult{}, errors.New("target branch is required")
	}
	return uc.pipeline.run(ctx, func(ctx context.Context, diffs DiffSource) (string, error) {
		return diffs.BranchDiff(ctx, target)
	}, func(s PromptSettings) domain.PromptTemplate {
		return s.PRTitleTemplate
	})
}

// suggestionPipeline is the shared flow: read diff, short-circuit when
// empty, build prompt, generate, parse. Commit and PR generation differ
// only in diff source and template.
type suggestionPipeline struct {
	gen   Generator
	diffs DiffSource
	cfg   ConfigRepository
}

func (p suggestionPipeline) run(
	ctx context.Context,
	readDiff func(context.Context, DiffSource) (string, error),
	pickTemplate func(PromptSettings) domain.PromptTemplate,
) (SuggestionsResult, error) {
	raw, err := readDiff(ctx, p.diffs)
	if err != nil {
		return SuggestionsResult{}, fmt.Errorf("reading diff: %w", err)
	}

	diff, err := domain.NewDiff(raw)
	if errors.Is(err, domain.ErrEmptyDiff) {
		return SuggestionsResult{NoChanges: true}, nil
	}
	if err != nil {
		return SuggestionsResult{}, err
	}

	settings, err := p.cfg.PromptSettings()
	if err != nil {
		return SuggestionsResult{}, fmt.Errorf("loading configuration: %w", err)
	}

	prompt := domain.NewPromptBuilder().
		WithSystemMessage(settings.SystemMessage).
		WithTemplate(pickTemplate(settings)).
		WithLanguage(settings.Language).
		WithSuggestionCount(settings.SuggestionCount).
		Build(diff)

	output, err := p.gen.Generate(ctx, prompt)
	if err != nil {
		return SuggestionsResult{}, fmt.Errorf("generating suggestions: %w", err)
	}

	count := settings.SuggestionCount
	if count <= 0 {
		count = domain.DefaultSuggestionCount
	}
	suggestions := domain.ParseSuggestions(output, count)
	if len(suggestions) == 0 {
		return SuggestionsResult{}, errors.New("backend returned no usable suggestions")
	}
	return SuggestionsResult{Suggestions: suggestions}, nil
}
