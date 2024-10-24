package features

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bil0u/galaxy-os/pkg"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
)

func init() {
	pkg.RegisterFeature[BotPresenceFeature]("bot_presence")
}

type BotPresenceFeature struct {
	Enabled  bool              `toml:"enabled"`
	Messages map[string]string `toml:"messages"`
}

func (f BotPresenceFeature) Name() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "Bot Presence",
		discord.LocaleFrench:    "Présence du bot",
	}
}

func (f BotPresenceFeature) Description() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "Set the bot presence based on multiple events",
		discord.LocaleFrench:    "Définir la présence du bot en fonction de plusieurs événements",
	}
}

func (f BotPresenceFeature) IsEnabled() bool {
	return f.Enabled
}

func (f BotPresenceFeature) IsProperlyConfigured() error {
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

func (f BotPresenceFeature) Setup(bot *pkg.Bot) error {
	bot.Client.AddEventListeners(SetPresenceWhenReady(bot))
	return nil
}

func SetPresenceWhenReady(bot *pkg.Bot) disbot.EventListener {
	return disbot.NewListenerFunc(func(event *events.Ready) {

		slog.Info(fmt.Sprintf("Bot '%s' is ready", bot.Name))

		for _, guildConfig := range bot.Config.Guilds {

			feature, _ := pkg.GetFeature[*BotPresenceFeature](guildConfig.Features)

			if !feature.Enabled {
				continue
			}

			slog.Info(fmt.Sprintf("Setting presence for guild '%s'", guildConfig.ID))

			// Get Shard ID from Guild ID
			shardID := sharding.ShardIDByGuild(guildConfig.ID, len(bot.Config.Guilds))

			// Set presence for each shard
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := bot.Client.SetPresenceForShard(ctx, shardID, gateway.WithCustomActivity(feature.Messages["bot_ready"]), gateway.WithOnlineStatus(discord.OnlineStatusOnline)); err != nil {
				slog.Error(fmt.Sprintf("Failed to set presence for guild '%s'", guildConfig.ID), slog.Any("err", err))
			}

		}
	})
}
