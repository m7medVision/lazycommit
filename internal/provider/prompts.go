package provider

import "fmt"

// GetCommitMessagePrompt returns the standardized prompt for generating commit messages
func GetCommitMessagePrompt(diff string) string {
	return fmt.Sprintf("Based on the following git diff, generate 10 conventional commit messages. Each message should be on a new line, without any numbering or bullet points:\n\n%s", diff)
}

// GetSystemMessage returns the standardized system message for commit message generation
func GetSystemMessage() string {
	return "You are a helpful assistant that generates git commit messages."
}
