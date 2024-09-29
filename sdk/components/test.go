package components

import (
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
)

func TestComponent(e *handler.ComponentEvent) error {
	return e.UpdateMessage(discord.MessageUpdate{
		Content: json.Ptr(sdk.LocalizedString{
			discord.LocaleEnglishUS: "The text has been updated",
			discord.LocaleFrench:    "Le texte a été mis à jour",
		}.String(e.Locale())),
	})
}
