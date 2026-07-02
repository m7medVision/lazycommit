// Package openaicompat adapts any OpenAI-compatible chat-completions
// endpoint (OpenAI, Ollama, OpenRouter, LM Studio, proxies) to the
// app.Generator port. It is the only direct-API backend.
package openaicompat

import (
	"context"
	"errors"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/m7medvision/lazycommit/internal/domain"
)

type Config struct {
	// BaseURL is optional; the official OpenAI endpoint is used when empty.
	BaseURL string
	// APIKey is optional to support local endpoints that ignore it.
	APIKey string
	Model  string
}

type Client struct {
	api   openai.Client
	model domain.ModelID
}

func New(cfg Config) (*Client, error) {
	model, err := domain.NewModelID(cfg.Model)
	if err != nil {
		return nil, fmt.Errorf("openai-compatible backend: %w", err)
	}

	opts := []option.RequestOption{option.WithAPIKey(cfg.APIKey)}
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}
	return &Client{api: openai.NewClient(opts...), model: model}, nil
}

func (c *Client) Generate(ctx context.Context, prompt domain.Prompt) (string, error) {
	resp, err := c.api.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModel(c.model.String()),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(prompt.System),
			openai.UserMessage(prompt.User),
		},
	})
	if err != nil {
		return "", fmt.Errorf("chat completion request: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("chat completion returned no choices")
	}
	return resp.Choices[0].Message.Content, nil
}
