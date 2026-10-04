package domain

import (
	"errors"
	"strconv"
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

// fileDiffText builds a minimal git diff section for path with n added lines.
func fileDiffText(path string, n int) string {
	var b strings.Builder
	b.WriteString("diff --git a/" + path + " b/" + path + "\n")
	b.WriteString("--- a/" + path + "\n+++ b/" + path + "\n")
	b.WriteString("@@ -0,0 +1," + strconv.Itoa(n) + " @@\n")
	for i := 0; i < n; i++ {
		b.WriteString("+line " + strconv.Itoa(i) + " of " + path + "\n")
	}
	return b.String()
}

func TestCompactKeepsSmallDiff(t *testing.T) {
	raw := fileDiffText("main.go", 3)
	d, _ := NewDiff(raw)
	if got := d.Compact(len(raw)).String(); got != raw {
		t.Fatalf("diff within budget must be unchanged, got %q", got)
	}
	if got := d.Compact(0).String(); got != raw {
		t.Fatalf("non-positive budget must keep diff unchanged, got %q", got)
	}
}

func TestCompactSkipsGeneratedAndBinaryFiles(t *testing.T) {
	binary := "diff --git a/logo.png b/logo.png\nindex 1..2 100644\nBinary files a/logo.png and b/logo.png differ\n"
	raw := fileDiffText("main.go", 3) + fileDiffText("go.sum", 500) + binary + fileDiffText("web/app.min.js", 200)
	d, _ := NewDiff(raw)
	got := d.Compact(2000).String()

	for _, want := range []string{
		"Changed files (4):",
		"main.go +3 -0\n",
		"go.sum +500 -0 (generated, diff omitted)",
		"logo.png (binary, diff omitted)",
		"web/app.min.js +200 -0 (generated, diff omitted)",
		fileDiffText("main.go", 3),
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("compacted diff misses %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"+line 0 of go.sum", "Binary files", "+line 0 of web/app.min.js"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("compacted diff should not contain %q:\n%s", unwanted, got)
		}
	}
}

func TestCompactFitsBudget(t *testing.T) {
	small := fileDiffText("a.go", 5)
	big := fileDiffText("b.go", 1000)
	raw := small + big + fileDiffText("c.go", 5)
	d, _ := NewDiff(raw)
	budget := 3000
	got := d.Compact(budget).String()

	if len(got) > budget {
		t.Fatalf("compacted diff is %d chars, budget %d", len(got), budget)
	}
	if !strings.Contains(got, small) {
		t.Fatal("file that fits must be kept in full")
	}
	if !strings.Contains(got, "+line 0 of b.go\n") || !strings.HasSuffix(got, "\n[truncated]\n") {
		t.Fatalf("file over the budget must be cut and marked:\n%s", got)
	}
	if !strings.Contains(got, fileDiffText("c.go", 5)) {
		t.Fatal("a later file that fits must be kept in full after skipping a large one")
	}
	if !strings.Contains(got, "b.go +1000 -0\n") {
		t.Fatal("summary must list every file")
	}
}

func TestCompactRaisesTinyBudget(t *testing.T) {
	raw := fileDiffText("a.go", 1000)
	d, _ := NewDiff(raw)
	got := d.Compact(1).String()
	if len(got) > MinDiffBudget || !strings.Contains(got, "Changed files (1):") {
		t.Fatalf("tiny budget must be raised to %d, got %d chars:\n%s", MinDiffBudget, len(got), got)
	}
}

func TestCompactDetectsGitBinaryPatch(t *testing.T) {
	binary := "diff --git a/logo.png b/logo.png\nindex 1..2 100644\nGIT binary patch\nliteral 9000\n" +
		strings.Repeat("zcmeHM\n", 1000)
	d, _ := NewDiff(binary + fileDiffText("main.go", 3))
	got := d.Compact(MinDiffBudget).String()
	if !strings.Contains(got, "logo.png (binary, diff omitted)") || strings.Contains(got, "zcmeHM") {
		t.Fatalf("GIT binary patch must be treated as binary:\n%s", got)
	}
}

func TestCompactParsesPaths(t *testing.T) {
	quoted := "diff --git \"a/d\\303\\266c/go.sum\" \"b/d\\303\\266c/go.sum\"\n" +
		"--- \"a/d\\303\\266c/go.sum\"\n+++ \"b/d\\303\\266c/go.sum\"\n@@ -0,0 +1 @@\n+x\n"
	spaced := "diff --git a/dir b/yarn.lock b/dir b/yarn.lock\n" +
		"--- a/dir b/yarn.lock\t\n+++ b/dir b/yarn.lock\t\n@@ -1 +1 @@\n-x\n+y\n"
	deleted := "diff --git a/old.go b/old.go\ndeleted file mode 100644\n" +
		"--- a/old.go\n+++ /dev/null\n@@ -1 +0,0 @@\n-x\n"
	binary := "diff --git a/img b/pic.png b/img b/pic.png\nindex 1..2 100644\n" +
		"Binary files a/img b/pic.png and b/img b/pic.png differ\n"
	renamed := "diff --git a/old b/x.go b/new b/x.go\nsimilarity index 100%\n" +
		"rename from old b/x.go\nrename to new b/x.go\n"
	raw := quoted + spaced + deleted + binary + renamed + fileDiffText("linker.map", 2) +
		fileDiffText("app.js.map", 2) + fileDiffText("big.go", 1000)
	d, _ := NewDiff(raw)
	got := d.Compact(MinDiffBudget).String()

	for _, want := range []string{
		"döc/go.sum +1 -0 (generated, diff omitted)",
		"dir b/yarn.lock +1 -1 (generated, diff omitted)",
		"old.go +0 -1\n",
		"img b/pic.png (binary, diff omitted)",
		"new b/x.go +0 -0\n",
		"linker.map +2 -0\n",
		"app.js.map +2 -0 (generated, diff omitted)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("summary misses %q:\n%s", want, got)
		}
	}
}

func TestCompactCutsLongFileList(t *testing.T) {
	var raw strings.Builder
	for i := 0; i < 500; i++ {
		raw.WriteString(fileDiffText("pkg/file"+strconv.Itoa(i)+".go", 1))
	}
	d, _ := NewDiff(raw.String())
	budget := 1000
	got := d.Compact(budget).String()

	if len(got) > budget {
		t.Fatalf("compacted diff is %d chars, budget %d", len(got), budget)
	}
	if !strings.Contains(got, "more files\n") {
		t.Fatalf("long file list must be cut:\n%s", got)
	}
}

func TestCompactNonGitText(t *testing.T) {
	for name, raw := range map[string]string{
		"lines":       strings.Repeat("+plain line\n", 200),
		"single line": strings.Repeat("x", 5000),
	} {
		t.Run(name, func(t *testing.T) {
			d, _ := NewDiff(raw)
			got := d.Compact(MinDiffBudget).String()
			if len(got) > MinDiffBudget || !strings.HasSuffix(got, "\n[truncated]\n") || len(got) < 100 {
				t.Fatalf("text without file headers must be cut to the budget: %q", got)
			}
		})
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
