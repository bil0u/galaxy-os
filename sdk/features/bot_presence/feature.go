package bot_presence

import (
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/discord"
)

func init() {
	sdk.RegisterFeature[BotPresenceFeature]("bot_presence")
}

type BotPresenceFeature struct {
	Enabled  bool              `toml:"enabled"`
	Messages map[string]string `toml:"messages"`
}

func (f BotPresenceFeature) Name() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Bot Presence",
		discord.LocaleFrench:    "Présence du bot",
	}
}

func (f BotPresenceFeature) Description() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Set the bot presence based on multiple events",
		discord.LocaleFrench:    "Définir la présence du bot en fonction de plusieurs événements",
	}
}

func (f BotPresenceFeature) IsEnabled() bool {
	return f.Enabled
}

func (f BotPresenceFeature) Setup(bot *sdk.Bot) error {
	bot.Client.AddEventListeners(SetPresenceWhenReady(bot))
	return nil
}

func (f BotPresenceFeature) CommandsCreate() []discord.ApplicationCommandCreate {
	return nil
}
