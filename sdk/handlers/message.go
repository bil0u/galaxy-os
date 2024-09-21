package handlers

import (
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
)

func MessageHandler(b *sdk.Bot) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.MessageCreate) {
		// TODO: handle message
	})
}
