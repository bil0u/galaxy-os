package main

import (
	"github.com/bil0u/galaxy-os/cmd/generators"
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/bil0u/galaxy-os/sdk/features/bot_infos"
	"github.com/bil0u/galaxy-os/sdk/features/bot_presence"
	"github.com/bil0u/galaxy-os/sdk/features/daily_message"
	"github.com/bil0u/galaxy-os/sdk/features/self_assign_roles"
	"github.com/bil0u/galaxy-os/sdk/features/suspicious_interview"
	"github.com/bil0u/galaxy-os/sdk/features/test_command"
)

var BotsFeatures = map[string]sdk.BotFeatureSet{}

func init() {

	sdk.RegisterGenerator(generators.RoleEnumGenerator)
	sdk.RegisterGenerator(generators.ChannelEnumGenerator)

	// Hue Features
	BotsFeatures["hue"] = sdk.BotFeatureSet{
		bot_infos.BotInfosFeature{},
		bot_presence.BotPresenceFeature{},
		self_assign_roles.SelfAssignRolesFeature{},
		daily_message.DailyMessageFeature{},
		suspicious_interview.SuspiciousInterwiewFeature{},
		test_command.TestCommandFeature{},
	}

	// Kevin Features
	BotsFeatures["kevin"] = sdk.BotFeatureSet{
		bot_infos.BotInfosFeature{},
		bot_presence.BotPresenceFeature{},
		self_assign_roles.SelfAssignRolesFeature{},
	}

	// DEVELOPMENT BOTS

	// Generator
	// sdk.RegisterBotParts("generator",
	// 	*sdk.NewBotParts().AddIntents(
	// 		gateway.IntentsAll,
	// 	).AddCaches(
	// 		cache.FlagRoles,
	// 		cache.FlagChannels,
	// 	),
	// )

	// PRODUCTION BOTS

	// Hue bot
	// sdk.RegisterBotParts("hue", sdk.BotParts{
	// 	Intents: []gateway.Intents{
	// 		gateway.IntentGuilds,
	// 		gateway.IntentGuildMessages,
	// 		gateway.IntentMessageContent,
	// 		gateway.IntentDirectMessages,
	// 	},
	// 	Caches: []cache.Flags{
	// 		cache.FlagGuilds,
	// 		cache.FlagChannels,
	// 		cache.FlagMembers,
	// 		cache.FlagRoles,
	// 	},
	// 	Commands: []discord.ApplicationCommandCreate{
	// 		commands.Test,
	// 		commands.Version,
	// 	},
	// 	CronJobs: []sdk.CronJob{
	// 		{
	// 			Name:     "DailyMessage",
	// 			Schedule: "30 10 * * *",
	// 			Creator:  jobs.DailyMessageJob,
	// 		},
	// 	},
	// 	CreateListeners: func(b *sdk.Bot) []bot.EventListener {
	// 		return []bot.EventListener{
	// 			handlers.OnMessageCreate(b),
	// 			handlers.OnRoleAssigned(b),
	// 		}
	// 	},
	// 	CreateRouter: func(b *sdk.Bot) *handler.Mux {
	// 		router := handler.New()
	// 		router.Command("/test", commands.TestHandler)
	// 		router.Autocomplete("/test", commands.TestAutocompleteHandler)
	// 		router.Component("/test-button", components.TestComponent)

	// 		router.Command("/version", commands.CreateVersionHandler(b))
	// 		return router
	// 	},
	// })

	// Kevin bot
	// sdk.RegisterBotParts("kevin", sdk.BotParts{
	// 	Intents: []gateway.Intents{},
	// 	Caches:  []cache.Flags{},
	// 	Commands: []discord.ApplicationCommandCreate{
	// 		commands.Version,
	// 	},
	// 	CronJobs: []sdk.CronJob{},
	// 	CreateListeners: func(b *sdk.Bot) []bot.EventListener {
	// 		return []bot.EventListener{
	// 			handlers.OnMessageCreate(b),
	// 		}
	// 	},
	// 	CreateRouter: func(b *sdk.Bot) *handler.Mux {
	// 		router := handler.New()
	// 		router.Command("/version", commands.CreateVersionHandler(b))
	// 		return router
	// 	},
	// })
}
