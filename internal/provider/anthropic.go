package provider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type AnthropicProvider struct {
	model          string
	numSuggestions int
}

func NewAnthropicProvider(model string, numSuggestions int) *AnthropicProvider {
	if model == "" {
		model = "claude-haiku-4-5"
	}
	if numSuggestions <= 0 {
		numSuggestions = 10
	}
	return &AnthropicProvider{
		model:          model,
		numSuggestions: numSuggestions,
	}
}

func (a *AnthropicProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	msgs, err := a.GenerateCommitMessages(ctx, diff)
	if err != nil {
		return "", err
	}
	if len(msgs) == 0 {
		return "", fmt.Errorf("no commit messages generated")
	}
	return msgs[0], nil
}

func (a *AnthropicProvider) GenerateCommitMessages(ctx context.Context, diff string) ([]string, error) {
	if strings.TrimSpace(diff) == "" {
		return nil, fmt.Errorf("no diff provided")
	}

	// Check if claude CLI is available
	if _, err := exec.LookPath("claude"); err != nil {
		return nil, fmt.Errorf("claude CLI not found in PATH. Please install Claude Code CLI: %w", err)
	}

	// Build the prompt
	systemMsg := GetSystemMessage()
	userPrompt := GetCommitMessagePrompt(diff)

	// Modify the prompt to request specific number of suggestions
	fullPrompt := fmt.Sprintf("%s\n\nUser request: %s\n\nIMPORTANT: Generate exactly %d commit messages, one per line. Do not include any other text, explanations, or formatting - just the commit messages.",
		systemMsg, userPrompt, a.numSuggestions)

	// Execute claude CLI with haiku model
	// Using -p flag for print mode and --model for model selection
	cmd := exec.CommandContext(ctx, "claude", "--model", a.model, "-p", fullPrompt)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error executing claude CLI: %w\nOutput: %s", err, string(output))
	}

	// Parse the output - split by newlines and clean
	content := string(output)
	lines := strings.Split(content, "\n")

	var commitMessages []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and lines that look like explanatory text
		if trimmed == "" {
			continue
		}
		// Skip lines that are clearly not commit messages (too long, contain certain patterns)
		if len(trimmed) > 200 {
			continue
		}
		// Skip markdown formatting or numbered lists
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
			// Try to extract the actual commit message
			parts := strings.SplitN(trimmed, " ", 2)
			if len(parts) == 2 {
				trimmed = strings.TrimSpace(parts[1])
			}
		}
		// Remove numbered list formatting like "1. " or "1) "
		if len(trimmed) > 3 {
			if (trimmed[0] >= '0' && trimmed[0] <= '9') && (trimmed[1] == '.' || trimmed[1] == ')') {
				trimmed = strings.TrimSpace(trimmed[2:])
			}
		}

		if trimmed != "" {
			commitMessages = append(commitMessages, trimmed)
		}

		// Stop once we have enough messages
		if len(commitMessages) >= a.numSuggestions {
			break
		}
	}

	if len(commitMessages) == 0 {
		return nil, fmt.Errorf("no valid commit messages generated from Claude output")
	}

	return commitMessages, nil
}
