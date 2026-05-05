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

func InitDiscordClient(token string, caches []cache.Flags, intents []gateway.Intents) (bot.Client, error) {
	return discord.InitClient(token, caches, intents)
}

func Client() bot.Client {
	return discord.Client()
}

func RestClient() rest.Rest {
	client := discord.Client()
	if client == nil {
		return nil
	}
	return client.Rest()
}

// - Paginator

func InitPaginator(client bot.Client) *paginator.Manager {
	p := discord.InitPaginator()
	client.AddEventListeners(p)
	return p
}

func Paginator() *paginator.Manager {
	return discord.Paginator()
}

// - Router

func InitRouter() *handler.Mux {
	return discord.InitRouter()
}

func Router() *handler.Mux {
	return discord.Router()
}

// EMAIL

func InitEmail(ctx context.Context) (*sesv2.Client, error) {
	return email.Init(ctx)
}

func EmailClient() *sesv2.Client {
	return email.Client()
}

// LOGGER

func InitLogger(level slog.Level, format string, addSource bool) (*slog.Logger, error) {
	return logger.Init(level, format, addSource)
}

func Logger() *slog.Logger {
	return logger.Logger()
}

// OAUTH

func InitOAuth(applicationID snowflake.ID, clientSecret, baseURL string) oauth2.Client {
	return oauth.Init(applicationID, clientSecret, baseURL)
}

func OAuthClient() oauth2.Client {
	return oauth.Client()
}

func StartOAuth() {
	oauth.Start()
}

// SQL

func InitSQL(ctx context.Context, pgURL string) (*pgxpool.Pool, error) {
	return sql.Init(ctx, pgURL)
}

func Pool() *pgxpool.Pool {
	return sql.Pool()
}

// TASKS

func InitCron() *cron.Cron {
	return cronjobs.InitCron()
}

func Cron() *cron.Cron {
	return cronjobs.Scheduler()
}

func StartCron() {
	cronjobs.StartCron()
}

func StopCron() {
	cronjobs.StopCron()
}
