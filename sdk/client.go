package sdk

import (
	"fmt"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
	"github.com/disgoorg/snowflake/v2"
)

// NewBotClient creates a new bot client, with the provided token and parts
func NewBotClient(token string, intents []gateway.Intents, caches []cache.Flags) (bot.Client, error) {
	client, err := disgo.New(token,
		bot.WithCacheConfigOpts(cache.WithCaches(caches...)),
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(intents...),
			gateway.WithCompress(true),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error while building disgo client: %w", err)
	}
	return client, nil
}

// NewBotShardedClient creates a new bot client, with the provided token and parts, and shards the bot using the provided shard count
func NewBotShardedClient(shardCount int, token string, intents []gateway.Intents, caches []cache.Flags) (bot.Client, error) {
	client, err := disgo.New(token,
		bot.WithCacheConfigOpts(cache.WithCaches(caches...)),
		bot.WithShardManagerConfigOpts(
			sharding.WithShardCount(shardCount),
			sharding.WithGatewayConfigOpts(
				gateway.WithIntents(intents...),
				gateway.WithCompress(true),
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error while building disgo sharded client: %w", err)
	}
	return client, nil
}

func GetGuildIDFromShardID(shardID int, guilds []snowflake.ID) snowflake.ID {
	// Iterate over all guilds and find one whose ShardID matches the given ShardID
	for _, guildID := range guilds {
		if sharding.ShardIDByGuild(guildID, len(guilds)) == shardID {
			return guildID
		}
	}
	return 0 // or return an error
}
