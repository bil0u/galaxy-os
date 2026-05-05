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

### Initialization order (strict — violating it causes nil panics)

Init functions are called once in a deterministic sequence from `cmd/bot.go`'s `startBot()`. Out-of-order calls cause nil-pointer panics.

1. `config.Init(bot)` — reads `config.toml`, populates `config.Global`, `config.Log`, `config.Bot`
2. `services.InitLogger(...)` — structured slog setup
3. `services.InitDiscordClient(...)` — single client for REST and gateway
4. `config.InitGuilds(ctx, restClient)` — queries Discord API for guilds, reads per-guild config files
5. `services.InitRouter()` / `InitPaginator(...)` / `InitCron()` / `InitOAuth(...)` — remaining services
6. `features.Init(featureSet, deps)` — creates FeatureManager, calls each feature's `Setup(deps)`

### Configuration system

TOML-based, managed by viper. Two tiers:

- `config.toml` — global + per-bot config under `[bot.<botname>]`, bot-level features under `[features.<botname>.<featurekey>]`
- `config.<guild_snowflake_id>.toml` — per-guild settings, guild-level features under `[features.<botname>.<featurekey>]`

Both are **gitignored**. Only `config.example.toml` is tracked. Feature configs are accessed at runtime via:

```go
features.GetConfig[MyConfig](guildID)   // guild feature
features.GetBotConfig[MyConfig]()       // bot feature (guildID = 0)
```

### Feature system

Each feature has:

- A config struct implementing `features.FeatureConfig` (must have `Validate() error`)
- A `Setup(deps features.SetupDeps) error` function that receives its dependencies explicitly
- A package-level var created with `features.New[ConfigType](setupFn, ...opts)`

Feature type must be declared explicitly with `features.WithType(features.BotFeature)` or `features.WithType(features.GuildFeature)`. The key is auto-derived from the config struct name (`DailyMessageConfig` → `daily_message`).

`SetupDeps` provides: `Client bot.Client`, `Router *handler.Mux`, `Cron *cron.Cron` (nil if `--cron` not set).

**Adding a new feature:**

1. Create file in `internal/features/bot/` or `internal/features/guild/`
2. Define config struct implementing `FeatureConfig`
3. Export `var XFeature = features.New[XConfig](setupFunc, features.WithType(...), ...opts)`
4. Register in `cmd/bot.go`'s `botFeatures` map for the relevant bot(s)

### Services layer

`internal/services/services.go` is the public facade over sub-packages (`discord/`, `cron/`, `email/`, `oauth/`, `sql/`, `logger/`). Sub-package vars are unexported — the facade is the only access path. Always use `services.Get*()` / `services.Init*()` — never import sub-packages directly from features.

One Discord client exists (`bot.Client`, not pointer-to-interface). Feature `Setup()` functions receive dependencies via `SetupDeps` instead of calling the facade; runtime event handlers may still use `services.GetRestClient()` for REST calls.
