package platform

import (
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/discord"
)

// Env discriminates deployment environments.
type Env int

const (
	Prod Env = iota
	Dev
	Test
)

// BotSpec declares a bot's identity, features, and gateway configuration.
// Defined in cmd/bot.go. Immutable after init.
type BotSpec struct {
	Name     string
	Env      Env
	Features []Feature
	Intents  gateway.Intents
	Presence *PresenceConfig
}

// PresenceConfig declares the bot's initial Discord presence.
type PresenceConfig struct {
	Status   discord.OnlineStatus
	Activity *discord.Activity
}
