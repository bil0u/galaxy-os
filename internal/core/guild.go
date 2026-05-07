package core

import (
	"context"

	"github.com/disgoorg/snowflake/v2"
)

// GuildManager owns guild discovery, onboarding, and teardown.
// Features never discover or iterate guilds directly.
type GuildManager interface {
	Onboard(ctx context.Context, guildID snowflake.ID) error
	Remove(ctx context.Context, guildID snowflake.ID) error
	Guilds() []snowflake.ID
	Accessor() GuildAccessor
}

// GuildAccessor provides read-only cross-guild access.
// Injected into CrossGuildScope features via Deps.Guilds.
type GuildAccessor interface {
	IDs() []snowflake.ID
	Config(guildID snowflake.ID, target any) error
}
