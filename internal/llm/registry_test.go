package llm

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/m7medvision/lazycommit/v2/internal/app"
	"github.com/m7medvision/lazycommit/v2/internal/domain"
)

type nullGenerator struct{ name string }

func (n nullGenerator) Generate(context.Context, domain.Prompt) (string, error) {
	return n.name, nil
}

func TestRegistrySelectsByName(t *testing.T) {
	r := NewRegistry()
	r.Register("beta", func(BackendConfig) (app.Generator, error) { return nullGenerator{"beta"}, nil })
	r.Register("alpha", func(BackendConfig) (app.Generator, error) { return nullGenerator{"alpha"}, nil })

	gen, err := r.New("beta", BackendConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, _ := gen.Generate(context.Background(), domain.Prompt{})
	if out != "beta" {
		t.Fatalf("wrong backend constructed: %q", out)
	}
}

func TestRegistryPassesConfigToFactory(t *testing.T) {
	r := NewRegistry()
	var got BackendConfig
	r.Register("x", func(cfg BackendConfig) (app.Generator, error) {
		got = cfg
		return nullGenerator{}, nil
	})

	cfg := BackendConfig{Model: "m1", APIKey: "k", BaseURL: "http://localhost"}
	if _, err := r.New("x", cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != cfg {
		t.Fatalf("factory got %+v, want %+v", got, cfg)
	}
}

func TestRegistryUnknownNameListsAvailable(t *testing.T) {
	r := NewRegistry()
	r.Register("opencode", func(BackendConfig) (app.Generator, error) { return nullGenerator{}, nil })
	r.Register("claude-code", func(BackendConfig) (app.Generator, error) { return nullGenerator{}, nil })

	_, err := r.New("copilot", BackendConfig{})
	if err == nil {
		t.Fatal("expected error for unknown backend")
	}
	if !strings.Contains(err.Error(), "claude-code, opencode") {
		t.Fatalf("error should list available backends sorted: %v", err)
	}
}

func TestRegistryFactoryErrorPropagates(t *testing.T) {
	r := NewRegistry()
	boom := errors.New("bad config")
	r.Register("x", func(BackendConfig) (app.Generator, error) { return nil, boom })

	if _, err := r.New("x", BackendConfig{}); !errors.Is(err, boom) {
		t.Fatalf("expected factory error, got %v", err)
	}
}
