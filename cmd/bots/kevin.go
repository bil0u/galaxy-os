package bots

import (
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/bil0u/galaxy-os/sdk/commands"
	"github.com/bil0u/galaxy-os/sdk/components"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

var KevinRouter sdk.RouterInitializer = func(b *sdk.Bot, h *handler.Mux) []discord.ApplicationCommandCreate {

	h.Command("/version", commands.VersionHandler(b))

	h.Command("/test", commands.TestHandler)
	h.Autocomplete("/test", commands.TestAutocompleteHandler)
	h.Component("/test-button", components.TestComponent)

	return []discord.ApplicationCommandCreate{
		commands.Test,
		commands.Version,
	}
}
