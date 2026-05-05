## Go Patterns

### Error handling

- Wrap errors with context: `fmt.Errorf("doing X: %w", err)`
- Place `%w` at the end of format string so printed output mirrors the chain
- Use `%w` only when callers need to inspect the chain — at system boundaries, translate to domain errors
- Error strings: lowercase, no ending punctuation (`"something failed"` not `"Something failed."`)
- Never discard errors with `_` — handle, wrap, or explicitly document why it's safe to ignore
- Use `errors.Is` / `errors.As` for programmatic error checking, not string matching
- Sentinel errors (`var ErrNotFound = errors.New(...)`) for errors callers need to distinguish
- Handle error first, don't nest happy path in `else` — early returns
- `panic` only at startup for unrecoverable init failures — everywhere else, return errors

### Concurrency

- Protect shared mutable state — mutex for state, channels for communication
- Annotate channel direction: `chan<-` (send-only), `<-chan` (receive-only)
- Every goroutine must have a clear exit path (context cancellation, done channel, or bounded work)
- Never capture loop variables in goroutine closures without rebinding
- Don't store `context.Context` in structs — pass explicitly to each method that needs it
- Pass `context.Context` as first param

### Testing

- Table-driven tests with subtests (`t.Run`)
- `t.Fatal` only for setup failures — within table loops use `t.Error` + `continue`
- Don't call `t.Fatal` from goroutines — only the test goroutine may call fatal functions
- Don't use assertion libraries — use `cmp.Diff` for complex comparisons
- Size hints (preallocating slices/maps) only with empirical data, not guesses
