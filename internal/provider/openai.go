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

func NewOpenAIProvider(apiKey, model, endpoint string) *OpenAIProvider {
	if model == "" {
		model = "gpt-5-mini"
	}

	// Set default endpoint if none provided
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}

	client := openai.NewClient(
		option.WithBaseURL(endpoint),
		option.WithAPIKey(apiKey),
	)
	return &OpenAIProvider{
		commonProvider: commonProvider{
			client: &client,
			model:  model,
		},
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

func (o *OpenAIProvider) GeneratePRTitle(ctx context.Context, diff string) (string, error) {
	titles, err := o.generatePRTitles(ctx, diff)
	if err != nil {
		return "", err
	}
	if len(titles) == 0 {
		return "", fmt.Errorf("no PR titles generated")
	}
	return titles[0], nil
}

func (o *OpenAIProvider) GeneratePRTitles(ctx context.Context, diff string) ([]string, error) {
	return o.generatePRTitles(ctx, diff)
}
