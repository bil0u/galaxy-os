package contracts

import (
	"context"
	"log/slog"

	"github.com/disgoorg/snowflake/v2"
)

// Scope declares a feature's relationship to guilds.
type Scope int

const (
	BotScope   Scope = iota // Framework calls Setup once. No guild context.
	GuildScope              // Framework calls Setup per guild.
	CrossGuildScope         // Framework calls Setup once with GuildAccessor in Deps.
)

// Feature is the unit of bot composition. Every feature implements this.
// The framework manages lifecycle; the feature manages behavior.
type Feature interface {
	Name() string
	Scope() Scope
	Needs() []ServiceID
	Setup(deps Deps) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Config is the constraint on feature config structs.
type Config interface {
	Validate() error
}

// Deps is built per-feature by the framework based on the feature's
// declared scope and service needs.
type Deps struct {
	Logger   *slog.Logger
	Rest     RestClient
	Commands Registrar
	Locale   LocaleResolver
	Configs  ConfigProvider
	Bus      Bus
	Env      Env
	BotName  string

	GuildID snowflake.ID
	Guilds  GuildAccessor

	Cron  CronScheduler
	OAuth OAuthProvider
}
