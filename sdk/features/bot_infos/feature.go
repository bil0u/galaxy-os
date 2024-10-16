package bot_infos

import (
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/discord"
)

func init() {
	sdk.RegisterFeature[BotInfosFeature]("bot_infos")
}

type BotInfosFeature struct{}

func (f BotInfosFeature) Name() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Version",
		discord.LocaleFrench:    "Version",
	}
}

func (f BotInfosFeature) Description() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Display the bot bot-infos",
		discord.LocaleFrench:    "Affiche la bot-infos du bot",
	}
}

func (f BotInfosFeature) IsEnabled() bool {
	return true
}

func (f BotInfosFeature) Setup(bot *sdk.Bot) error {
	return nil
}

func (f BotInfosFeature) CommandsCreate() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		botInfosCommand,
	}
}
