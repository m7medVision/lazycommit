package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
)

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
