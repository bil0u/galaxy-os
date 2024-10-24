package features

import (
	"fmt"

	"github.com/bil0u/galaxy-os/pkg"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func init() {
	pkg.RegisterFeature[BotInfosFeature]("bot_infos")
}

type BotInfosFeature struct{}

func (f BotInfosFeature) Name() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "Bot Infos",
		discord.LocaleFrench:    "Bot Infos",
	}
}

func (f BotInfosFeature) Description() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "Display the bot informations",
		discord.LocaleFrench:    "Affiche les informations du bot",
	}
}

func (f BotInfosFeature) IsEnabled() bool {
	return true
}

func (f BotInfosFeature) IsProperlyConfigured() error {
	return nil
}

func (f BotInfosFeature) Setup(bot *pkg.Bot) error {
	bot.Router.Command("/botinfos", CreateBotInfosHandler(bot))
	bot.AddCommandsToSync(botInfosCommand)
	return nil
}

var botInfosCommand = discord.SlashCommandCreate{
	Name: "botinfos",
	NameLocalizations: pkg.LocalizedString{
		discord.LocaleEnglishUS: "bot-infos",
		discord.LocaleFrench:    "bot-infos",
	},
	Description: "Display the bot bot-infos",
	DescriptionLocalizations: pkg.LocalizedString{
		discord.LocaleEnglishUS: "Display the bot bot-infos",
		discord.LocaleFrench:    "Affiche la bot-infos du bot",
	},
}

func CreateBotInfosHandler(b *pkg.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Version: %s\nCommit: %s", b.Version, b.Commit),
		})
	}
}
