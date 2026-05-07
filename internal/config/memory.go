package config

import (
	"context"
	"fmt"

	"github.com/bil0u/galaxy-os/internal/contracts"
)

// MemoryStore is an in-memory contracts.ConfigStore for tests.
type MemoryStore struct {
	data map[contracts.ConfigScope][]byte
}

// NewMemoryStore creates an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[contracts.ConfigScope][]byte),
	}
}

// Set stores raw config bytes for the given scope.
func (m *MemoryStore) Set(scope contracts.ConfigScope, data []byte) {
	m.data[scope] = data
}

// Load returns the stored bytes for the given scope, or an error if
// no data has been set.
func (m *MemoryStore) Load(_ context.Context, scope contracts.ConfigScope) ([]byte, error) {
	data, ok := m.data[scope]
	if !ok {
		return nil, fmt.Errorf("no config data for scope %v", scope)
	}
	return data, nil
}

// Watch returns nil; the memory store does not support change notifications.
func (m *MemoryStore) Watch(_ context.Context, _ contracts.ConfigScope) <-chan struct{} {
	return nil
}
