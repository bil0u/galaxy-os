# Roadmap

## Completed

- [x] Audit of current repo state (bugs, dead code, error handling, architecture, naming)
- [x] Disgo v0.18 → v0.19 migration
- [x] Per-bot scoped logging — logger injected via `platform.Deps`, scoped per feature and guild
- [x] Config system rewrite — `ConfigStore`, `ScopedResolver`, `ConfigProvider` with bot/guild inheritance, no package-level vars
- [x] Architecture refactor (9 phases) — `platform.Feature` interface, declarative boot pipeline, demand-driven services, dependency injection via `Deps`
- [x] Post-refactor fixes — `Feature.Start()` wired, `Deps.Cron`/`OAuth` filtered by `Needs()`, `crypto/rand` in OAuth, dead code removal, CLAUDE.md updated
- [x] Structure and naming refactor — directory reorganization, `feature.go` convention, `.gen/` for build artifacts, secrets/config split

## Architecture backlog

### High priority — unblocks features

- [ ] Gateway client access in Deps — expose event listener registration and `bot.Client` in `core.Deps`. Blocks `selfassign`, `presence`, and `suspiciousinterview`
- [ ] Populate `CronScheduler` interface methods — currently empty, `dailymessage` type-asserts to `*service.CronService`
- [ ] Command sync — features register commands but they're not bulk-synced to Discord API. Blocked on features declaring their command definitions
- [ ] Global config access in Deps — `info` feature needs `version`/`commit` from `GlobalConfig`, no path currently

### Medium priority — framework quality

- [ ] Populate `OAuthProvider` interface methods — currently empty, no feature uses OAuth through Deps
- [ ] Type `ConfigBundle` fields — currently `any` to avoid import cycle between core and config
- [ ] Config caching + invalidation — `Resolver.Invalidate()` is a no-op, no cache layer exists yet
- [ ] Remove legacy `rest.Rest` imports from `selfassign` — dead utility functions using concrete REST
- [ ] Dynamic guild onboarding — subscribe to `GuildCreate`/`GuildDelete` gateway events for runtime guild management
- [ ] Per-method doc comments on core interfaces — spec has detailed comments, implementation has interface-level only
- [ ] Unit tests — zero test files across `core/`, `config/`, `guild/`, `service/`

### Low priority — cleanup

- [ ] `permissions` feature — complete stub, needs design and implementation
- [ ] Clean up stale plan files in `.claude/plans/`

## Infrastructure

- [ ] Switch config format from TOML to YAML — viper supports it natively, just needs filename/type changes
- [ ] Commit config files — secrets are split out, `conf/config*.toml` can now be committed after migrating tokens to `secrets.toml`
- [ ] Health HTTP endpoint — `HealthAggregator` and `Service.Health()` implemented, need ~20-line HTTP server for K8s probes
- [ ] OpenTelemetry integration (`oteldisgo`) — `Observability` struct ready with `any`-typed `Tracer`/`Meter` placeholders, plug in real providers when OTel collector is deployed
- [ ] Health checks for shard/gateway disconnections — detect and recover from silent disconnects
- [ ] Inter-bot `Bus` implementation — interface and `NoBus` default wired, channel-based impl ~50 lines when coordination is needed
- [ ] Multi-bot process — run multiple bots in one binary (currently one-bot-per-binary via ldflags)
- [ ] Disgo optimization — audit API usage patterns, reduce redundant REST calls
- [ ] Gateway intents optimization (per-bot minimum intents)
- [ ] Cache flags optimization (per-bot minimum cache)
- [ ] SQL schema management tooling
- [ ] i18n with string IDs — reference localized strings by namespaced IDs instead of inline maps, store translations in files, code provides ID + English default, other locales resolved at runtime with English fallback

## Ideas (to be challenged)

- [ ] Per-guild config subfolders — `conf/<guild_id>/` to store auxiliary config files per feature (e.g. `suspicious_interview.toml` for interview questions) instead of bloating the base config. Two options for how features access these:
  - **Feature-driven**: each feature knows its filename convention and loads its own files
  - **Framework API**: `ConfigProvider.LoadAuxiliary(guildID, filename, target)` resolves `conf/<guildID>/<filename>` — cleaner but couples the framework to the folder convention

## Planned features

- [ ] Prevent soundboard spam (guild feature)
- [ ] Hot-reload configuration — `ConfigStore.Watch` and `ConfigProvider.OnChange` interfaces ready (no-op stubs), implement file watcher when restart cost becomes painful
- [ ] Runtime feature enable/disable via Discord admin commands
- [ ] Per-guild feature gating — allowlist in guild config to restrict which features are available (supports paywalling features per guild)
- [ ] Code generation for scaffolding new features (if boilerplate grows)
