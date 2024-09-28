package bots

import (
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/bil0u/galaxy-os/sdk/commands"
	"github.com/bil0u/galaxy-os/sdk/components"
	"github.com/bil0u/galaxy-os/sdk/handlers"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

var HueCommands = []discord.ApplicationCommandCreate{
	commands.Test,
	commands.Version,
}

func HueEventListeners(b *sdk.Bot) []bot.EventListener {
	h := handler.New()

	h.Command("/test", commands.TestHandler)
	h.Autocomplete("/test", commands.TestAutocompleteHandler)
	h.Component("/test-button", components.TestComponent)

	h.Command("/version", commands.CreateVersionHandler(b))

	return []bot.EventListener{
		h,
		handlers.MessageHandler(b),
	}
}
