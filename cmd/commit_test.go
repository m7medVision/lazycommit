package cmd

import "testing"

func TestApplyOutputOverrides(t *testing.T) {
	in := []string{"feat: add provider and model flags"}
	out := applyOutputOverrides(in, true)

	if len(out) != 1 {
		t.Fatalf("expected 1 message, got %d", len(out))
	}
	if out[0] != "✨ feat: add provider and model flags" {
		t.Fatalf("unexpected output: %q", out[0])
	}
}

func TestEnsureGitmojiPrefix(t *testing.T) {
	got := ensureGitmojiPrefix("fix: handle empty staged diff")
	want := "🐛 fix: handle empty staged diff"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestEnsureGitmojiPrefix_NoDoubleEmoji(t *testing.T) {
	msg := "✨ feat: add provider override"
	if got := ensureGitmojiPrefix(msg); got != msg {
		t.Fatalf("emoji should not be duplicated, got %q", got)
	}
}

func TestEnsureGitmojiPrefix_NonASCIIDescription(t *testing.T) {
	got := ensureGitmojiPrefix("feat: añade validación")
	want := "✨ feat: añade validación"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
