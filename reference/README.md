# reference/

The reference implementation of OWL, in **Go**: the parser, the
canonical-form compiler, `resolve`, and `progress` (see
[`../spec/`](../spec/) for what each of those means).

## Architecture

Compilation and resolution happen **server-side**, not on the phone. The
OpenWorkout backend (a separate service — this repo ships the language,
not the product's server) imports this package (or shells out to the
`owlc` CLI below) to:

1. parse a `.owl` program to canonical form ([`../spec/canonical-form.md`](../spec/canonical-form.md)),
2. `resolve` it against an athlete's `state` into a concrete `Session`,
3. serve that `Session` JSON to the phone — the QR code the app scans
   encodes however the server chooses to hand it off (a URL to fetch, a
   short-lived token, or the JSON itself if it's small enough).

The phone is a **pure JSON consumer**: it only needs to deserialize
against [`../spec/canonical-form.schema.json`](../spec/canonical-form.schema.json) /
[`../conformance/schema/session.schema.json`](../conformance/schema/session.schema.json)
and render it — it never runs an OWL parser itself. That's why
[`../bindings/`](../bindings/) and [`../tools/playground/`](../tools/playground/)
are lower priority than they'd be if the phone had to embed a compiler:
see those directories' READMEs.

## Package layout (planned)

```
reference/
├── go.mod                    module github.com/open-workout/owl/reference
├── owl.go                    public API: Compile, Resolve, Progress
├── internal/
│   ├── lexer/                 tokenizer for spec/grammar.ebnf
│   ├── parser/                recursive-descent parser → AST
│   └── compiler/               AST → canonical-form (desugaring: rounds,
│                                name*N, exercise/superset/emom/amrap/for_time)
└── cmd/
    └── owlc/                  CLI: `owlc parse|resolve|progress <file>`,
                                 used locally and by CI to run the
                                 conformance corpus for real (closing the
                                 TODO in .github/workflows/ci.yml)
```

A hand-written recursive-descent parser (not a parser-generator like
ANTLR/pigeon) is the plan — `spec/grammar.ebnf`'s productions are simple
enough not to need one, and it keeps the toolchain to just Go. That's why
[`../grammar/`](../grammar/) has no Go-specific grammar artifact in it.

## Status

`go.mod` and the Go struct types mirroring `canonical-form.schema.json`
exist (`types.go`, with the JSON codec for its sum-typed fields in
`codec.go`) — nothing parses OWL source yet. `owl.go`'s `Compile`,
`Resolve`, `Progress` are stubs; `ParseCanonicalJSON` (decode
already-compiled JSON into these types) works.

`cmd/owlc` has two real subcommands today: `validate` (round-trip a
canonical-form JSON file through the Go types) and `validate-fixtures`
(walk `../conformance/{parse,resolve,progress}/` and check every fixture
against its JSON Schema — this is the whole repo's schema-validation
tooling now; there's no separate Node/npm step). The lexer/parser/
compiler is the next real chunk of work.
