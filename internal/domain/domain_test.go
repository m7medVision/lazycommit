package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNewDiff(t *testing.T) {
	if _, err := NewDiff("   \n\t"); !errors.Is(err, ErrEmptyDiff) {
		t.Fatalf("expected ErrEmptyDiff, got %v", err)
	}
	d, err := NewDiff("+added line")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.String() != "+added line" {
		t.Fatalf("diff text mangled: %q", d.String())
	}
}

func TestNewSuggestion(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"valid", "feat: add login", false},
		{"trims whitespace", "  fix: bug  ", false},
		{"empty", "   ", true},
		{"multiline", "feat: a\nfix: b", true},
		{"too long", strings.Repeat("x", MaxSuggestionLength+1), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewSuggestion(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("NewSuggestion(%q) error = %v, wantErr %v", tc.in, err, tc.wantErr)
			}
		})
	}
}

func TestNewLanguage(t *testing.T) {
	if got := NewLanguage("").String(); got != "English" {
		t.Fatalf("blank language should default to English, got %q", got)
	}
	if got := NewLanguage("  Arabic  ").String(); got != "Arabic" {
		t.Fatalf("expected Arabic, got %q", got)
	}
}

func TestNewModelID(t *testing.T) {
	if _, err := NewModelID(" "); err == nil {
		t.Fatal("expected error for blank model id")
	}
	m, err := NewModelID("gpt-4o")
	if err != nil || m.String() != "gpt-4o" {
		t.Fatalf("unexpected: %v %q", err, m.String())
	}
}

func TestNewPromptTemplate(t *testing.T) {
	if _, err := NewPromptTemplate(""); err == nil {
		t.Fatal("expected error for empty template")
	}
	if _, err := NewPromptTemplate("no placeholder"); err == nil {
		t.Fatal("expected error for missing placeholder")
	}
	if _, err := NewPromptTemplate("two %s and %s"); err == nil {
		t.Fatal("expected error for duplicate placeholder")
	}
	if _, err := NewPromptTemplate("diff: %s"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPromptBuilderDefaults(t *testing.T) {
	diff, _ := NewDiff("+x")
	p := NewPromptBuilder().Build(diff)

	if p.System != DefaultSystemMessage {
		t.Fatalf("unexpected system message: %q", p.System)
	}
	if !strings.Contains(p.User, "+x") {
		t.Fatal("diff not embedded in user prompt")
	}
	if !strings.Contains(p.User, "Generate exactly 10 suggestions.") {
		t.Fatalf("default count missing: %q", p.User)
	}
	if !strings.Contains(p.User, "in English.") {
		t.Fatalf("default language missing: %q", p.User)
	}
}

func TestPromptBuilderOverrides(t *testing.T) {
	diff, _ := NewDiff("+x")
	tmpl, _ := NewPromptTemplate("CUSTOM %s CUSTOM")
	p := NewPromptBuilder().
		WithSystemMessage("sys").
		WithTemplate(tmpl).
		WithLanguage(NewLanguage("Korean")).
		WithSuggestionCount(5).
		Build(diff)

	if p.System != "sys" {
		t.Fatalf("system override lost: %q", p.System)
	}
	if !strings.HasPrefix(p.User, "CUSTOM +x CUSTOM") {
		t.Fatalf("template override lost: %q", p.User)
	}
	if !strings.Contains(p.User, "exactly 5 suggestions") {
		t.Fatalf("count override lost: %q", p.User)
	}
	if !strings.Contains(p.User, "in Korean.") {
		t.Fatalf("language override lost: %q", p.User)
	}
}

func TestPromptBuilderIgnoresInvalidOverrides(t *testing.T) {
	diff, _ := NewDiff("+x")
	p := NewPromptBuilder().
		WithSystemMessage("   ").
		WithSuggestionCount(-3).
		Build(diff)

	if p.System != DefaultSystemMessage {
		t.Fatal("blank system message should keep default")
	}
	if !strings.Contains(p.User, "exactly 10 suggestions") {
		t.Fatal("non-positive count should keep default")
	}
}

func TestParseSuggestions(t *testing.T) {
	raw := strings.Join([]string{
		"```",
		"# Suggestions",
		"1. feat: add user authentication",
		"2) fix: handle empty diff",
		"- chore: update dependencies",
		"* docs: improve readme",
		"",
		"   ",
		strings.Repeat("y", 300),
		"plain: no prefix at all",
		"```",
	}, "\n")

	got := ParseSuggestions(raw, 10)
	want := []string{
		"Suggestions",
		"feat: add user authentication",
		"fix: handle empty diff",
		"chore: update dependencies",
		"docs: improve readme",
		"plain: no prefix at all",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d suggestions %v, want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i].String() != want[i] {
			t.Fatalf("suggestion %d = %q, want %q", i, got[i].String(), want[i])
		}
	}
}

func TestParseSuggestionsCapsAtMax(t *testing.T) {
	raw := "one\ntwo\nthree\nfour"
	got := ParseSuggestions(raw, 2)
	if len(got) != 2 {
		t.Fatalf("expected cap at 2, got %d", len(got))
	}
	if got[0].String() != "one" || got[1].String() != "two" {
		t.Fatalf("wrong order: %v", got)
	}
}

func TestParseSuggestionsZeroMax(t *testing.T) {
	if got := ParseSuggestions("one\ntwo", 0); got != nil {
		t.Fatalf("expected nil for max<=0, got %v", got)
	}
}
