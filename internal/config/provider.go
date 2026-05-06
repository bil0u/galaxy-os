package config

import (
	"github.com/disgoorg/snowflake/v2"
)

// Provider implements platform.ConfigProvider by delegating to a
// Resolver with a pre-bound feature key. The framework builds one
// Provider per feature so features never see the resolver directly.
type Provider struct {
	resolver   *Resolver
	featureKey string
}

// NewProvider creates a Provider for the given feature key.
func NewProvider(resolver *Resolver, featureKey string) *Provider {
	return &Provider{
		resolver:   resolver,
		featureKey: featureKey,
	}
}

// Bot unmarshals the bot-level config for this feature into target.
func (p *Provider) Bot(target any) error {
	return p.resolver.ResolveBot(p.featureKey, target)
}

// Guild unmarshals the guild-level config for this feature into target,
// with bot defaults as fallback.
func (p *Provider) Guild(guildID snowflake.ID, target any) error {
	return p.resolver.ResolveGuild(p.featureKey, guildID, target)
}

// All returns a map of guild ID to config for every known guild.
func (p *Provider) All() (map[snowflake.ID]any, error) {
	return p.resolver.ResolveAll(p.featureKey)
}

// OnChange registers a callback for config changes. Hot-reload is
// deferred, so this is currently a no-op.
func (p *Provider) OnChange(_ func()) {}
