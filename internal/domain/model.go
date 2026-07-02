package domain

import (
	"errors"
	"strings"
)

// ModelID identifies an LLM model within a backend.
type ModelID struct {
	id string
}

func NewModelID(id string) (ModelID, error) {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return ModelID{}, errors.New("model id is empty")
	}
	return ModelID{id: trimmed}, nil
}

func (m ModelID) String() string {
	return m.id
}
