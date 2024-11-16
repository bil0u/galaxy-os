package guild_features

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bil0u/galaxy-os/internal/config"
	"github.com/bil0u/galaxy-os/internal/features"
	"github.com/bil0u/galaxy-os/internal/services"
	"github.com/bil0u/galaxy-os/internal/utils"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
)

var BotPresenceFeature = features.New[BotPresenceConfig](
	setupBotPresenceFeature,
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

func setupBotPresenceFeature() error {

	client := services.GetDiscordShardedClient()

	if client == nil {
		return fmt.Errorf("client is nil")
	}

	(*client).AddEventListeners(bot.NewListenerFunc(setPresenceWhenReady))
	return nil
}

func setPresenceWhenReady(event *events.Ready) {

	client := services.GetDiscordShardedClient()

	for guildID, _ := range config.Guilds.All() {

		cfg, _ := features.GetConfig[BotPresenceConfig](guildID)

		if !cfg.Enabled {
			slog.Warn(fmt.Sprintf("Feature 'BotPresence' is disabled for guild '%s'", guildID))
			continue
		}

		// Get Shard ID from Guild ID
		shardID := sharding.ShardIDByGuild(guildID, config.Guilds.Count())

		// Set presence for shard
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		slog.Info(fmt.Sprintf("Setting presence for guild '%s'", guildID))
		if err := (*client).SetPresenceForShard(ctx, shardID, gateway.WithCustomActivity(cfg.Messages["bot_ready"]), gateway.WithOnlineStatus(discord.OnlineStatusOnline)); err != nil {
			slog.Error(fmt.Sprintf("Failed to set presence for guild '%s'", guildID), slog.Any("err", err))
		}

	}
}
