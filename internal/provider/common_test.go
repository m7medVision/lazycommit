package provider

import (
	"strings"
	"testing"
)

func TestParseOutputLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLines int
		expected []string
	}{
		{
			name:     "plain lines",
			input:    "feat: add login\nfix: typo in readme\nchore: update deps",
			maxLines: 10,
			expected: []string{"feat: add login", "fix: typo in readme", "chore: update deps"},
		},
		{
			name:     "single-digit numbered list with dot",
			input:    "1. feat: add login\n2. fix: typo\n3. chore: deps",
			maxLines: 10,
			expected: []string{"feat: add login", "fix: typo", "chore: deps"},
		},
		{
			name:     "single-digit numbered list with paren",
			input:    "1) feat: add login\n2) fix: typo",
			maxLines: 10,
			expected: []string{"feat: add login", "fix: typo"},
		},
		{
			name:     "multi-digit numbered list",
			input:    "10. feat: add login\n11. fix: typo\n12. chore: deps",
			maxLines: 10,
			expected: []string{"feat: add login", "fix: typo", "chore: deps"},
		},
		{
			name:     "markdown bullet dashes",
			input:    "- feat: add login\n- fix: typo",
			maxLines: 10,
			expected: []string{"feat: add login", "fix: typo"},
		},
		{
			name:     "markdown bullet asterisks",
			input:    "* feat: add login\n* fix: typo",
			maxLines: 10,
			expected: []string{"feat: add login", "fix: typo"},
		},
		{
			name:     "markdown headings stripped",
			input:    "# feat: add login\n## fix: typo",
			maxLines: 10,
			expected: []string{"feat: add login", "fix: typo"},
		},
		{
			name:     "empty lines skipped",
			input:    "feat: add login\n\n\nfix: typo\n\n",
			maxLines: 10,
			expected: []string{"feat: add login", "fix: typo"},
		},
		{
			name:     "lines over 200 chars skipped",
			input:    "feat: add login\n" + strings.Repeat("x", 201) + "\nfix: typo",
			maxLines: 10,
			expected: []string{"feat: add login", "fix: typo"},
		},
		{
			name:     "respects maxLines limit",
			input:    "line1\nline2\nline3\nline4\nline5",
			maxLines: 3,
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "whitespace-only input",
			input:    "   \n  \n\t\n",
			maxLines: 10,
			expected: nil,
		},
		{
			name:     "mixed formatting",
			input:    "1. feat: login\n- fix: typo\n* chore: deps\n## docs: readme\nplain message",
			maxLines: 10,
			expected: []string{"feat: login", "fix: typo", "chore: deps", "docs: readme", "plain message"},
		},
		{
			name:     "number at start but no list separator",
			input:    "3rd attempt at fixing auth",
			maxLines: 10,
			expected: []string{"3rd attempt at fixing auth"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOutputLines(tt.input, tt.maxLines)
			if len(got) != len(tt.expected) {
				t.Fatalf("got %d lines %v, want %d lines %v", len(got), got, len(tt.expected), tt.expected)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("line %d: got %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}
