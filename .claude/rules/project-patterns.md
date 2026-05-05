## Project Patterns

- Logging: `log/slog` with structured fields (`slog.Any`, `slog.String`)
- Localization: `utils.LocalizedString` (`map[discord.Locale]string`) — always provide at least `LocaleEnglishUS`
- Error accumulation: `var errs []error` + `append` + `errors.Join(errs...)` at return
- Struct tags: `mapstructure` for viper unmarshaling in config structs
- Singletons: package-level `var` initialized once during deterministic boot sequence in `cmd/bot.go`
- Discord IDs: always `snowflake.ID`, never raw `string` or `uint64`
- Functional options pattern for constructors with optional config (see `featureOption`, `FeatureManagerOpts`)
- Unexported by default — only export what's part of the package's public API
