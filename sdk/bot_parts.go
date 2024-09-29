package sdk

import (
	"fmt"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
)

// NewClient creates a new disgo client with the provided token and parts
func NewBotClient(token string, parts BotParts) (*bot.Client, error) {
	client, err := disgo.New(token,
		bot.WithGatewayConfigOpts(gateway.WithIntents(parts.Intents...)),
		bot.WithCacheConfigOpts(cache.WithCaches(parts.Caches...)),
	)
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func NewBotParts() *BotParts {
	return &BotParts{
		Intents:         []gateway.Intents{},
		Caches:          []cache.Flags{},
		Commands:        []discord.ApplicationCommandCreate{},
		CreateListeners: func(b *Bot) []bot.EventListener { return nil },
		CreateRouter:    func(b *Bot) *handler.Mux { return nil },
	}
}

// BotParts holds the pieces for a bot
type BotParts struct {
	Intents         []gateway.Intents
	Caches          []cache.Flags
	Commands        []discord.ApplicationCommandCreate
	CreateListeners func(b *Bot) []bot.EventListener
	CreateRouter    func(b *Bot) *handler.Mux
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

func (p *BotParts) SetCreateListeners(createListeners func(b *Bot) []bot.EventListener) *BotParts {
	p.CreateListeners = createListeners
	return p
}

func (p *BotParts) SetCreateRouter(createRouter func(b *Bot) *handler.Mux) *BotParts {
	p.CreateRouter = createRouter
	return p
}
