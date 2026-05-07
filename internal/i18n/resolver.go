package i18n

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

// LocaleResolver provides a default-locale fallback for guild and user locale
// lookups. Satisfies platform.LocaleResolver.
type LocaleResolver struct {
	defaultLocale discord.Locale
}

// NewLocaleResolver returns a LocaleResolver with the given default locale.
func NewLocaleResolver(defaultLocale discord.Locale) *LocaleResolver {
	return &LocaleResolver{defaultLocale: defaultLocale}
}

func (r *LocaleResolver) GuildLocale(guildID snowflake.ID) discord.Locale {
	return r.defaultLocale
}

func (r *LocaleResolver) UserLocale(interaction discord.Interaction) discord.Locale {
	if interaction.Locale() != "" {
		return interaction.Locale()
	}
	return r.defaultLocale
}
