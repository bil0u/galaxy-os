package bot_features

import (
	"fmt"

	"github.com/bil0u/galaxy-os/internal/config"
	"github.com/bil0u/galaxy-os/internal/features"
	"github.com/bil0u/galaxy-os/internal/locale"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

var BotInfosFeature = features.New[BotInfosConfig](
	setupBotInfosFeature,
	features.WithType(features.BotFeature),
	features.WithLocalizedName(discord.LocaleFrench, "Bot Infos"),
	features.WithDescription(locale.Text{
		discord.LocaleEnglishUS: "Display the bot informations",
		discord.LocaleFrench:    "Affiche les informations du bot",
	}),
	features.WithCommandsToSync(botInfosCommand),
)

type BotInfosConfig struct{}

func (f BotInfosConfig) Validate() error {
	return nil
}

func setupBotInfosFeature(deps features.SetupDeps) error {
	deps.Router.Command("/botinfos", botInfosHandler)
	return nil
}

var botInfosCommand = discord.SlashCommandCreate{
	Name: "botinfos",
	NameLocalizations: locale.Text{
		discord.LocaleEnglishUS: "bot-infos",
		discord.LocaleFrench:    "bot-infos",
	},
	Description: "Display the bot informations",
	DescriptionLocalizations: locale.Text{
		discord.LocaleEnglishUS: "Display the bot informations",
		discord.LocaleFrench:    "Affiche les informations du bot",
	},
}

func botInfosHandler(e *handler.CommandEvent) error {
	return e.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("Version: %s\nCommit: %s", config.GlobalCfg.Version, config.GlobalCfg.Commit),
	})
}
