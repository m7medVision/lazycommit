package provider

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenRouterProvider struct {
	commonProvider
}

func NewOpenRouterProvider(apiKey, model string) *OpenRouterProvider {
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	client := openai.NewClient(
		option.WithBaseURL("https://openrouter.ai/api/v1"),
		option.WithAPIKey(apiKey),
		option.WithHeaderAdd("HTTP-Referer", "https://github.com/m7medvision/lazycommit"),
		option.WithHeaderAdd("User-Agent", "LazyCommit/1.0"),
		option.WithHeaderAdd("X-Title", "LazyCommit"),
	)
	return &OpenRouterProvider{
		commonProvider: commonProvider{
			client: &client,
			model:  model,
		},
	}
}

func (o *OpenRouterProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	messages, err := o.generateCommitMessages(ctx, diff)
	if err != nil {
		return "", err
	}
	if len(messages) == 0 {
		return "", fmt.Errorf("no commit messages generated")
	}
	return messages[0], nil
}

func (o *OpenRouterProvider) GenerateCommitMessages(ctx context.Context, diff string) ([]string, error) {
	return o.generateCommitMessages(ctx, diff)
}
