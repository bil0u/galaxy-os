package core

import (
	"context"

	"github.com/disgoorg/snowflake/v2"
)

// ConfigScope identifies a config document (bot-level or guild-level).
type ConfigScope struct {
	Bot     string
	GuildID snowflake.ID
}

// ConfigStore abstracts where config data comes from.
// Implementations: FileStore (TOML/viper), MemoryStore (tests).
type ConfigStore interface {
	Load(ctx context.Context, scope ConfigScope) ([]byte, error)
	Watch(ctx context.Context, scope ConfigScope) <-chan struct{}
}

// ConfigProvider is the feature-facing config accessor.
// The framework pre-binds it to the feature's config type and bot name.
type ConfigProvider interface {
	Bot(target any) error
	Guild(guildID snowflake.ID, target any) error
	All() (map[snowflake.ID]any, error)
	OnChange(fn func())
}

// ScopedResolver merges bot defaults + guild overrides, caches results,
// and invalidates on change signals from the store.
// Features never use this directly — they use ConfigProvider.
type ScopedResolver interface {
	ResolveBot(featureKey string, target any) error
	ResolveGuild(featureKey string, guildID snowflake.ID, target any) error
	ResolveAll(featureKey string) (map[snowflake.ID]any, error)
	Invalidate(scope ConfigScope)
}

// ResolveBot is a generic convenience for extracting typed bot config.
func ResolveBot[T Config](p ConfigProvider) (T, error) {
	var t T
	err := p.Bot(&t)
	return t, err
}

// ResolveGuild is a generic convenience for extracting typed guild config.
func ResolveGuild[T Config](p ConfigProvider, guildID snowflake.ID) (T, error) {
	var t T
	err := p.Guild(guildID, &t)
	return t, err
}

// ResolveAll is a generic convenience for extracting all typed guild configs.
func ResolveAll[T Config](p ConfigProvider) (map[snowflake.ID]T, error) {
	raw, err := p.All()
	if err != nil {
		return nil, err
	}
	result := make(map[snowflake.ID]T, len(raw))
	for id, v := range raw {
		t, ok := v.(T)
		if !ok {
			continue
		}
		result[id] = t
	}
	return result, nil
}
