package commands

import (
	"fmt"

	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

var Version = discord.SlashCommandCreate{
	Name:        "version",
	Description: "Affiche la version du bot",
}

func CreateVersionHandler(b *sdk.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Version: %s\nCommit: %s", b.Version, b.Commit),
		})
	}
}
