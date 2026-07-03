package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/m7medvision/lazycommit/internal/domain"
)

type fakeDiffSource struct {
	staged     string
	stagedErr  error
	branch     string
	branchErr  error
	lastTarget string
}

func (f *fakeDiffSource) StagedDiff(context.Context) (string, error) {
	return f.staged, f.stagedErr
}

func (f *fakeDiffSource) BranchDiff(_ context.Context, target string) (string, error) {
	f.lastTarget = target
	return f.branch, f.branchErr
}

type fakeGenerator struct {
	output     string
	err        error
	lastPrompt domain.Prompt
	calls      int
}

func (f *fakeGenerator) Generate(_ context.Context, p domain.Prompt) (string, error) {
	f.calls++
	f.lastPrompt = p
	return f.output, f.err
}

type fakeConfig struct {
	settings PromptSettings
	err      error
}

func (f *fakeConfig) PromptSettings() (PromptSettings, error) {
	return f.settings, f.err
}

func testSettings(t *testing.T) PromptSettings {
	t.Helper()
	commit, err := domain.NewPromptTemplate("COMMIT %s")
	if err != nil {
		t.Fatal(err)
	}
	pr, err := domain.NewPromptTemplate("PR %s")
	if err != nil {
		t.Fatal(err)
	}
	return PromptSettings{
		SystemMessage:   "sys",
		CommitTemplate:  commit,
		PRTitleTemplate: pr,
		Language:        domain.NewLanguage("English"),
		SuggestionCount: 3,
	}
}

func TestCommitSuggestionsHappyPath(t *testing.T) {
	gen := &fakeGenerator{output: "1. feat: one\n2. fix: two\n\nthree\nfour"}
	uc := NewGenerateCommitSuggestions(gen,
		&fakeDiffSource{staged: "+change"},
		&fakeConfig{settings: testSettings(t)})

	res, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NoChanges {
		t.Fatal("unexpected NoChanges")
	}
	got := make([]string, len(res.Suggestions))
	for i, s := range res.Suggestions {
		got[i] = s.String()
	}
	want := []string{"feat: one", "fix: two", "three"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("suggestions = %v, want %v (count capped at 3)", got, want)
	}
	if !strings.HasPrefix(gen.lastPrompt.User, "COMMIT +change") {
		t.Fatalf("commit template not used: %q", gen.lastPrompt.User)
	}
	if gen.lastPrompt.System != "sys" {
		t.Fatalf("system message not applied: %q", gen.lastPrompt.System)
	}
}

func TestCommitSuggestionsEmptyDiffShortCircuits(t *testing.T) {
	gen := &fakeGenerator{output: "unused"}
	uc := NewGenerateCommitSuggestions(gen,
		&fakeDiffSource{staged: "   \n"},
		&fakeConfig{settings: testSettings(t)})

	res, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.NoChanges {
		t.Fatal("expected NoChanges for empty staged diff")
	}
	if gen.calls != 0 {
		t.Fatalf("generator must not be called on empty diff, got %d calls", gen.calls)
	}
}

func TestCommitSuggestionsGeneratorFailure(t *testing.T) {
	uc := NewGenerateCommitSuggestions(
		&fakeGenerator{err: errors.New("backend down")},
		&fakeDiffSource{staged: "+change"},
		&fakeConfig{settings: testSettings(t)})

	if _, err := uc.Execute(context.Background()); err == nil || !strings.Contains(err.Error(), "backend down") {
		t.Fatalf("expected wrapped backend error, got %v", err)
	}
}

func TestCommitSuggestionsNoUsableOutput(t *testing.T) {
	uc := NewGenerateCommitSuggestions(
		&fakeGenerator{output: "\n\n```\n```\n"},
		&fakeDiffSource{staged: "+change"},
		&fakeConfig{settings: testSettings(t)})

	if _, err := uc.Execute(context.Background()); err == nil {
		t.Fatal("expected error when output parses to nothing")
	}
}

func TestCommitSuggestionsConfigFailure(t *testing.T) {
	uc := NewGenerateCommitSuggestions(
		&fakeGenerator{output: "x"},
		&fakeDiffSource{staged: "+change"},
		&fakeConfig{err: errors.New("bad yaml")})

	if _, err := uc.Execute(context.Background()); err == nil || !strings.Contains(err.Error(), "bad yaml") {
		t.Fatalf("expected wrapped config error, got %v", err)
	}
}

func TestPRTitlesUsesTargetAndTemplate(t *testing.T) {
	gen := &fakeGenerator{output: "title one\ntitle two"}
	diffs := &fakeDiffSource{branch: "+branch change"}
	uc := NewGeneratePRTitles(gen, diffs, &fakeConfig{settings: testSettings(t)})

	res, err := uc.Execute(context.Background(), "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diffs.lastTarget != "main" {
		t.Fatalf("target = %q, want main", diffs.lastTarget)
	}
	if !strings.HasPrefix(gen.lastPrompt.User, "PR +branch change") {
		t.Fatalf("pr template not used: %q", gen.lastPrompt.User)
	}
	if len(res.Suggestions) != 2 {
		t.Fatalf("expected 2 titles, got %d", len(res.Suggestions))
	}
}

func TestPRTitlesRequiresTarget(t *testing.T) {
	uc := NewGeneratePRTitles(&fakeGenerator{}, &fakeDiffSource{}, &fakeConfig{settings: testSettings(t)})
	if _, err := uc.Execute(context.Background(), ""); err == nil {
		t.Fatal("expected error for missing target branch")
	}
}

func TestPRTitlesEmptyBranchDiff(t *testing.T) {
	gen := &fakeGenerator{}
	uc := NewGeneratePRTitles(gen, &fakeDiffSource{branch: ""}, &fakeConfig{settings: testSettings(t)})

	res, err := uc.Execute(context.Background(), "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.NoChanges {
		t.Fatal("expected NoChanges for empty branch diff")
	}
	if gen.calls != 0 {
		t.Fatal("generator must not be called on empty diff")
	}
}
