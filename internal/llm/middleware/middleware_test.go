package middleware

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/m7medvision/lazycommit/internal/domain"
)

type scriptedGenerator struct {
	outputs []string
	errs    []error
	calls   int
}

func (g *scriptedGenerator) Generate(context.Context, domain.Prompt) (string, error) {
	i := g.calls
	g.calls++
	if i >= len(g.outputs) {
		i = len(g.outputs) - 1
	}
	return g.outputs[i], g.errs[i]
}

type blockingGenerator struct{}

func (blockingGenerator) Generate(ctx context.Context, _ domain.Prompt) (string, error) {
	<-ctx.Done()
	return "", ctx.Err()
}

func TestWithTimeoutCancelsHungBackend(t *testing.T) {
	gen := WithTimeout(blockingGenerator{}, 10*time.Millisecond)
	start := time.Now()
	_, err := gen.Generate(context.Background(), domain.Prompt{})
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("timeout did not bound the call")
	}
}

func TestWithTimeoutPassesThroughSuccess(t *testing.T) {
	inner := &scriptedGenerator{outputs: []string{"ok"}, errs: []error{nil}}
	out, err := WithTimeout(inner, time.Second).Generate(context.Background(), domain.Prompt{})
	if err != nil || out != "ok" {
		t.Fatalf("unexpected: %q %v", out, err)
	}
}

func TestWithTimeoutZeroDisables(t *testing.T) {
	inner := &scriptedGenerator{outputs: []string{"ok"}, errs: []error{nil}}
	if got := WithTimeout(inner, 0); got != inner {
		t.Fatal("non-positive timeout should return the generator unchanged")
	}
}

func TestWithRetryEventualSuccess(t *testing.T) {
	inner := &scriptedGenerator{
		outputs: []string{"", "", "ok"},
		errs:    []error{errors.New("one"), errors.New("two"), nil},
	}
	out, err := WithRetry(inner, 3).Generate(context.Background(), domain.Prompt{})
	if err != nil || out != "ok" {
		t.Fatalf("unexpected: %q %v", out, err)
	}
	if inner.calls != 3 {
		t.Fatalf("calls = %d, want 3", inner.calls)
	}
}

func TestWithRetryExhaustionKeepsLastError(t *testing.T) {
	last := errors.New("final failure")
	inner := &scriptedGenerator{
		outputs: []string{"", ""},
		errs:    []error{errors.New("first"), last},
	}
	_, err := WithRetry(inner, 2).Generate(context.Background(), domain.Prompt{})
	if !errors.Is(err, last) {
		t.Fatalf("expected last error wrapped, got %v", err)
	}
	if inner.calls != 2 {
		t.Fatalf("calls = %d, want 2", inner.calls)
	}
}

func TestWithRetryStopsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	inner := &scriptedGenerator{outputs: []string{""}, errs: []error{errors.New("boom")}}
	_, err := WithRetry(inner, 5).Generate(ctx, domain.Prompt{})
	if err == nil {
		t.Fatal("expected error")
	}
	if inner.calls != 1 {
		t.Fatalf("must not retry after cancellation, calls = %d", inner.calls)
	}
}

func TestFallbackChainOrder(t *testing.T) {
	first := &scriptedGenerator{outputs: []string{""}, errs: []error{errors.New("first down")}}
	second := &scriptedGenerator{outputs: []string{"from second"}, errs: []error{nil}}
	third := &scriptedGenerator{outputs: []string{"unused"}, errs: []error{nil}}

	chain, err := NewFallbackChain(first, second, third)
	if err != nil {
		t.Fatal(err)
	}
	out, err := chain.Generate(context.Background(), domain.Prompt{})
	if err != nil || out != "from second" {
		t.Fatalf("unexpected: %q %v", out, err)
	}
	if third.calls != 0 {
		t.Fatal("chain must stop at first success")
	}
}

func TestFallbackChainExhaustion(t *testing.T) {
	last := errors.New("last cause")
	chain, err := NewFallbackChain(
		&scriptedGenerator{outputs: []string{""}, errs: []error{errors.New("a")}},
		&scriptedGenerator{outputs: []string{""}, errs: []error{last}},
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = chain.Generate(context.Background(), domain.Prompt{})
	if !errors.Is(err, last) {
		t.Fatalf("expected last cause wrapped, got %v", err)
	}
}

func TestFallbackChainRequiresGenerators(t *testing.T) {
	if _, err := NewFallbackChain(); err == nil {
		t.Fatal("expected error for empty chain")
	}
}

func TestFallbackChainSingleUnwrapped(t *testing.T) {
	inner := &scriptedGenerator{outputs: []string{"x"}, errs: []error{nil}}
	chain, err := NewFallbackChain(inner)
	if err != nil {
		t.Fatal(err)
	}
	if chain != inner {
		t.Fatal("single-element chain should return the generator itself")
	}
}
