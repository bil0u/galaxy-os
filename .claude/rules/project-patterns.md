## Project Patterns

- Logging: `log/slog` with structured fields (`slog.Any`, `slog.String`). Per-bot scoped logger via `slog.Default().With("bot", name)`
- Localization: `locale.Text` (`map[discord.Locale]string`) — always provide at least `LocaleEnglishUS`
- Error accumulation: `var errs []error` + `append` + `errors.Join(errs...)` at return
- Struct tags: `mapstructure` for viper unmarshaling in config structs
- Config types: `config.Bot`, `config.Global`, `config.Log`, `config.Guild` — no package-level globals, `config.Init()` returns values
- Feature types: `features.Config` (interface), `features.Set`, `features.FeatureRegistry` (config resolution only)
- Dependency injection: `features.SetupDeps` groups `BotServices`, `SharedServices`, `Configs`. Features capture deps in closures — no globals or facade calls at runtime
- Config access in features: `features.GetConfigFrom[T](registry, guildID)` — registry comes from `deps.Configs.Features`
- Services facade: `services.InitLogger()`, `services.InitCron()`, `services.Cron()`, etc. — only shared services. Discord resources created directly in `cmd/bot.go`
- Discord client: stored as `bot.Client` (interface), never `*bot.Client` (pointer-to-interface). Same for `oauth2.Client`
- REST access: derive from `client.Rest()` — no separate singleton or field
- Feature type: always declared explicitly with `features.WithType(...)`, never auto-detected from package path
- Discord IDs: always `snowflake.ID`, never raw `string` or `uint64`
- Functional options pattern for constructors with optional config (see `featureOption`)
- Unexported by default — only export what's part of the package's public API
