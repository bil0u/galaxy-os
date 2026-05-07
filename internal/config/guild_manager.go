package config

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/bil0u/galaxy-os/internal/contracts"
	"github.com/disgoorg/snowflake/v2"
)

// Manager owns guild lifecycle: onboarding, removal, and read-only access.
// It implements contracts.GuildManager and delegates config resolution
// to a Resolver for guild-level config loading.
type Manager struct {
	mu       sync.RWMutex
	guildIDs []snowflake.ID
	resolver *Resolver
	logger   *slog.Logger
}

// NewManager creates a Manager seeded with the initially discovered guild IDs.
// The resolver is updated to reflect the same set of guild IDs.
func NewManager(guildIDs []snowflake.ID, resolver *Resolver, logger *slog.Logger) *Manager {
	resolver.SetGuildIDs(guildIDs)
	return &Manager{
		guildIDs: slices.Clone(guildIDs),
		resolver: resolver,
		logger:   logger,
	}
}

// Onboard adds a guild to the active list and loads its config into the
// resolver. If the guild is already active, this is a no-op.
func (m *Manager) Onboard(_ context.Context, guildID snowflake.ID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if slices.Contains(m.guildIDs, guildID) {
		m.logger.Debug("guild already onboarded", slog.String("guild", guildID.String()))
		return nil
	}

	m.guildIDs = append(m.guildIDs, guildID)
	m.resolver.SetGuildIDs(m.guildIDs)
	m.logger.Info("guild onboarded", slog.String("guild", guildID.String()))
	return nil
}

// Remove removes a guild from the active list. If the guild is not active,
// this is a no-op.
func (m *Manager) Remove(_ context.Context, guildID snowflake.ID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	idx := slices.Index(m.guildIDs, guildID)
	if idx < 0 {
		m.logger.Debug("guild not found for removal", slog.String("guild", guildID.String()))
		return nil
	}

	m.guildIDs = slices.Delete(m.guildIDs, idx, idx+1)
	m.resolver.SetGuildIDs(m.guildIDs)
	m.logger.Info("guild removed", slog.String("guild", guildID.String()))
	return nil
}

// Guilds returns a copy of the active guild IDs.
func (m *Manager) Guilds() []snowflake.ID {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return slices.Clone(m.guildIDs)
}

// Accessor returns a read-only GuildAccessor backed by this manager.
func (m *Manager) Accessor() contracts.GuildAccessor {
	return &guildAccessor{mgr: m}
}

// guildAccessor provides read-only cross-guild access for CrossGuildScope features.
type guildAccessor struct {
	mgr *Manager
}

// IDs returns the active guild IDs.
func (a *guildAccessor) IDs() []snowflake.ID {
	return a.mgr.Guilds()
}

// Config resolves guild-level config into target. The target must be a
// pointer to a struct with mapstructure tags. The featureKey is not needed
// here because the accessor resolves raw guild config (not feature config).
// For feature-specific config, features use their ConfigProvider.
func (a *guildAccessor) Config(guildID snowflake.ID, target any) error {
	scope := contracts.ConfigScope{Bot: a.mgr.resolver.botName, GuildID: guildID}
	v, err := a.mgr.resolver.loadViper(scope)
	if err != nil {
		return fmt.Errorf("loading guild config for %s: %w", guildID, err)
	}
	if err := v.Unmarshal(target); err != nil {
		return fmt.Errorf("unmarshaling guild config for %s: %w", guildID, err)
	}
	return nil
}
