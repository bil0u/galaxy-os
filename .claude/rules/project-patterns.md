## Project Patterns

- Logging: `log/slog` with structured fields (`slog.Any`, `slog.String`)
- Localization: `utils.LocalizedString` (`map[discord.Locale]string`) — always provide at least `LocaleEnglishUS`
- Error accumulation: `var errs []error` + `append` + `errors.Join(errs...)` at return
- Struct tags: `mapstructure` for viper unmarshaling in config structs
- Services: sub-package vars are unexported, accessed only through the `services` facade. Feature `Setup()` receives a `SetupDeps` struct — use deps, not facade calls, for setup-time wiring
- Discord client: stored as `bot.Client` (interface), never `*bot.Client` (pointer-to-interface). Same for `oauth2.Client`
- Feature type: always declared explicitly with `features.WithType(...)`, never auto-detected from package path
- Discord IDs: always `snowflake.ID`, never raw `string` or `uint64`
- Functional options pattern for constructors with optional config (see `featureOption`, `FeatureManagerOpts`)
- Unexported by default — only export what's part of the package's public API
