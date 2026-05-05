## Security

- Never log secrets, tokens, or credentials — even at debug level
- Validate and sanitize at system boundaries (user input, external API responses, webhook payloads)
- Use `crypto/subtle.ConstantTimeCompare` for sensitive string comparisons
- `crypto/rand` for anything security-sensitive — never `math/rand`
