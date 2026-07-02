// main is the composition root: the only place concrete adapters are
// constructed and wired into the use cases.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/m7medvision/lazycommit/cmd"
	"github.com/m7medvision/lazycommit/internal/app"
	"github.com/m7medvision/lazycommit/internal/config"
	"github.com/m7medvision/lazycommit/internal/domain"
	"github.com/m7medvision/lazycommit/internal/git"
	"github.com/m7medvision/lazycommit/internal/llm"
	"github.com/m7medvision/lazycommit/internal/llm/middleware"
	"github.com/m7medvision/lazycommit/internal/llm/openaicompat"
)

// version is injected by goreleaser via ldflags.
var version = "dev"

const (
	generationTimeout = 2 * time.Minute
	retryAttempts     = 2
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, os.Stdin))
}

func run(args []string, stdout, stderr io.Writer, stdin io.Reader) int {
	deps, err := buildDeps()
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "Error:", err)
		return 1
	}

	root := cmd.NewRoot(deps)
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetIn(stdin)

	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(stderr, "Error:", err)
		return 1
	}
	return 0
}

func buildDeps() (cmd.Deps, error) {
	gitCLI := git.New()

	globalDir, err := config.DefaultGlobalDir()
	if err != nil {
		return cmd.Deps{}, err
	}
	cfgRepo := config.NewRepository(globalDir, gitCLI.RepoRoot(context.Background()))

	registry := newRegistry()

	// Generator construction is deferred to the first Generate call so that
	// empty-diff runs succeed (and short-circuit) even with broken or
	// missing backend configuration.
	gen := lazyGenerator{build: func() (app.Generator, error) {
		return buildGenerator(registry, cfgRepo)
	}}

	return cmd.Deps{
		NewCommitUC: func() (*app.GenerateCommitSuggestions, error) {
			return app.NewGenerateCommitSuggestions(gen, gitCLI, cfgRepo), nil
		},
		NewPRUC: func() (*app.GeneratePRTitles, error) {
			return app.NewGeneratePRTitles(gen, gitCLI, cfgRepo), nil
		},
		ConfigRepo:   cfgRepo,
		BackendNames: registry.Names(),
		Version:      version,
		V1Hint:       v1Hint(cfgRepo),
	}, nil
}

// newRegistry is the single table where backends are wired; adding one is a
// new factory line, nothing else changes.
func newRegistry() *llm.Registry {
	r := llm.NewRegistry()
	r.Register("openai-compatible", func(cfg llm.BackendConfig) (app.Generator, error) {
		return openaicompat.New(openaicompat.Config{
			BaseURL: cfg.BaseURL,
			APIKey:  cfg.APIKey,
			Model:   cfg.Model,
		})
	})
	return r
}

// buildGenerator assembles the active backend: one generator per configured
// model, each bounded by timeout and retried, then chained for fallback.
func buildGenerator(registry *llm.Registry, cfgRepo *config.Repository) (app.Generator, error) {
	backends, err := cfgRepo.LoadBackends()
	if err != nil {
		return nil, err
	}
	settings := backends.Backends[backends.Active]

	models := dedupe(append([]string{settings.Model}, settings.FallbackModels...))
	if len(models) == 0 {
		return nil, fmt.Errorf("backend %q has no model configured; run 'lazycommit config set'", backends.Active)
	}

	gens := make([]app.Generator, 0, len(models))
	for _, model := range models {
		gen, err := registry.New(backends.Active, llm.BackendConfig{
			Model:   model,
			APIKey:  settings.APIKey,
			BaseURL: settings.BaseURL,
		})
		if err != nil {
			return nil, err
		}
		gens = append(gens, middleware.WithRetry(middleware.WithTimeout(gen, generationTimeout), retryAttempts))
	}
	return middleware.NewFallbackChain(gens...)
}

// lazyGenerator defers backend construction until generation is actually
// needed.
type lazyGenerator struct {
	build func() (app.Generator, error)
}

func (l lazyGenerator) Generate(ctx context.Context, prompt domain.Prompt) (string, error) {
	gen, err := l.build()
	if err != nil {
		return "", err
	}
	return gen.Generate(ctx, prompt)
}

func v1Hint(cfgRepo *config.Repository) string {
	if cfgRepo.HasBackendsFile() {
		return ""
	}
	if p := cfgRepo.V1ConfigPath(); p != "" {
		return fmt.Sprintf("Note: found v1 config at %s; lazycommit v2 uses a new format — run 'lazycommit config set'.", p)
	}
	return ""
}

func dedupe(models []string) []string {
	seen := make(map[string]struct{}, len(models))
	var out []string
	for _, m := range models {
		if m == "" {
			continue
		}
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		out = append(out, m)
	}
	return out
}
