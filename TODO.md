# Roadmap

## Repository dust removal

- [ ] Audit of current repo state:
  - [x] Bugs and dead code
  - [x] Error handling
  - [x] Architecture
  - [ ] Naming
  - [ ] Disgo optimization
  - [ ] Multi-bot process
- [ ] Disgo v0.18 → v0.19 migration
- [ ] OpenTelemetry integration (`oteldisgo`)
- [ ] Per-bot scoped logging (`slog.With("bot", name)`)
- [ ] Health checks (detect shard/bot disconnections)
- [ ] SQL schema management tooling
- [ ] Gateway intents optimization (per-bot minimum intents)
- [ ] Cache flags optimization (per-bot minimum cache)
- [ ] Config container struct — replace package-level vars with a struct returned from `Init`, fix `InitGuilds` mutation-after-init smell

## Planned features

- [ ] Prevent soundboard spam (guild feature)
- [ ] Hot-reload configuration via `viper.WatchConfig()` → allow deployed bots to pick up config changes without restart
- [ ] Code generation for scaffolding new features (if boilerplate grows)
- [ ] Runtime feature enable/disable via Discord admin commands
- [ ] Per-guild feature gating → allowlist in guild config to restrict which features are available (supports paywalling features per guild)
