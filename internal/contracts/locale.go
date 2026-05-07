package contracts

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

// LocaleResolver unifies guild and user locale with fallback.
type LocaleResolver interface {
	GuildLocale(guildID snowflake.ID) discord.Locale
	UserLocale(interaction discord.Interaction) discord.Locale
}
