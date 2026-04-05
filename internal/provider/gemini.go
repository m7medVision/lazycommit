package provider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type GeminiProvider struct {
	model          string
	numSuggestions int
}

func NewGeminiProvider(model string, numSuggestions int) *GeminiProvider {
	if model == "" {
		model = "flash"
	}
	if numSuggestions <= 0 {
		numSuggestions = 10
	}
	return &GeminiProvider{
		model:          model,
		numSuggestions: numSuggestions,
	}
}

func (g *GeminiProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	msgs, err := g.GenerateCommitMessages(ctx, diff)
	if err != nil {
		return "", err
	}
	if len(msgs) == 0 {
		return "", fmt.Errorf("no commit messages generated")
	}
	return msgs[0], nil
}

func (g *GeminiProvider) GenerateCommitMessages(ctx context.Context, diff string) ([]string, error) {
	if strings.TrimSpace(diff) == "" {
		return nil, fmt.Errorf("no diff provided")
	}

	// Check if gemini CLI is available
	if _, err := exec.LookPath("gemini"); err != nil {
		return nil, fmt.Errorf("gemini CLI not found in PATH. Please install Gemini CLI: %w", err)
	}

	// Build the prompt
	systemMsg := GetSystemMessage()
	userPrompt := GetCommitMessagePrompt(diff)

	// Modify the prompt to request specific number of suggestions
	fullPrompt := fmt.Sprintf("%s\n\nUser request: %s\n\nIMPORTANT: Generate exactly %d commit messages, one per line. Do not include any other text, explanations, or formatting - just the commit messages.",
		systemMsg, userPrompt, g.numSuggestions)

	output, err := g.runCLI(ctx, fullPrompt)
	if err != nil {
		return nil, err
	}

	commitMessages := parseOutputLines(output, g.numSuggestions)
	if len(commitMessages) == 0 {
		return nil, fmt.Errorf("no valid commit messages generated from Gemini output")
	}

	return commitMessages, nil
}

func (g *GeminiProvider) GeneratePRTitle(ctx context.Context, diff string) (string, error) {
	titles, err := g.GeneratePRTitles(ctx, diff)
	if err != nil {
		return "", err
	}
	if len(titles) == 0 {
		return "", fmt.Errorf("no PR titles generated")
	}
	return titles[0], nil
}

func (g *GeminiProvider) GeneratePRTitles(ctx context.Context, diff string) ([]string, error) {
	if strings.TrimSpace(diff) == "" {
		return nil, fmt.Errorf("no diff provided")
	}

	// Check if gemini CLI is available
	if _, err := exec.LookPath("gemini"); err != nil {
		return nil, fmt.Errorf("gemini CLI not found in PATH. Please install Gemini CLI: %w", err)
	}

	// Build the prompt using PR title template
	systemMsg := GetSystemMessage()
	userPrompt := GetPRTitlePrompt(diff)

	// Modify the prompt to request specific number of suggestions
	fullPrompt := fmt.Sprintf("%s\n\nUser request: %s\n\nIMPORTANT: Generate exactly %d pull request titles, one per line. Do not include any other text, explanations, or formatting - just the PR titles.",
		systemMsg, userPrompt, g.numSuggestions)

	output, err := g.runCLI(ctx, fullPrompt)
	if err != nil {
		return nil, err
	}

	prTitles := parseOutputLines(output, g.numSuggestions)
	if len(prTitles) == 0 {
		return nil, fmt.Errorf("no valid PR titles generated from Gemini output")
	}

	return prTitles, nil
}

// runCLI executes the gemini CLI with the given prompt via stdin and returns stdout.
func (g *GeminiProvider) runCLI(ctx context.Context, prompt string) (string, error) {
	// Piping into gemini triggers Headless mode.
	cmd := exec.CommandContext(ctx, "gemini", "--model", g.model)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("error creating stdin pipe: %w", err)
	}

	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error starting gemini CLI: %w", err)
	}

	_, writeErr := stdin.Write([]byte(prompt))
	stdin.Close()

	waitErr := cmd.Wait()

	if writeErr != nil {
		return "", fmt.Errorf("error writing to gemini CLI stdin: %w", writeErr)
	}

	if waitErr != nil {
		return "", fmt.Errorf("error executing gemini CLI: %w\nStderr: %s", waitErr, stderrBuf.String())
	}

	return stdoutBuf.String(), nil
}
