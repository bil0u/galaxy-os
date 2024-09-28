package sdk

import (
	"fmt"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
)

// BotParts holds the pieces for a bot
type BotParts struct {
	Commands  func(b *Bot) []discord.ApplicationCommandCreate
	Listeners func(b *Bot) []bot.EventListener
}

var (
	// Registry holds the registered bots
	partsRegistry = map[string]BotParts{}
)

// RegisterBot registers a bot with its pieces
func RegisterBotParts(name string, config BotParts) {
	if _, exists := partsRegistry[name]; exists {
		panic(fmt.Sprintf("Bot %s is already registered", name))
	}
	partsRegistry[name] = config
}

// GetBotParts retrieves the pieces for a bot
func GetBotParts(name string) (BotParts, error) {
	config, exists := partsRegistry[name]
	if !exists {
		return BotParts{}, fmt.Errorf("unknown bot: %s", name)
	}
	return config, nil
}
