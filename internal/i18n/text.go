package i18n

import "github.com/disgoorg/disgo/discord"

// Text is a map of localized strings keyed by Discord locale.
type Text map[discord.Locale]string

func (t Text) String() string {
	for _, val := range t {
		return val
	}
	return "No localization found"
}

func (t Text) Using(locale discord.Locale) string {
	if val, ok := t[locale]; ok {
		return val
	}
	if val, ok := t[discord.LocaleEnglishUS]; ok {
		return val
	}
	return t.String()
}
