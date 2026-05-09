package provider

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/m7medvision/lazycommit/internal/config"
)

var languageInstructionPattern = regexp.MustCompile(`\n\nIMPORTANT: Generate all content in .*\.`)

// GetCommitMessagePrompt returns the standardized prompt for generating commit messages
func GetCommitMessagePrompt(diff string) string {
	basePrompt := config.GetCommitMessagePromptFromConfig(diff)
	opts := GetRuntimeCommitPromptOptions()
	if strings.TrimSpace(opts.Language) != "" {
		basePrompt = languageInstructionPattern.ReplaceAllString(basePrompt, "")
	}

	var extraRules []string
	if opts.Generate > 0 {
		extraRules = append(extraRules, fmt.Sprintf("IMPORTANT: Generate exactly %d commit messages. Put each message on its own line.", opts.Generate))
	}
	if strings.TrimSpace(opts.Language) != "" {
		extraRules = append(extraRules, fmt.Sprintf("IMPORTANT: Generate all content in %s.", strings.TrimSpace(opts.Language)))
	}
	if opts.Emoji {
		extraRules = append(extraRules, `Gitmoji rules:
- Prefix exactly one emoji before commit type.
- Mapping: feat=✨ fix=🐛 refactor=♻️ perf=⚡️ docs=📝 style=🎨 test=🧪 chore=🔧 ci=👷 build=📦 revert=⏪️ security=🔒️.
- Format: <emoji> <type>(<scope>): <description>
- Every output line must start with an emoji and a space.`)
	}

	if len(extraRules) == 0 {
		return basePrompt
	}

	return fmt.Sprintf("%s\n\n%s", basePrompt, strings.Join(extraRules, "\n"))
}

// GetPRTitlePrompt returns the standardized prompt for generating pull request titles
func GetPRTitlePrompt(diff string) string {
	return config.GetPRTitlePromptFromConfig(diff)
}

// GetSystemMessage returns the standardized system message for commit message generation
func GetSystemMessage() string {
	return config.GetSystemMessageFromConfig()
}
