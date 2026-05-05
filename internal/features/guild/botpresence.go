package guild_features

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bil0u/galaxy-os/internal/config"
	"github.com/bil0u/galaxy-os/internal/features"
	"github.com/bil0u/galaxy-os/internal/utils"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
)

var BotPresenceFeature = features.New[BotPresenceConfig](
	setupBotPresenceFeature,
	features.WithType(features.GuildFeature),
	features.WithLocalizedName(discord.LocaleFrench, "Présence du bot"),
	features.WithDescription(utils.LocalizedString{
		discord.LocaleEnglishUS: "Automatically sets the bot presence based on multiple events",
		discord.LocaleFrench:    "Définit automatiquement la présence du bot en fonction de plusieurs événements",
	}),
)

type BotPresenceConfig struct {
	Enabled  bool
	Messages map[string]string
}

func (f BotPresenceConfig) Validate() error {
	var errs []error
	if len(f.Messages) == 0 {
		errs = append(errs, fmt.Errorf("messages are required"))
	}
	_, ok := f.Messages["bot_ready"]
	if !ok {
		errs = append(errs, fmt.Errorf("message 'bot_ready' is required"))
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %v", errs)
	}
	return nil
}

func setupBotPresenceFeature(deps features.SetupDeps) error {
	deps.Client.AddEventListeners(bot.NewListenerFunc(func(_ *events.Ready) {
		setPresenceWhenReady(deps.Client)
	}))
	return nil
}

func setPresenceWhenReady(client bot.Client) {
	for guildID := range config.Guilds.All() {
		cfg, err := features.GetConfig[BotPresenceConfig](guildID)
		if err != nil || !cfg.Enabled {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = client.SetPresenceForShard(ctx, 0, gateway.WithCustomActivity(cfg.Messages["bot_ready"]), gateway.WithOnlineStatus(discord.OnlineStatusOnline))
		cancel()
		if err != nil {
			slog.Error(fmt.Sprintf("failed to set presence for guild '%s'", guildID), slog.Any("err", err))
		}
		slog.Info(fmt.Sprintf("Presence set using config from guild '%s'", guildID))
		return
	}
	slog.Warn("no enabled BotPresence config found")
}
