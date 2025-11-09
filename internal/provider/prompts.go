package provider

import "github.com/m7medvision/lazycommit/internal/config"

// GetCommitMessagePrompt returns the standardized prompt for generating commit messages
func GetCommitMessagePrompt(diff string) string {
	return config.GetCommitMessagePromptFromConfig(diff)
}

// GetPRTitlePrompt returns the standardized prompt for generating pull request titles
func GetPRTitlePrompt(diff string) string {
	return config.GetPRTitlePromptFromConfig(diff)
}

// GetSystemMessage returns the standardized system message for commit message generation
func GetSystemMessage() string {
	return config.GetSystemMessageFromConfig()
}
