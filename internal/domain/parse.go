package domain

import (
	"strings"
	"unicode"
)

// ParseSuggestions cleans raw LLM output into at most max suggestions. It
// strips markdown headings, bullets, numbered-list prefixes, and code fences,
// and drops blank or over-long lines, so every backend yields identical
// output shape.
func ParseSuggestions(raw string, max int) []Suggestion {
	if max <= 0 {
		return nil
	}

	var result []Suggestion
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || len(trimmed) > MaxSuggestionLength {
			continue
		}
		if isCodeFence(trimmed) {
			continue
		}
		trimmed = stripListPrefix(trimmed)
		if s, err := NewSuggestion(trimmed); err == nil {
			result = append(result, s)
		}
		if len(result) >= max {
			break
		}
	}
	return result
}

func isCodeFence(line string) bool {
	return strings.TrimLeft(line, "`") == ""
}

// stripListPrefix removes a leading markdown heading, bullet, or numbered
// list marker ("# ", "- ", "* ", "1. ", "10) ").
func stripListPrefix(line string) string {
	if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
		return line
	}

	i := 0
	for i < len(line) && unicode.IsDigit(rune(line[i])) {
		i++
	}
	if i > 0 && i < len(line) && (line[i] == '.' || line[i] == ')') {
		return strings.TrimSpace(line[i+1:])
	}
	return line
}
