package presence

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bil0u/galaxy-os/internal/core"
	"github.com/disgoorg/snowflake/v2"
)

// Presence automatically sets the bot presence based on guild config.
var Feature = &presence{}

type presence struct {
	logger  *slog.Logger
	configs core.ConfigProvider
	guildID snowflake.ID
}

type presenceConfig struct {
	Enabled  bool
	Messages map[string]string
}

func (c presenceConfig) Validate() error {
	var errs []error
	if len(c.Messages) == 0 {
		errs = append(errs, fmt.Errorf("messages are required"))
	}
	_, ok := c.Messages["bot_ready"]
	if !ok {
		errs = append(errs, fmt.Errorf("message 'bot_ready' is required"))
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %v", errs)
	}
	return nil
}

func (f *presence) Name() string                 { return "presence" }
func (f *presence) Scope() core.Scope       { return core.GuildScope }
func (f *presence) Needs() []core.ServiceID { return nil }

func (f *presence) Setup(deps core.Deps) error {
	f.logger = deps.Logger
	f.configs = deps.Configs
	f.guildID = deps.GuildID

	// TODO: presence setting requires gateway client access (client.SetPresenceForShard)
	// and event listeners (bot.NewListenerFunc for events.Ready), neither of which are
	// available in core.Deps. Wire when gateway/client access is added to Deps.

	return nil
}

func (f *presence) Start(ctx context.Context) error { return nil }
func (f *presence) Stop(ctx context.Context) error  { return nil }
