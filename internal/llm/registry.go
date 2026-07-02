// Package llm selects and constructs Generator backends. The registry keeps
// LLM calling open/closed: a new backend is a new implementation plus one
// Register call in the composition root — no edits to existing code.
package llm

import (
	"fmt"
	"sort"
	"strings"

	"github.com/m7medvision/lazycommit/v2/internal/app"
)

// BackendConfig carries everything a factory may need to build a backend for
// one specific model. Fields irrelevant to a given backend are ignored by it.
type BackendConfig struct {
	Model   string
	APIKey  string
	BaseURL string
}

// Factory builds a Generator from its configuration.
type Factory func(cfg BackendConfig) (app.Generator, error)

// Registry maps backend names to factories.
type Registry struct {
	factories map[string]Factory
}

func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

func (r *Registry) Register(name string, f Factory) {
	r.factories[name] = f
}

// New constructs the named backend, or fails listing the registered names so
// a typo in the config is immediately actionable.
func (r *Registry) New(name string, cfg BackendConfig) (app.Generator, error) {
	f, ok := r.factories[name]
	if !ok {
		return nil, fmt.Errorf("unknown backend %q (available: %s)", name, strings.Join(r.Names(), ", "))
	}
	return f(cfg)
}

// Names returns the registered backend names, sorted.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
