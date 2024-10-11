package sdk

import (
	"fmt"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/robfig/cron/v3"
)

func NewBotParts() *BotParts {
	return &BotParts{
		Intents:         []gateway.Intents{},
		Caches:          []cache.Flags{},
		Commands:        []discord.ApplicationCommandCreate{},
		CronJobs:        []CronJob{},
		CreateListeners: func(b *Bot) []bot.EventListener { return nil },
		CreateRouter:    func(b *Bot) *handler.Mux { return nil },
	}
}

// BotParts holds the pieces for a bot
type BotParts struct {
	Intents         []gateway.Intents
	Caches          []cache.Flags
	Commands        []discord.ApplicationCommandCreate
	CronJobs        []CronJob
	CreateListeners func(b *Bot) []bot.EventListener
	CreateRouter    func(b *Bot) *handler.Mux
}

type CronJob struct {
	Name     string
	Schedule string
	Creator  func(b *Bot, guildID snowflake.ID) cron.FuncJob
}

var partsRegistry = map[string]BotParts{}

// RegisterBotParts registers the bot parts for the bot
func RegisterBotParts(botName string, parts BotParts) error {
	if _, exists := partsRegistry[botName]; exists {
		return fmt.Errorf("bot parts for bot %s already registered", botName)
	}
	partsRegistry[botName] = parts
	return nil
}

// GetBotParts returns the bot parts for the bot
func GetBotParts(botName string) (BotParts, error) {
	parts, ok := partsRegistry[botName]
	if !ok {
		return BotParts{}, fmt.Errorf("bot parts for bot %s not registered", botName)
	}
	return parts, nil
}

func (p *BotParts) AddIntents(intents ...gateway.Intents) *BotParts {
	p.Intents = append(p.Intents, intents...)
	return p
}

func (p *BotParts) AddCaches(caches ...cache.Flags) *BotParts {
	p.Caches = append(p.Caches, caches...)
	return p
}

func (p *BotParts) AddCommands(commands ...discord.ApplicationCommandCreate) *BotParts {
	p.Commands = append(p.Commands, commands...)
	return p
}

func (p *BotParts) AddCronJobs(jobs ...CronJob) *BotParts {
	p.CronJobs = append(p.CronJobs, jobs...)
	return p
}
