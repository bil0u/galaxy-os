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
make run/hue          # Build and run Hue
make run/kevin        # Build and run Kevin
make build/hue        # Build only
make test             # go test -v -race -buildvcs ./...
make audit            # Tests + vet + govulncheck + deadcode
make tidy             # go mod tidy + go fmt
```

The `bot` name is injected at compile time via `-ldflags` into `cmd.bot`. It determines which config section to read (`[bot.<name>]`) and which feature set to activate (defined in `cmd/bot.go`'s `bots` map). It is not a runtime argument.

---

## Architecture

### Package structure

```
internal/
  platform/    # Framework contracts — interfaces, boot pipeline, lifecycle
  config/      # Config implementation (FileStore, Resolver, Provider, GuildManager)
  feature/     # Flat feature catalog — one file per feature, scope as metadata
  service/     # Shared service implementations (cron, oauth, email, sql, logger)
  discord/     # Discord adapters (RestAdapter, LocaleResolver)
  locale/      # Internationalization utilities
```

### Dependency rules

- `platform/` imports only stdlib + disgo types. Never imports feature/, service/, config/
- `feature/` imports only `platform/` (interfaces). Never imports config/, service/, or other features
- `service/` imports only `platform/` (interfaces). Never imports feature/ or config/
- `config/` imports only `platform/` (interfaces)
- `cmd/bot.go` is the sole composition root — imports everything

### Boot pipeline

All initialization runs through a declarative stage pipeline in `cmd/bot.go` via `platform.Run()`:

1. **config** — loads `config.toml`, parses Global/Bot/Log
2. **logger** — initializes structured slog
3. **discord** — creates disgo bot.Client, sets up handler router
4. **guilds** — queries Discord API, loads per-guild config, initializes GuildManager
5. **services** — demand-driven activation: scans all feature `Needs()`, starts only required services
6. **features** — builds per-feature `Deps`, calls `Setup(deps)` then `Start(ctx)` per scope
7. **gateway** — opens Discord gateway connection
8. **ready** — waits for SIGINT/SIGTERM, triggers graceful shutdown

Shutdown is reverse order: features Stop, services StopAll, gateway Close.

### Feature system

Each feature is a struct implementing `platform.Feature`:

```go
type Feature interface {
    Name() string
    Scope() Scope                    // BotScope | GuildScope | CrossGuildScope
    Needs() []ServiceID              // demand-driven service activation
    Setup(deps Deps) error           // wire dependencies
    Start(ctx context.Context) error // begin active work
    Stop(ctx context.Context) error  // graceful teardown
}
```

Features receive dependencies via `platform.Deps` — a struct built per-feature by the framework:

- Always populated: Logger (scoped), Rest (interface), Commands (Registrar), Locale, Configs (ConfigProvider), Bus, Env, BotName
- Scope-dependent: GuildID (GuildScope), Guilds (CrossGuildScope)
- Needs-dependent: Cron (only if CronService declared), OAuth (only if OAuthService declared)

**Adding a new feature:**

1. Create one file in `internal/feature/`
2. Define a struct implementing `platform.Feature` with a config struct implementing `platform.Config`
3. Export a package-level var: `var MyFeature = &myFeature{}`
4. Register in `cmd/bot.go`'s `bots` map for the relevant bot(s)

### Configuration system

TOML-based, managed by viper. Two tiers:

- `config.toml` — global + per-bot config under `[bot.<botname>]`, features under `[features.<botname>.<featurekey>]`
- `config.<guild_snowflake_id>.toml` — per-guild overrides with inheritance from bot defaults

Both are **gitignored**. Only `config.example.toml` is tracked. Config access in features via `platform.ConfigProvider`:

```go
platform.ResolveGuild[MyConfig](deps.Configs, guildID)  // guild-scoped
platform.ResolveBot[MyConfig](deps.Configs)              // bot-scoped
```

### Services layer

Each service implements `platform.Service` (Name/Start/Health/Stop). Services are demand-activated — they start only when a feature declares them in `Needs()`. Current services: CronService, OAuthService, SQLService (deferred), EmailService (deferred).

### Localization

`internal/locale` provides `locale.Text` (`map[discord.Locale]string`) for Discord-facing localized strings. Always provide at least `discord.LocaleEnglishUS`.
