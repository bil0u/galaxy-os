package discord

import (
	"fmt"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
)

var client bot.Client

func InitClient(token string, cacheFlags cache.Flags, intents gateway.Intents) (bot.Client, error) {
	configOpts := []bot.ConfigOpt{
		bot.WithCacheConfigOpts(cache.WithCaches(cacheFlags)),
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(intents),
			gateway.WithCompress(true),
		),
	}

	clt, err := disgo.New(token, configOpts...)
	if err != nil {
		return nil, fmt.Errorf("building disgo client: %w", err)
	}

	client = clt
	return client, nil
}

func Client() bot.Client {
	return client
}
