package bot_infos

import (
	"fmt"

	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

var botInfosCommand = discord.SlashCommandCreate{
	Name: "bot-infos",
	NameLocalizations: sdk.LocalizedString{
		discord.LocaleEnglishUS: "bot-infos",
		discord.LocaleFrench:    "bot-infos",
	},
	Description: "Display the bot bot-infos",
	DescriptionLocalizations: sdk.LocalizedString{
		discord.LocaleEnglishUS: "Display the bot bot-infos",
		discord.LocaleFrench:    "Affiche la bot-infos du bot",
	},
}

func CreateBotInfosHandler(b *sdk.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Version: %s\nCommit: %s", b.Version, b.Commit),
		})
	}
}
