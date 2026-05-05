package services

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	cronjobs "github.com/bil0u/galaxy-os/internal/services/cron"
	"github.com/bil0u/galaxy-os/internal/services/email"
	"github.com/bil0u/galaxy-os/internal/services/logger"
	"github.com/bil0u/galaxy-os/internal/services/oauth"
	"github.com/bil0u/galaxy-os/internal/services/sql"
	"github.com/disgoorg/disgo/oauth2"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
)

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
