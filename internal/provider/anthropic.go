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

	output, err := a.runCLI(ctx, fullPrompt)
	if err != nil {
		return nil, err
	}

	commitMessages := parseOutputLines(output, a.numSuggestions)
	if len(commitMessages) == 0 {
		return nil, fmt.Errorf("no valid commit messages generated from Claude output")
	}

	return commitMessages, nil
}

func (a *AnthropicProvider) GeneratePRTitle(ctx context.Context, diff string) (string, error) {
	titles, err := a.GeneratePRTitles(ctx, diff)
	if err != nil {
		return "", err
	}
	if len(titles) == 0 {
		return "", fmt.Errorf("no PR titles generated")
	}
	return titles[0], nil
}

func (a *AnthropicProvider) GeneratePRTitles(ctx context.Context, diff string) ([]string, error) {
	if strings.TrimSpace(diff) == "" {
		return nil, fmt.Errorf("no diff provided")
	}

	// Check if claude CLI is available
	if _, err := exec.LookPath("claude"); err != nil {
		return nil, fmt.Errorf("claude CLI not found in PATH. Please install Claude Code CLI: %w", err)
	}

	// Build the prompt using PR title template
	systemMsg := GetSystemMessage()
	userPrompt := GetPRTitlePrompt(diff)

	// Modify the prompt to request specific number of suggestions
	fullPrompt := fmt.Sprintf("%s\n\nUser request: %s\n\nIMPORTANT: Generate exactly %d pull request titles, one per line. Do not include any other text, explanations, or formatting - just the PR titles.",
		systemMsg, userPrompt, a.numSuggestions)

	output, err := a.runCLI(ctx, fullPrompt)
	if err != nil {
		return nil, err
	}

	prTitles := parseOutputLines(output, a.numSuggestions)
	if len(prTitles) == 0 {
		return nil, fmt.Errorf("no valid PR titles generated from Claude output")
	}

	return prTitles, nil
}

// runCLI executes the claude CLI with the given prompt via stdin and returns stdout.
func (a *AnthropicProvider) runCLI(ctx context.Context, prompt string) (string, error) {
	// Using -p flag for print mode and --model for model selection
	// Pipe prompt via stdin to avoid Windows command line length limits (8191 chars)
	cmd := exec.CommandContext(ctx, "claude", "--model", a.model, "-p", "-")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("error creating stdin pipe: %w", err)
	}

	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error starting claude CLI: %w", err)
	}

	_, writeErr := stdin.Write([]byte(prompt))
	stdin.Close()

	waitErr := cmd.Wait()

	if writeErr != nil {
		return "", fmt.Errorf("error writing to claude CLI stdin: %w", writeErr)
	}

	if waitErr != nil {
		return "", fmt.Errorf("error executing claude CLI: %w\nStderr: %s", waitErr, stderrBuf.String())
	}

	return stdoutBuf.String(), nil
}
