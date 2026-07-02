// Package middleware provides cross-cutting Generator decorators so
// individual backends stay free of retry, timeout, and fallback logic.
package middleware

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/m7medvision/lazycommit/internal/app"
	"github.com/m7medvision/lazycommit/internal/domain"
)

type timeoutGenerator struct {
	next app.Generator
	d    time.Duration
}

// WithTimeout bounds every Generate call; a hung backend returns a timeout
// error instead of blocking forever.
func WithTimeout(next app.Generator, d time.Duration) app.Generator {
	if d <= 0 {
		return next
	}
	return timeoutGenerator{next: next, d: d}
}

func (g timeoutGenerator) Generate(ctx context.Context, prompt domain.Prompt) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, g.d)
	defer cancel()
	out, err := g.next.Generate(ctx, prompt)
	if err != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("generation timed out after %s: %w", g.d, err)
	}
	return out, err
}

type retryGenerator struct {
	next     app.Generator
	attempts int
}

// WithRetry retries failed Generate calls up to attempts total tries. It
// never retries once the context is cancelled or its deadline passed.
func WithRetry(next app.Generator, attempts int) app.Generator {
	if attempts <= 1 {
		return next
	}
	return retryGenerator{next: next, attempts: attempts}
}

func (g retryGenerator) Generate(ctx context.Context, prompt domain.Prompt) (string, error) {
	var lastErr error
	for i := 0; i < g.attempts; i++ {
		out, err := g.next.Generate(ctx, prompt)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			break
		}
	}
	return "", fmt.Errorf("after %d attempts: %w", g.attempts, lastErr)
}

type fallbackChain struct {
	gens []app.Generator
}

// NewFallbackChain tries each generator in order and returns the first
// success. Composition builds one generator per configured model, so model
// fallback and backend fallback are the same mechanism.
func NewFallbackChain(gens ...app.Generator) (app.Generator, error) {
	if len(gens) == 0 {
		return nil, errors.New("fallback chain needs at least one generator")
	}
	if len(gens) == 1 {
		return gens[0], nil
	}
	return fallbackChain{gens: gens}, nil
}

func (c fallbackChain) Generate(ctx context.Context, prompt domain.Prompt) (string, error) {
	var lastErr error
	for _, gen := range c.gens {
		out, err := gen.Generate(ctx, prompt)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			break
		}
	}
	return "", fmt.Errorf("all %d generators in fallback chain failed, last error: %w", len(c.gens), lastErr)
}
