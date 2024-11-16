package services

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	cronjobs "github.com/bil0u/galaxy-os/internal/services/cron"
	"github.com/bil0u/galaxy-os/internal/services/discord"
	"github.com/bil0u/galaxy-os/internal/services/email"
	"github.com/bil0u/galaxy-os/internal/services/logger"
	"github.com/bil0u/galaxy-os/internal/services/oauth"
	"github.com/bil0u/galaxy-os/internal/services/sql"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/oauth2"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/paginator"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
)

// DISCORD

func GetClient() *bot.Client {
	if discord.ShardedClient != nil {
		return discord.ShardedClient
	}
	return discord.Client
}

func GetRestClient() rest.Rest {
	client := GetClient()
	if client == nil {
		return nil
	}
	return (*client).Rest()
}

// - Client

func InitDiscordClient(token string, caches []cache.Flags, intents []gateway.Intents) *bot.Client {
	return discord.InitClient(token, caches, intents)
}

func GetDiscordClient() *bot.Client {
	return discord.Client
}

// - ShardedClient

func InitDiscordShardedClient(token string, shardCount int, caches []cache.Flags, intents []gateway.Intents) *bot.Client {
	return discord.InitShardedClient(token, shardCount, caches, intents)
}

func GetDiscordShardedClient() *bot.Client {
	return discord.ShardedClient
}

// - Paginator

func InitPaginator(client *bot.Client) *paginator.Manager {
	p := discord.InitPaginator()
	(*client).AddEventListeners(p)
	return p
}

func GetPaginator() *paginator.Manager {
	return discord.Paginator
}

// - Router

func InitRouter() *handler.Mux {
	return discord.InitRouter()
}

func GetRouter() *handler.Mux {
	return discord.Router
}

// EMAIL

func InitEmail(ctx context.Context) *sesv2.Client {
	return email.Init(ctx)
}

func GetEmailClient() *sesv2.Client {
	return email.Client
}

// LOGGER

func InitLogger(level slog.Level, format string, addSource bool) *slog.Logger {
	return logger.Init(level, format, addSource)
}

func GetLogger() *slog.Logger {
	return logger.Logger
}

// OAUTH

func InitOAuth(applicationID snowflake.ID, clientSecret, baseURL string) *oauth2.Client {
	return oauth.Init(applicationID, clientSecret, baseURL)
}

func GetOAuthClient() *oauth2.Client {
	return oauth.Client
}

func StartOAuth(ctx context.Context) {
	oauth.Start(ctx)
}

// SQL

func InitSQL(ctx context.Context, pgURL string) *pgxpool.Pool {
	pool, err := sql.Init(ctx, pgURL)
	if err != nil {
		slog.Error("Failed to initialize SQL pool", slog.Any("error", err))
		panic(err)
	}

	return pool
}

func GetPool() *pgxpool.Pool {
	return sql.PgPool
}

// TASKS

func InitCron() *cron.Cron {
	return cronjobs.InitCron()
}

func GetCron() *cron.Cron {
	return cronjobs.Cron
}

func StartCron(ctx context.Context) {
	cronjobs.StartCron(ctx)
}
