package domain

import (
	"errors"
	"fmt"
	"strings"
)

// Built-in defaults used when no configuration overrides them. Templates are
// count-agnostic: the builder appends the suggestion count and language.
const (
	DefaultSystemMessage = "You are a helpful assistant that generates git commit messages and pull request titles."

	DefaultCommitTemplate = "Based on the following git diff, generate conventional commit messages. " +
		"Each message must be on its own line, without any numbering, bullet points, or markdown formatting:\n\n%s"

	DefaultPRTitleTemplate = "Based on the following git diff, generate pull request title suggestions. " +
		"Each title must be on its own line, without any numbering, bullet points, or markdown formatting:\n\n%s"

	DefaultSuggestionCount = 10
)

// PromptTemplate is a user-facing text template with a single %s placeholder
// for the diff.
type PromptTemplate struct {
	text string
}

func NewPromptTemplate(text string) (PromptTemplate, error) {
	if strings.TrimSpace(text) == "" {
		return PromptTemplate{}, errors.New("prompt template is empty")
	}
	if strings.Count(text, "%s") != 1 {
		return PromptTemplate{}, errors.New("prompt template must contain exactly one %s placeholder for the diff")
	}
	return PromptTemplate{text: text}, nil
}

func (t PromptTemplate) String() string {
	return t.text
}

// Prompt is the fully assembled input for a Generator.
type Prompt struct {
	System string
	User   string
}

// PromptBuilder assembles a Prompt from its parts, falling back to defaults
// for anything not set.
type PromptBuilder struct {
	system   string
	template PromptTemplate
	language Language
	count    int
}

func NewPromptBuilder() *PromptBuilder {
	tmpl, _ := NewPromptTemplate(DefaultCommitTemplate)
	return &PromptBuilder{
		system:   DefaultSystemMessage,
		template: tmpl,
		language: DefaultLanguage(),
		count:    DefaultSuggestionCount,
	}
}

func (b *PromptBuilder) WithSystemMessage(msg string) *PromptBuilder {
	if strings.TrimSpace(msg) != "" {
		b.system = msg
	}
	return b
}

func (b *PromptBuilder) WithTemplate(t PromptTemplate) *PromptBuilder {
	b.template = t
	return b
}

func (b *PromptBuilder) WithLanguage(l Language) *PromptBuilder {
	b.language = l
	return b
}

func (b *PromptBuilder) WithSuggestionCount(n int) *PromptBuilder {
	if n > 0 {
		b.count = n
	}
	return b
}

func (b *PromptBuilder) Build(diff Diff) Prompt {
	var user strings.Builder
	fmt.Fprintf(&user, b.template.String(), diff.String())
	fmt.Fprintf(&user, "\n\nGenerate exactly %d suggestions.", b.count)
	fmt.Fprintf(&user, " Write every suggestion in %s.", b.language)
	return Prompt{System: b.system, User: user.String()}
}

func (b *PromptBuilder) SuggestionCount() int {
	return b.count
}
