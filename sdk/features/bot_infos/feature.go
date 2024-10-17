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
		discord.LocaleEnglishUS: "Bot Infos",
		discord.LocaleFrench:    "Bot Infos",
	}
}

func (f BotInfosFeature) Description() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Display the bot informations",
		discord.LocaleFrench:    "Affiche les informations du bot",
	}
}

func (f BotInfosFeature) IsEnabled() bool {
	return true
}

func (f BotInfosFeature) Setup(bot *sdk.Bot) error {
	bot.Router.Command("/botinfos", CreateBotInfosHandler(bot))
	bot.AddCommandsToSync(botInfosCommand)
	return nil
}
