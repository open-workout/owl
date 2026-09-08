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

## Package layout

```
reference/
├── go.mod                    module github.com/open-workout/owl/reference
├── owl.go                    public API: Compile, Resolve, Progress
├── types.go                  canonical-form types, re-exported from internal/ir as aliases
├── compile_test.go           runs Compile against every conformance/parse/* fixture
├── internal/
│   ├── ir/                    canonical-form wire types + JSON codec (leaf package —
│   │                           see types.go's doc comment for why it's not in package owl)
│   ├── lexer/                  tokenizer for spec/grammar.ebnf
│   ├── parser/                 recursive-descent parser → AST
│   └── compiler/                AST → canonical-form (desugaring: rounds,
│                                 name*N, exercise/superset/emom/amrap/for_time,
│                                 conditionals, single-owner-rule validation)
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

`Compile` is implemented: lexer → parser → compiler, wired in `owl.go`.
It passes every `conformance/parse/*` fixture (`compile_test.go`, run as
part of `go test ./...`) and compiles all of `conformance/programs/*.owl`
without error. `Resolve` and `Progress` are still stubs (`ErrNotImplemented`)
— that's the next chunk of work, a separate per-athlete tree-walk over
the `*Program` `Compile` already produces, not more parsing.
`ParseCanonicalJSON` (decode already-compiled JSON into these types)
works, and is what `compile_test.go` uses to turn a fixture's
`expected.json` into a comparable Go value.

`cmd/owlc` has three real subcommands: `parse` (run `Compile` on a
`.owl` file and print its canonical form), `validate` (round-trip a
canonical-form JSON file through the Go types), and `validate-fixtures`
(walk `../conformance/{parse,resolve,progress}/` and check every fixture
against its JSON Schema — structural only, not correctness; `go test`
is what actually exercises `Compile` against the fixtures now).
