package discord

import (
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
)

var (
	Client         *bot.Client
	ShardedClient  *bot.Client
	guildsPerShard = 1
)

func InitClient(token string, caches []cache.Flags, intents []gateway.Intents) (*bot.Client, error) {
	configOpts := []bot.ConfigOpt{
		bot.WithCacheConfigOpts(cache.WithCaches(caches...)),
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(intents...),
			gateway.WithCompress(true),
		),
	}

	clt, err := disgo.New(token, configOpts...)
	if err != nil {
		return nil, fmt.Errorf("building disgo client: %w", err)
	}

	Client = &clt
	return Client, nil
}

func InitShardedClient(token string, shardCount int, caches []cache.Flags, intents []gateway.Intents) (*bot.Client, error) {
	if shardCount == 0 {
		return nil, fmt.Errorf("shard count cannot be 0")
	}

	shardedOpts := []bot.ConfigOpt{
		bot.WithCacheConfigOpts(cache.WithCaches(caches...)),
		bot.WithShardManagerConfigOpts(
			sharding.WithShardCount(shardCount),
			sharding.WithGatewayConfigOpts(
				gateway.WithIntents(intents...),
				gateway.WithCompress(true),
			),
		),
	}

	clt, err := disgo.New(token, shardedOpts...)
	if err != nil {
		return nil, fmt.Errorf("building disgo sharded client: %w", err)
	}

	ShardedClient = &clt
	return ShardedClient, nil
}

// OptimalShardCount calculates the optimal shard count for the bot
// based on the number of guilds the bot is in and a constant.
// Should be called after Init since we need the client. It returns -1 in case of error.
func OptimalShardCount() int {
	if Client == nil {
		slog.Error("Init must be called before OptimalShardCount")
		return -1
	}

	guilds, err := (*Client).Rest().GetCurrentUserGuilds("", 0, 0, 0, true)
	if err != nil {
		slog.Error("Failed to fetch bot guilds", slog.Any("error", err))
		return -1
	}

	return len(guilds) / guildsPerShard
}
