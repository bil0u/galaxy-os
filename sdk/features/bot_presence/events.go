package bot_presence

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bil0u/galaxy-os/sdk"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
)

func SetPresenceWhenReady(bot *sdk.Bot) disbot.EventListener {
	return disbot.NewListenerFunc(func(event *events.Ready) {

		slog.Info(fmt.Sprintf("Bot '%s' is ready", bot.Name))

		for _, guildConfig := range bot.Config.Guilds {

			feature, _ := sdk.GetFeature[*BotPresenceFeature](guildConfig.Features)

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
