## Go Style

### Naming

- No `Get` prefix on methods — use `Counts()` not `GetCounts()`. Use `Fetch`/`Compute` to signal expensive operations.
- Variable name length proportional to scope — single letters for loop indices, descriptive for package-level
- No package-name stuttering (`config.Bot` not `config.BotConfig`)
- Receiver names: short, consistent, first letter(s) of type — never `this`/`self`
- Don't shadow standard library package names with variables
- Avoid generic package names (`util`, `common`, `helper`)

### Formatting

- No fixed line length, but prefer refactoring over splitting
- Don't split long strings (URLs, paths) across lines
- Use literal field names in struct literals from external packages
- Prefer `nil` over empty slice for declarations (`var s []T` not `s := []T{}`)

### Imports

- Grouping: stdlib → external → internal, separated by blank lines
- Never use dot imports (`import .`)
- Blank imports only in `main` or tests

### Documentation

- Exported symbols get doc comments starting with the symbol name: `// Manager handles...`
- Comments explain WHY, not WHAT
- Don't state the obvious — omit docs that just restate the signature
- Document cleanup requirements explicitly (close, cancel, release)
