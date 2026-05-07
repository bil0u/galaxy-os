package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bil0u/galaxy-os/internal/core"
)

// FileStore reads TOML config files from disk.
// It implements core.ConfigStore.
type FileStore struct {
	dir string
}

// NewFileStore creates a FileStore rooted at the given directory.
func NewFileStore(dir string) *FileStore {
	return &FileStore{dir: dir}
}

// Load reads the raw bytes of the TOML file matching the given scope.
// Bot-level config (scope.GuildID == 0) reads "config.toml".
// Guild-level config reads "config.<guildID>.toml".
func (s *FileStore) Load(_ context.Context, scope core.ConfigScope) ([]byte, error) {
	name := "config.toml"
	if scope.GuildID != 0 {
		name = fmt.Sprintf("config.%s.toml", scope.GuildID)
	}
	path := filepath.Join(s.dir, name)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %q: %w", path, err)
	}
	return data, nil
}

// Watch returns a channel that signals when the config file changes.
// Hot-reload is deferred; this always returns nil.
func (s *FileStore) Watch(_ context.Context, _ core.ConfigScope) <-chan struct{} {
	return nil
}
