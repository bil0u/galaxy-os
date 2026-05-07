package config

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bil0u/galaxy-os/internal/contracts"
	"github.com/disgoorg/snowflake/v2"
	"github.com/spf13/viper"
)

// Resolver implements contracts.ScopedResolver by loading TOML config
// from a ConfigStore and extracting feature sections via viper.
type Resolver struct {
	store   contracts.ConfigStore
	botName string

	// guildIDs is the set of known guild IDs, populated after guild discovery.
	// ResolveAll iterates over these.
	guildIDs []snowflake.ID
}

// NewResolver creates a Resolver bound to a store and bot name.
func NewResolver(store contracts.ConfigStore, botName string) *Resolver {
	return &Resolver{
		store:   store,
		botName: botName,
	}
}

// SetGuildIDs sets the list of guild IDs that ResolveAll iterates over.
func (r *Resolver) SetGuildIDs(ids []snowflake.ID) {
	r.guildIDs = ids
}

// loadViper reads raw bytes from the store for the given scope and
// returns a viper instance with the parsed TOML content.
func (r *Resolver) loadViper(scope contracts.ConfigScope) (*viper.Viper, error) {
	data, err := r.store.Load(context.Background(), scope)
	if err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigType("toml")
	if err := v.ReadConfig(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("parsing config for scope %v: %w", scope, err)
	}
	return v, nil
}

// featurePath returns the viper key path for a feature section.
func (r *Resolver) featurePath(featureKey string) string {
	return "features." + r.botName + "." + featureKey
}

// ResolveBot unmarshals the bot-level feature config into target.
// It reads from [features.<botName>.<featureKey>] in config.toml.
func (r *Resolver) ResolveBot(featureKey string, target any) error {
	scope := contracts.ConfigScope{Bot: r.botName, GuildID: 0}
	v, err := r.loadViper(scope)
	if err != nil {
		return fmt.Errorf("loading bot config: %w", err)
	}

	sub := v.Sub(r.featurePath(featureKey))
	if sub == nil {
		// No config section found; leave target at zero values.
		return nil
	}
	if err := sub.Unmarshal(target); err != nil {
		return fmt.Errorf("unmarshaling feature %q bot config: %w", featureKey, err)
	}
	return nil
}

// ResolveGuild unmarshals the guild-level feature config into target,
// falling back to bot-level defaults when the guild file is missing
// or lacks the feature section.
func (r *Resolver) ResolveGuild(featureKey string, guildID snowflake.ID, target any) error {
	// Try bot-level defaults first.
	botScope := contracts.ConfigScope{Bot: r.botName, GuildID: 0}
	botV, err := r.loadViper(botScope)
	if err == nil {
		if sub := botV.Sub(r.featurePath(featureKey)); sub != nil {
			if err := sub.Unmarshal(target); err != nil {
				return fmt.Errorf("unmarshaling feature %q bot defaults: %w", featureKey, err)
			}
		}
	}

	// Overlay guild-level config on top.
	guildScope := contracts.ConfigScope{Bot: r.botName, GuildID: guildID}
	guildV, err := r.loadViper(guildScope)
	if err != nil {
		// No guild config file — bot defaults (if any) remain.
		return nil
	}

	sub := guildV.Sub(r.featurePath(featureKey))
	if sub == nil {
		return nil
	}
	if err := sub.Unmarshal(target); err != nil {
		return fmt.Errorf("unmarshaling feature %q guild %s config: %w", featureKey, guildID, err)
	}
	return nil
}

// ResolveAll returns a map of guild ID to config for every known guild.
// Each entry is unmarshaled independently via ResolveGuild, so the
// caller receives concrete values (not pointers). The returned map uses
// `any` to satisfy the contracts.ScopedResolver interface; callers
// should use contracts.ResolveAll[T] for typed access.
func (r *Resolver) ResolveAll(featureKey string) (map[snowflake.ID]any, error) {
	result := make(map[snowflake.ID]any, len(r.guildIDs))
	for _, guildID := range r.guildIDs {
		// We unmarshal into a map[string]any because we don't know the
		// concrete type here. The generic helpers in platform/ handle
		// the type assertion.
		var cfg map[string]any
		if err := r.ResolveGuild(featureKey, guildID, &cfg); err != nil {
			return nil, fmt.Errorf("resolving feature %q for guild %s: %w", featureKey, guildID, err)
		}
		if cfg != nil {
			result[guildID] = cfg
		}
	}
	return result, nil
}

// LoadGuildConfig loads and unmarshals the full guild-level config into target.
func (r *Resolver) LoadGuildConfig(guildID snowflake.ID, target any) error {
	scope := contracts.ConfigScope{Bot: r.botName, GuildID: guildID}
	v, err := r.loadViper(scope)
	if err != nil {
		return fmt.Errorf("loading guild config for %s: %w", guildID, err)
	}
	if err := v.Unmarshal(target); err != nil {
		return fmt.Errorf("unmarshaling guild config for %s: %w", guildID, err)
	}
	return nil
}

// Invalidate is a no-op until caching is added.
func (r *Resolver) Invalidate(_ contracts.ConfigScope) {}
