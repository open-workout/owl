# bindings/

Per-language wrappers around [`../reference/`](../reference/).

**Likely unnecessary, given the chosen architecture.** Compilation and
resolution run server-side in Go (see `../reference/README.md`); the
phone only ever consumes plain JSON (canonical form / a resolved
`Session`), so it needs a JSON decoder, not an OWL binding, in whatever
language the mobile app is written in. The `js/` and `python/`
placeholders would only matter if something *other* than the Go server
needs to run `Compile`/`Resolve`/`Progress` itself — e.g. a browser
playground doing client-side compilation instead of calling the server.
Don't build these speculatively; add one only when a concrete consumer
needs it.
