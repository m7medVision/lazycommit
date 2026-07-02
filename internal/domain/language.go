package domain

import "strings"

// Language is the natural language suggestions are written in.
type Language struct {
	name string
}

func DefaultLanguage() Language {
	return Language{name: "English"}
}

// NewLanguage returns the default language for blank input so that a missing
// config field never produces an invalid value object.
func NewLanguage(name string) Language {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return DefaultLanguage()
	}
	return Language{name: trimmed}
}

func (l Language) String() string {
	return l.name
}
