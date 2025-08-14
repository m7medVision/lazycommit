package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)


type OpenAIProvider struct {
	client *openai.Client
	model  string
}


func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	return &OpenAIProvider{
		client: &client,
		model:  model,
	}
}



func (o *OpenAIProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	messages, err := o.GenerateCommitMessages(ctx, diff)
	if err != nil {
		return "", err
	}
	if len(messages) == 0 {
		return "", fmt.Errorf("no commit messages generated")
	}
	return messages[0], nil
}


func (o *OpenAIProvider) GenerateCommitMessages(ctx context.Context, diff string) ([]string, error) {
	if diff == "" {
		return nil, fmt.Errorf("no diff provided")
	}


	prompt := fmt.Sprintf("Based on the following git diff, generate 10 conventional commit messages. Each message should be on a new line, without any numbering or bullet points:\n\n%s", diff)


	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(o.model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			{OfSystem: &openai.ChatCompletionSystemMessageParam{Content: openai.ChatCompletionSystemMessageParamContentUnion{OfString: openai.String("You are a helpful assistant that generates git commit messages.")}}},
			{OfUser: &openai.ChatCompletionUserMessageParam{Content: openai.ChatCompletionUserMessageParamContentUnion{OfString: openai.String(prompt)}}},
		},
	}


	resp, err := o.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error making request to OpenAI: %w", err)
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
