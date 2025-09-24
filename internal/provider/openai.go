package provider

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAIProvider struct {
	commonProvider
}

func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	return NewOpenAIProviderWithBaseURL(apiKey, model, "")
}

func NewOpenAIProviderWithBaseURL(apiKey, model, baseURL string) *OpenAIProvider {
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	
	if baseURL != "" {
		client := openai.NewClient(
			option.WithAPIKey(apiKey),
			option.WithBaseURL(baseURL),
		)
		return &OpenAIProvider{
			commonProvider: commonProvider{
				client: &client,
				model:  model,
			},
		}
	} else {
		client := openai.NewClient(
			option.WithAPIKey(apiKey),
		)
		return &OpenAIProvider{
			commonProvider: commonProvider{
				client: &client,
				model:  model,
			},
		}
	}
}

func (o *OpenAIProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	messages, err := o.generateCommitMessages(ctx, diff)
	if err != nil {
		return "", err
	}
	if len(messages) == 0 {
		return "", fmt.Errorf("no commit messages generated")
	}
	return messages[0], nil
}

func (o *OpenAIProvider) GenerateCommitMessages(ctx context.Context, diff string) ([]string, error) {
	return o.generateCommitMessages(ctx, diff)
}
