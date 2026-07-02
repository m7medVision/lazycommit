package domain

import (
	"errors"
	"strings"
)

// MaxSuggestionLength bounds a single suggestion line; anything longer is
// prose or a stray diff fragment, not a usable commit message or PR title.
const MaxSuggestionLength = 200

// Suggestion is a single non-empty, single-line generated message.
type Suggestion struct {
	text string
}

func NewSuggestion(text string) (Suggestion, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return Suggestion{}, errors.New("suggestion is empty")
	}
	if strings.ContainsAny(trimmed, "\n\r") {
		return Suggestion{}, errors.New("suggestion spans multiple lines")
	}
	if len(trimmed) > MaxSuggestionLength {
		return Suggestion{}, errors.New("suggestion exceeds maximum length")
	}
	return Suggestion{text: trimmed}, nil
}

func (s Suggestion) String() string {
	return s.text
}
