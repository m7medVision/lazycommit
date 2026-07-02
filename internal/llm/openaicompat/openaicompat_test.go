package openaicompat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/m7medvision/lazycommit/internal/domain"
)

type chatRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

func TestNewRequiresModel(t *testing.T) {
	if _, err := New(Config{Model: "  "}); err == nil {
		t.Fatal("expected error for blank model")
	}
}

func TestGenerateSpeaksChatCompletions(t *testing.T) {
	var got chatRequest
	var gotAuth, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("bad request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"feat: one\nfix: two"}}]}`))
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL, APIKey: "test-key", Model: "test-model"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, err := client.Generate(context.Background(), domain.Prompt{System: "sys", User: "user"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "feat: one\nfix: two" {
		t.Fatalf("unexpected output: %q", out)
	}
	if gotPath != "/chat/completions" {
		t.Fatalf("unexpected path: %q", gotPath)
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("unexpected auth header: %q", gotAuth)
	}
	if got.Model != "test-model" {
		t.Fatalf("model = %q", got.Model)
	}
	if len(got.Messages) != 2 || got.Messages[0].Role != "system" || got.Messages[0].Content != "sys" ||
		got.Messages[1].Role != "user" || got.Messages[1].Content != "user" {
		t.Fatalf("unexpected messages: %+v", got.Messages)
	}
}

func TestGenerateNoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL, Model: "m"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := client.Generate(context.Background(), domain.Prompt{}); err == nil {
		t.Fatal("expected error for empty choices")
	}
}

func TestGenerateHTTPErrorSurfaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":{"message":"invalid api key"}}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL, Model: "m"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = client.Generate(context.Background(), domain.Prompt{})
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected 401 error to surface, got %v", err)
	}
}
