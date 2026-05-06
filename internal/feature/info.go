package feature

import (
	"context"

	"github.com/bil0u/galaxy-os/internal/platform"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

// Info displays bot version and commit information.
var Info = &info{}

type info struct{}

type infoConfig struct{}

func (c infoConfig) Validate() error { return nil }

func (f *info) Name() string               { return "info" }
func (f *info) Scope() platform.Scope       { return platform.BotScope }
func (f *info) Needs() []platform.ServiceID { return nil }

func (f *info) Setup(deps platform.Deps) error {
	// TODO: version/commit info was previously read from deps.Configs.Global
	// which is not available in platform.Deps. Wire when Global config is accessible.
	deps.Commands.SlashCommand("/botinfos", func(e *handler.CommandEvent) error {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Bot info unavailable (global config not wired yet)",
		})
	})
	return nil
}

func (f *info) Start(ctx context.Context) error { return nil }
func (f *info) Stop(ctx context.Context) error  { return nil }
