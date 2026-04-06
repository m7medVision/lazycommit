package provider

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/openai/openai-go"
)

// parseOutputLines parses raw LLM output into clean lines, stripping markdown
// formatting, numbered/bulleted list prefixes, and skipping empty or overly long lines.
// It returns at most maxLines results.
func parseOutputLines(raw string, maxLines int) []string {
	lines := strings.Split(raw, "\n")

	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || len(trimmed) > 200 {
			continue
		}
		// Strip markdown heading, bullet, or asterisk prefix
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
			parts := strings.SplitN(trimmed, " ", 2)
			if len(parts) == 2 {
				trimmed = strings.TrimSpace(parts[1])
			}
		}
		// Strip numbered list prefix like "1. ", "10) ", "3. "
		if len(trimmed) > 0 && trimmed[0] >= '0' && trimmed[0] <= '9' {
			i := 0
			for i < len(trimmed) && unicode.IsDigit(rune(trimmed[i])) {
				i++
			}
			if i < len(trimmed) && (trimmed[i] == '.' || trimmed[i] == ')') {
				trimmed = strings.TrimSpace(trimmed[i+1:])
			}
		}
		if trimmed != "" {
			result = append(result, trimmed)
		}
		if len(result) >= maxLines {
			break
		}
	}
	return result
}

// commonProvider holds the common fields and methods for OpenAI-compatible providers.
type commonProvider struct {
	client *openai.Client
	model  string
}

// generateCommitMessages is a helper function to generate commit messages using the OpenAI API.
func (c *commonProvider) generateCommitMessages(ctx context.Context, diff string) ([]string, error) {
	if diff == "" {
		return nil, fmt.Errorf("no diff provided")
	}

	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(c.model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			{OfSystem: &openai.ChatCompletionSystemMessageParam{Content: openai.ChatCompletionSystemMessageParamContentUnion{OfString: openai.String(GetSystemMessage())}}},
			{OfUser: &openai.ChatCompletionUserMessageParam{Content: openai.ChatCompletionUserMessageParamContentUnion{OfString: openai.String(GetCommitMessagePrompt(diff))}}},
		},
	}

	resp, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error making request to OpenAI compatible API: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no commit messages generated")
	}

	content := resp.Choices[0].Message.Content
	messages := strings.Split(content, "\n")
	var cleanMessages []string
	for _, msg := range messages {
		if strings.TrimSpace(msg) != "" {
			cleanMessages = append(cleanMessages, strings.TrimSpace(msg))
		}
	}
	return cleanMessages, nil
}

// generatePRTitles is a helper function to generate pull request titles using the OpenAI API.
func (c *commonProvider) generatePRTitles(ctx context.Context, diff string) ([]string, error) {
	if diff == "" {
		return nil, fmt.Errorf("no diff provided")
	}

	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(c.model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			{OfSystem: &openai.ChatCompletionSystemMessageParam{Content: openai.ChatCompletionSystemMessageParamContentUnion{OfString: openai.String(GetSystemMessage())}}},
			{OfUser: &openai.ChatCompletionUserMessageParam{Content: openai.ChatCompletionUserMessageParamContentUnion{OfString: openai.String(GetPRTitlePrompt(diff))}}},
		},
	}

	resp, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error making request to OpenAI compatible API: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no pr titles generated")
	}

	content := resp.Choices[0].Message.Content
	messages := strings.Split(content, "\n")
	var cleanMessages []string
	for _, msg := range messages {
		if strings.TrimSpace(msg) != "" {
			cleanMessages = append(cleanMessages, strings.TrimSpace(msg))
		}
	}
	return cleanMessages, nil
}
