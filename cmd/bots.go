package main

import (
	"github.com/bil0u/galaxy-os/cmd/generators"
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/bil0u/galaxy-os/sdk/commands"
	"github.com/bil0u/galaxy-os/sdk/components"
	"github.com/bil0u/galaxy-os/sdk/handlers"
	"github.com/bil0u/galaxy-os/sdk/utils"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
)

func init() {

	// DEVELOPMENT BOTS

	// Generator
	sdk.RegisterBotParts("generator",
		*sdk.NewBotParts().AddIntents(
			gateway.IntentsAll,
		).AddCaches(
			cache.FlagRoles,
			cache.FlagChannels,
		),
	)

	utils.RegisterGenerator(generators.RoleEnumGenerator)
	utils.RegisterGenerator(generators.ChannelEnumGenerator)

	// PRODUCTION BOTS

	// Hue bot
	sdk.RegisterBotParts("hue", sdk.BotParts{
		Intents: []gateway.Intents{
			gateway.IntentGuilds,
			gateway.IntentGuildMessages,
			gateway.IntentMessageContent,
			gateway.IntentDirectMessages,
		},
		Caches: []cache.Flags{
			cache.FlagGuilds,
			cache.FlagChannels,
			cache.FlagMembers,
			cache.FlagRoles,
		},
		Commands: []discord.ApplicationCommandCreate{
			commands.Test,
			commands.Version,
		},
		CreateListeners: func(b *sdk.Bot) []bot.EventListener {
			return []bot.EventListener{
				handlers.MessageHandler(b),
			}
		},
		CreateRouter: func(b *sdk.Bot) *handler.Mux {
			router := handler.New()
			router.Command("/test", commands.TestHandler)
			router.Autocomplete("/test", commands.TestAutocompleteHandler)
			router.Component("/test-button", components.TestComponent)

			router.Command("/version", commands.CreateVersionHandler(b))
			return router
		},
	})

	// Kevin bot
	sdk.RegisterBotParts("kevin", sdk.BotParts{
		Intents: []gateway.Intents{},
		Caches:  []cache.Flags{},
		Commands: []discord.ApplicationCommandCreate{
			commands.Version,
		},
		CreateListeners: func(b *sdk.Bot) []bot.EventListener {
			return []bot.EventListener{
				handlers.MessageHandler(b),
			}
		},
		CreateRouter: func(b *sdk.Bot) *handler.Mux {
			router := handler.New()
			router.Command("/version", commands.CreateVersionHandler(b))
			return router
		},
	})
}
