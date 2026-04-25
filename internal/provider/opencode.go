package provider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

const defaultOpencodeModel = "opencode/minimax-m2.5-free"

var defaultOpencodeFallbackModels = []string{
	"opencode/minimax-m2.5-free",
	"opencode/ling-2.6-flash-free",
	"opencode/hy3-preview-free",
	"opencode/nemotron-3-super-free",
}

type OpencodeProvider struct {
	model          string
	fallbackModels []string
	numSuggestions int
}

func NewOpencodeProvider(model string, fallbackModels []string, numSuggestions int) *OpencodeProvider {
	if strings.TrimSpace(model) == "" {
		model = defaultOpencodeModel
	}
	if len(fallbackModels) == 0 {
		fallbackModels = defaultOpencodeFallbackModels
	}
	if numSuggestions <= 0 {
		numSuggestions = 10
	}
	return &OpencodeProvider{
		model:          strings.TrimSpace(model),
		fallbackModels: fallbackModels,
		numSuggestions: numSuggestions,
	}
}

func opencodeModelCandidates(model string, fallbackModels []string) []string {
	seen := make(map[string]struct{})
	var candidates []string

	add := func(m string) {
		m = strings.TrimSpace(m)
		if m == "" {
			return
		}
		if _, ok := seen[m]; ok {
			return
		}
		seen[m] = struct{}{}
		candidates = append(candidates, m)
	}

	add(model)
	for _, fallback := range fallbackModels {
		add(fallback)
	}
	return candidates
}

func (o *OpencodeProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	msgs, err := o.GenerateCommitMessages(ctx, diff)
	if err != nil {
		return "", err
	}
	if len(msgs) == 0 {
		return "", fmt.Errorf("no commit messages generated")
	}
	return msgs[0], nil
}

func (o *OpencodeProvider) GenerateCommitMessages(ctx context.Context, diff string) ([]string, error) {
	if strings.TrimSpace(diff) == "" {
		return nil, fmt.Errorf("no diff provided")
	}

	systemMsg := GetSystemMessage()
	userPrompt := GetCommitMessagePrompt(diff)
	fullPrompt := fmt.Sprintf("%s\n\nUser request: %s\n\nIMPORTANT: Generate exactly %d commit messages, one per line. Do not include any other text, explanations, or formatting - just the commit messages.",
		systemMsg, userPrompt, o.numSuggestions)

	output, err := o.runCLI(ctx, fullPrompt)
	if err != nil {
		return nil, err
	}

	commitMessages := parseOutputLines(output, o.numSuggestions)
	if len(commitMessages) == 0 {
		return nil, fmt.Errorf("no valid commit messages generated from opencode output")
	}

	return commitMessages, nil
}

func (o *OpencodeProvider) GeneratePRTitle(ctx context.Context, diff string) (string, error) {
	titles, err := o.GeneratePRTitles(ctx, diff)
	if err != nil {
		return "", err
	}
	if len(titles) == 0 {
		return "", fmt.Errorf("no PR titles generated")
	}
	return titles[0], nil
}

func (o *OpencodeProvider) GeneratePRTitles(ctx context.Context, diff string) ([]string, error) {
	if strings.TrimSpace(diff) == "" {
		return nil, fmt.Errorf("no diff provided")
	}

	systemMsg := GetSystemMessage()
	userPrompt := GetPRTitlePrompt(diff)
	fullPrompt := fmt.Sprintf("%s\n\nUser request: %s\n\nIMPORTANT: Generate exactly %d pull request titles, one per line. Do not include any other text, explanations, or formatting - just the PR titles.",
		systemMsg, userPrompt, o.numSuggestions)

	output, err := o.runCLI(ctx, fullPrompt)
	if err != nil {
		return nil, err
	}

	prTitles := parseOutputLines(output, o.numSuggestions)
	if len(prTitles) == 0 {
		return nil, fmt.Errorf("no valid PR titles generated from opencode output")
	}

	return prTitles, nil
}

func (o *OpencodeProvider) runCLI(ctx context.Context, prompt string) (string, error) {
	if _, err := exec.LookPath("opencode"); err != nil {
		return "", fmt.Errorf("opencode CLI not found in PATH. Please install opencode CLI: %w", err)
	}

	candidates := opencodeModelCandidates(o.model, o.fallbackModels)
	var errors []string
	for _, model := range candidates {
		cmd := exec.CommandContext(ctx, "opencode", "run", "--model", model, "--", prompt)

		var stdoutBuf, stderrBuf strings.Builder
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf

		if err := cmd.Run(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v: %s", model, err, strings.TrimSpace(stderrBuf.String())))
			continue
		}

		output := strings.TrimSpace(stdoutBuf.String())
		if output != "" {
			return output, nil
		}
		errors = append(errors, fmt.Sprintf("%s: empty output", model))
	}

	return "", fmt.Errorf("error executing opencode CLI with models %q: %s", candidates, strings.Join(errors, "; "))
}
