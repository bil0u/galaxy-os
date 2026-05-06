package platform

import "github.com/disgoorg/disgo/bot"

// OnEvent registers a typed gateway event handler on the client.
// Covers the full Discord event surface with zero per-event boilerplate.
func OnEvent[E bot.Event](client *bot.Client, h func(e E)) {
	client.AddEventListeners(bot.NewListenerFunc(h))
}
