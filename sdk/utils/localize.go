package utils

import (
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
)

type LocalizedString map[discord.Locale]string

func (str LocalizedString) String(locale discord.Locale) string {
	if val, ok := str[locale]; ok {
		return val
	}
	if val, ok := str[discord.LocaleEnglishUS]; ok {
		return val
	}
	slog.Warn("No localization found for locale", slog.Any("locale", locale))
	return fmt.Sprintf("No localization found for locale %s", locale)
}
