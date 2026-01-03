package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/m7medvision/lazycommit/internal/config"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type CopilotProvider struct {
	apiKey     string
	model      string
	endpoint   string
	httpClient *http.Client

	clientMu     sync.RWMutex
	openaiClient *openai.Client
	clientToken  string
}

func newOptimizedHTTPClient() *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		DisableKeepAlives:   false,
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}

func NewCopilotProvider(token, endpoint string) *CopilotProvider {
	if endpoint == "" {
		endpoint = "https://api.githubcopilot.com"
	}
	return &CopilotProvider{
		apiKey:     token,
		model:      "gpt-4o",
		endpoint:   endpoint,
		httpClient: newOptimizedHTTPClient(),
	}
}

func NewCopilotProviderWithModel(token, model, endpoint string) *CopilotProvider {
	m := normalizeCopilotModel(model)
	if endpoint == "" {
		endpoint = "https://api.githubcopilot.com"
	}
	return &CopilotProvider{
		apiKey:     token,
		model:      m,
		endpoint:   endpoint,
		httpClient: newOptimizedHTTPClient(),
	}
}

func normalizeCopilotModel(model string) string {
	m := strings.TrimSpace(model)
	if m == "" {
		return "gpt-4o"
	}
	if strings.Contains(m, "/") {
		parts := strings.SplitN(m, "/", 2)
		if len(parts) == 2 && parts[1] != "" {
			return parts[1]
		}
	}
	return m
}

func (c *CopilotProvider) exchangeGitHubToken(ctx context.Context, githubToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/copilot_internal/v2/token", nil)
	if err != nil {
		return "", fmt.Errorf("failed creating token request: %w", err)
	}
	req.Header.Set("Authorization", "Token "+githubToken)
	req.Header.Set("User-Agent", "lazycommit/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed exchanging token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body struct {
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		return "", fmt.Errorf("token exchange failed: %d %s", resp.StatusCode, body.Message)
	}

	var tr struct {
		Token     string `json:"token"`
		ExpiresAt int64  `json:"expires_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("failed decoding token response: %w", err)
	}
	if tr.Token == "" {
		return "", fmt.Errorf("empty copilot bearer token")
	}

	if err := config.SaveCopilotTokenCache(tr.Token, tr.ExpiresAt, githubToken); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache token: %v\n", err)
	}

	return tr.Token, nil
}

func (c *CopilotProvider) getGitHubToken() string {
	if c.apiKey != "" {
		return c.apiKey
	}
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	return ""
}

func (c *CopilotProvider) getBearerToken(ctx context.Context) (string, error) {
	githubToken := c.getGitHubToken()
	if githubToken == "" {
		return "", fmt.Errorf("GitHub token is required for Copilot provider")
	}

	if cached := config.GetCachedCopilotToken(githubToken); cached != nil {
		return cached.Token, nil
	}

	return c.exchangeGitHubToken(ctx, githubToken)
}

func (c *CopilotProvider) getOrCreateClient(ctx context.Context) (*openai.Client, error) {
	bearer, err := c.getBearerToken(ctx)
	if err != nil {
		return nil, err
	}

	c.clientMu.Lock()
	defer c.clientMu.Unlock()

	if c.openaiClient == nil || c.clientToken != bearer {
		client := openai.NewClient(
			option.WithBaseURL(c.endpoint),
			option.WithAPIKey(bearer),
			option.WithHeader("Editor-Version", "lazycommit/1.0"),
			option.WithHeader("Editor-Plugin-Version", "lazycommit/1.0"),
			option.WithHeader("Copilot-Integration-Id", "vscode-chat"),
		)
		c.openaiClient = &client
		c.clientToken = bearer
	}

	return c.openaiClient, nil
}

func (c *CopilotProvider) invalidateAndRetry(ctx context.Context) (*openai.Client, error) {
	config.InvalidateCopilotTokenCache()

	c.clientMu.Lock()
	c.openaiClient = nil
	c.clientToken = ""
	c.clientMu.Unlock()

	return c.getOrCreateClient(ctx)
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "Unauthorized")
}

func parseResponseLines(content string) []string {
	parts := strings.Split(content, "\n")
	var out []string
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func (c *CopilotProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	msgs, err := c.GenerateCommitMessages(ctx, diff)
	if err != nil {
		return "", err
	}
	if len(msgs) == 0 {
		return "", fmt.Errorf("no commit messages generated")
	}
	return msgs[0], nil
}

func (c *CopilotProvider) GenerateCommitMessages(ctx context.Context, diff string) ([]string, error) {
	if strings.TrimSpace(diff) == "" {
		return nil, fmt.Errorf("no diff provided")
	}

	client, err := c.getOrCreateClient(ctx)
	if err != nil {
		return nil, err
	}

	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(c.model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			{OfSystem: &openai.ChatCompletionSystemMessageParam{Content: openai.ChatCompletionSystemMessageParamContentUnion{OfString: openai.String(GetSystemMessage())}}},
			{OfUser: &openai.ChatCompletionUserMessageParam{Content: openai.ChatCompletionUserMessageParamContentUnion{OfString: openai.String(GetCommitMessagePrompt(diff))}}},
		},
	}

	resp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		if isAuthError(err) {
			client, err = c.invalidateAndRetry(ctx)
			if err != nil {
				return nil, err
			}
			resp, err = client.Chat.Completions.New(ctx, params)
			if err != nil {
				return nil, fmt.Errorf("error making request to Copilot (after retry): %w", err)
			}
		} else {
			return nil, fmt.Errorf("error making request to Copilot: %w", err)
		}
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no commit messages generated")
	}

	content := resp.Choices[0].Message.Content
	result := parseResponseLines(content)
	if len(result) == 0 {
		return nil, fmt.Errorf("no valid commit messages generated")
	}
	return result, nil
}

func (c *CopilotProvider) GeneratePRTitle(ctx context.Context, diff string) (string, error) {
	titles, err := c.GeneratePRTitles(ctx, diff)
	if err != nil {
		return "", err
	}
	if len(titles) == 0 {
		return "", fmt.Errorf("no PR titles generated")
	}
	return titles[0], nil
}

func (c *CopilotProvider) GeneratePRTitles(ctx context.Context, diff string) ([]string, error) {
	if strings.TrimSpace(diff) == "" {
		return nil, fmt.Errorf("no diff provided")
	}

	client, err := c.getOrCreateClient(ctx)
	if err != nil {
		return nil, err
	}

	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(c.model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			{OfSystem: &openai.ChatCompletionSystemMessageParam{Content: openai.ChatCompletionSystemMessageParamContentUnion{OfString: openai.String(GetSystemMessage())}}},
			{OfUser: &openai.ChatCompletionUserMessageParam{Content: openai.ChatCompletionUserMessageParamContentUnion{OfString: openai.String(GetPRTitlePrompt(diff))}}},
		},
	}

	resp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		if isAuthError(err) {
			client, err = c.invalidateAndRetry(ctx)
			if err != nil {
				return nil, err
			}
			resp, err = client.Chat.Completions.New(ctx, params)
			if err != nil {
				return nil, fmt.Errorf("error making request to Copilot (after retry): %w", err)
			}
		} else {
			return nil, fmt.Errorf("error making request to Copilot: %w", err)
		}
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no PR titles generated")
	}

	content := resp.Choices[0].Message.Content
	result := parseResponseLines(content)
	if len(result) == 0 {
		return nil, fmt.Errorf("no valid PR titles generated")
	}
	return result, nil
}
