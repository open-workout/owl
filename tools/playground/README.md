# tools/playground/

A web REPL: paste an OWL program, see it parsed, compiled to canonical
form, and (given some `state`) resolved into a session.

**Lower priority given the chosen architecture.** The product doesn't
need a browser-side compiler — compilation happens server-side in Go (see
[`../../reference/README.md`](../../reference/README.md)) and the phone
just receives JSON. If built, this would most simply be a small static
page that calls a local instance of the `owlc` CLI's HTTP-server mode (or
a thin dev-only endpoint) rather than compiling in-browser via WASM — no
need for Go's `js/wasm` target unless client-side compilation becomes a
real requirement later.
