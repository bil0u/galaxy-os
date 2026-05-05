# CLAUDE.md

This file provides guidance to AI agents when working with code in this repository.

---

## Project

Reusable Discord bot platform designed to deploy multiple bots across multiple servers. The architecture provides a shared base layer for common concerns (config, services, features) so that each bot is a thin composition of features on top of the platform.

Currently deployed on the Galaxy One server with two bots: `hue` and `kevin`.

---

## Tech Stack

Go project — see `go.mod` for current versions of all dependencies.

| Library                 | Role                                                                 |
| ----------------------- | -------------------------------------------------------------------- |
| `disgoorg/disgo`        | Discord framework (gateway, REST, sharding, OAuth2, command handler) |
| `spf13/viper`           | Config management with Sub-tree access for nested feature configs    |
| `spf13/cobra`           | CLI structure (`bot start`, `generate`)                              |
| `robfig/cron/v3`        | Cron scheduling with timezone support                                |
| `disgoorg/snowflake/v2` | Discord snowflake ID type                                            |
| `jackc/pgx/v5`          | PostgreSQL connection pool                                           |
| `aws/aws-sdk-go-v2`     | Email via SES                                                        |

---

## Build and Run

```sh
make run/hue          # Build and run Hue (with --cron --oauth2)
make run/kevin        # Build and run Kevin
make build/hue        # Build only
make test             # go test -v -race -buildvcs ./...
make audit            # Tests + vet + govulncheck + deadcode
make tidy             # go mod tidy + go fmt
```

The `bot` name is injected at compile time via `-ldflags` into `cmd.bot`. It determines which config section to read (`[bot.<name>]`) and which feature set to activate (defined in `cmd/bot.go`'s `botFeatures` map). It is not a runtime argument.

---

## Architecture

### Initialization order

All initialization happens in `cmd/bot.go`'s `startBot()`. No package-level singletons — each resource is a local variable passed through `SetupDeps`.

1. `config.Init(bot)` — reads `config.toml`, returns `(*Global, *Log, *Bot, error)`
2. `services.InitLogger(...)` — structured slog setup
3. `disgo.New(token, ...)` — creates per-bot Discord client directly (no singleton)
4. `config.InitGuilds(ctx, client.Rest, botName)` — queries Discord API for guilds, reads per-guild config files
5. `handler.New()` / `services.InitCron()` / `services.InitOAuth(...)` — per-bot router, shared cron/oauth
6. `features.NewRegistry(...)` + `features.SetupFeatures(fs, deps)` — builds config registry, calls each feature's `Setup(deps)`

### Configuration system

TOML-based, managed by viper. Two tiers:

- `config.toml` — global + per-bot config under `[bot.<botname>]`, bot-level features under `[features.<botname>.<featurekey>]`
- `config.<guild_snowflake_id>.toml` — per-guild settings, guild-level features under `[features.<botname>.<featurekey>]`

Both are **gitignored**. Only `config.example.toml` is tracked. Feature configs are accessed at runtime via the `FeatureRegistry` (passed through `deps.Configs.Features`):

```go
features.GetConfigFrom[MyConfig](registry, guildID)   // guild feature
features.GetConfigFrom[MyConfig](registry, 0)         // bot feature (guildID = 0)
```

### Feature system

Each feature has:

- A config struct implementing `features.Config` (must have `Validate() error`)
- A `Setup(deps features.SetupDeps) error` function that receives its dependencies explicitly
- A package-level var created with `features.New[ConfigType](setupFn, ...opts)`

Feature type must be declared explicitly with `features.WithType(features.BotFeature)` or `features.WithType(features.GuildFeature)`. The key is auto-derived from the config struct name (`DailyMessageConfig` → `daily_message`).

`SetupDeps` is grouped into three semantic sections:

- `deps.Bot` — `BotServices{Client *bot.Client, Router *handler.Mux, Logger *slog.Logger}`
- `deps.Shared` — `SharedServices{Cron *cron.Cron}` (nil if `--cron` not set)
- `deps.Configs` — `Configs{Bot *config.Bot, Guilds *config.GuildMap, Global *config.Global, Features *FeatureRegistry}`

Features capture what they need from deps in closures — no globals, no facade calls at runtime.

**Adding a new feature:**

1. Create file in `internal/features/bot/` or `internal/features/guild/`
2. Define config struct implementing `features.Config`
3. Export `var XFeature = features.New[XConfig](setupFunc, features.WithType(...), ...opts)`
4. Register in `cmd/bot.go`'s `bots` map for the relevant bot(s)

### Services layer

`internal/services/services.go` is a thin facade over shared service sub-packages (`cron/`, `email/`, `oauth/`, `sql/`, `logger/`). Discord resources (client, router) are created directly in `cmd/bot.go` — no singleton wrapper.

Features receive all dependencies via `SetupDeps` and derive REST from `deps.Bot.Client.Rest`. No feature should import the services package.

### Localization

`internal/locale` provides `locale.Text` (`map[discord.Locale]string`) for Discord-facing localized strings. Always provide at least `discord.LocaleEnglishUS`.
