package cmd

import (
	"log/slog"

	"github.com/bil0u/galaxy-os/internal/contracts"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

// muxRegistrar adapts handler.Mux to contracts.Registrar.
// disgo v0.19.3 handler types include a typed data parameter;
// the contracts.Registrar signatures omit it for simplicity.
// Each handler wrapper includes panic recovery so a panicking
// feature handler is logged and the bot continues running.
type muxRegistrar struct {
	mux    *handler.Mux
	logger *slog.Logger
}

func (r *muxRegistrar) SlashCommand(path string, h contracts.SlashCommandHandler) {
	r.mux.SlashCommand(path, func(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
		defer func() {
			if rec := recover(); rec != nil {
				r.logger.Error("panic in slash command handler", slog.String("path", path), slog.Any("panic", rec))
			}
		}()
		return h(e)
	})
}

func (r *muxRegistrar) ButtonComponent(customID string, h contracts.ButtonComponentHandler) {
	r.mux.ButtonComponent(customID, func(_ discord.ButtonInteractionData, e *handler.ComponentEvent) error {
		defer func() {
			if rec := recover(); rec != nil {
				r.logger.Error("panic in button component handler", slog.String("custom_id", customID), slog.Any("panic", rec))
			}
		}()
		return h(e)
	})
}

func (r *muxRegistrar) Autocomplete(path string, h contracts.AutocompleteHandler) {
	r.mux.Autocomplete(path, func(e *handler.AutocompleteEvent) error {
		defer func() {
			if rec := recover(); rec != nil {
				r.logger.Error("panic in autocomplete handler", slog.String("path", path), slog.Any("panic", rec))
			}
		}()
		return h(e)
	})
}
