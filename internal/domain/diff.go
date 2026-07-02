package domain

import (
	"errors"
	"strings"
)

// ErrEmptyDiff signals that there are no changes to describe. Callers use it
// to short-circuit before any backend is invoked.
var ErrEmptyDiff = errors.New("diff is empty")

// Diff is a non-empty git diff.
type Diff struct {
	text string
}

func NewDiff(text string) (Diff, error) {
	if strings.TrimSpace(text) == "" {
		return Diff{}, ErrEmptyDiff
	}
	return Diff{text: text}, nil
}

func (d Diff) String() string {
	return d.text
}
