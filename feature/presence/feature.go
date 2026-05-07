package presence

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bil0u/galaxy-os/internal/core"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
)

var Feature = &presence{}

type presence struct{}

type presenceConfig struct {
	Enabled  bool
	Messages map[string]string
}

func (c presenceConfig) Validate() error {
	if len(c.Messages) == 0 {
		return fmt.Errorf("messages are required")
	}
	if _, ok := c.Messages["bot_ready"]; !ok {
		return fmt.Errorf("message 'bot_ready' is required")
	}
	return nil
}

func (f *presence) Name() string            { return "presence" }
func (f *presence) Scope() core.Scope       { return core.GuildScope }
func (f *presence) Needs() []core.ServiceID { return nil }

func (f *presence) Setup(deps core.Deps) error {
	guildID := deps.GuildID
	logger := deps.Logger
	configs := deps.Configs
	gw := deps.Gateway

	gw.AddEventListeners(disbot.NewListenerFunc(func(_ *events.Ready) {
		cfg, err := core.ResolveGuild[presenceConfig](configs, guildID)
		if err != nil {
			logger.Error("resolving config", slog.Any("error", err))
			return
		}
		if !cfg.Enabled {
			return
		}
		if err := cfg.Validate(); err != nil {
			logger.Error("invalid config", slog.Any("error", err))
			return
		}

		msg := cfg.Messages["bot_ready"]
		if err := gw.SetPresence(context.Background(),
			gateway.WithOnlineStatus(discord.OnlineStatusOnline),
			gateway.WithPlayingActivity(msg),
		); err != nil {
			logger.Error("setting presence", slog.Any("error", err))
		}
	}))

	return nil
}

func (f *presence) Start(_ context.Context) error { return nil }
func (f *presence) Stop(_ context.Context) error  { return nil }
