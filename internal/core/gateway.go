package core

import (
	"context"

	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
)

// DiscordGateway exposes gateway-level operations: event subscription, presence, and bot identity.
// *bot.Client satisfies this interface directly — no adapter needed.
type DiscordGateway interface {
	AddEventListeners(listeners ...disbot.EventListener)
	SetPresence(ctx context.Context, opts ...gateway.PresenceOpt) error
	ID() snowflake.ID
}
