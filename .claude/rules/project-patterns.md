## Project Patterns

- Logging: `log/slog` with structured fields (`slog.Any`, `slog.String`)
- Localization: `locale.Text` (`map[discord.Locale]string`) — always provide at least `LocaleEnglishUS`
- Error accumulation: `var errs []error` + `append` + `errors.Join(errs...)` at return
- Struct tags: `mapstructure` for viper unmarshaling in config structs
- Config types: `config.Bot`, `config.Global`, `config.Log`, `config.Guild` — vars use `Cfg` suffix (`config.BotCfg`, `config.GlobalCfg`, etc.)
- Feature types: `features.Config` (interface), `features.Set`, `features.Manager` — manager var is unexported, use package-level functions
- Services facade: `services.Client()`, `services.RestClient()`, `services.Router()`, etc. — no `Get` prefix. Sub-package vars are unexported, accessed only through the facade. Feature `Setup()` receives a `SetupDeps` struct — use deps, not facade calls, for setup-time wiring
- Discord client: stored as `bot.Client` (interface), never `*bot.Client` (pointer-to-interface). Same for `oauth2.Client`
- Feature type: always declared explicitly with `features.WithType(...)`, never auto-detected from package path
- Discord IDs: always `snowflake.ID`, never raw `string` or `uint64`
- Functional options pattern for constructors with optional config (see `featureOption`, `ManagerOption`)
- Unexported by default — only export what's part of the package's public API
