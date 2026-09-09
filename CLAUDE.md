# CLAUDE.md

Orientation for working in this repo. Each directory has its own README
with more detail — this file is the map, not a replacement for them.

## What this is

OWL (Open Workout Language) is a DSL for workout programs, compiled to a
canonical JSON IR and resolved against an athlete's `state` into a
concrete session. This repo ships the *language* — grammar, semantics,
reference compiler, conformance corpus — not the OpenWorkout product
itself.

`spec/` is the source of truth (prose + `grammar.ebnf`, not code).
Everything else either implements it (`reference/`) or tests an
implementation against it (`conformance/`).

## Architecture, in one paragraph

Compilation and resolution run **server-side in Go** (`reference/`), not
on the phone. The OpenWorkout backend imports `reference/` (or shells out
to `owlc`) to compile a `.owl` program to canonical form, `resolve` it
against an athlete's state into a `Session`, and hand that JSON to the
phone. The phone only ever deserializes JSON against
`spec/canonical-form.schema.json` / `conformance/schema/session.schema.json`
— it never runs a parser. This is why `bindings/` and `tools/playground/`
are explicitly low-priority placeholders (see their READMEs) rather than
built out: nothing needs them yet.

## Repository layout

| Path | What's there |
|---|---|
| `spec/` | Source of truth: `grammar.ebnf`, canonical form, semantics. Prose, not code. Read `spec/README.md`'s reading order before editing anything here. |
| `reference/` | The reference implementation, in Go. See below. |
| `conformance/` | Implementation-independent fixture corpus (`parse/`, `resolve/`, `progress/`, `programs/`). |
| `bindings/` | Per-language wrappers around `reference/`. Placeholders — don't build speculatively. |
| `examples/` | Real `.owl` programs, for reading. |
| `docs/` | User-facing tutorials. Not written yet. |
| `tools/playground/` | Web REPL. Not built yet, low priority. |
| `grammar/` | Deliberately empty — no parser-generator artifact, since `reference/` hand-writes its parser. |

## `reference/` — where the actual code lives

```
reference/
├── owl.go              public API: Compile, Resolve, Progress (+ ParseCanonicalJSON)
├── types.go             canonical-form types, re-exported from internal/ir as aliases
├── internal/
│   ├── ir/               canonical-form wire types + JSON codec (leaf package)
│   ├── lexer/            tokenizer for spec/grammar.ebnf
│   ├── parser/           recursive-descent parser → AST
│   └── compiler/         AST → canonical form (desugaring, conditionals,
│                          single-owner-rule validation)
└── cmd/owlc/            CLI: parse | validate | validate-fixtures
```

**Current status**: `Compile` is implemented (lexer → parser →
compiler) and passes every `conformance/parse/*` fixture plus all of
`conformance/programs/*.owl`, run via `go test ./...`. `Resolve` and
`Progress` in `owl.go` are still stubs that return `ErrNotImplemented`
— a separate per-athlete tree-walk over the already-compiled `*Program`,
not more parsing, and the next real chunk of work.

A hand-written recursive-descent parser is the deliberate choice over a
parser generator (ANTLR/pigeon/etc.) — `grammar.ebnf`'s productions don't
need one, and it keeps the toolchain to just Go.

## Building and testing

```sh
cd reference
gofmt -l .                        # must be clean
go vet ./...
go build ./...
go run ./cmd/owlc validate-fixtures ..   # validates conformance/ fixtures against JSON Schema
go test ./...                            # runs Compile against every conformance/parse/* fixture
```

This is exactly what CI (`.github/workflows/ci.yml`) runs.
`validate-fixtures` only checks fixtures are well-formed against the
schemas; `go test` is what actually runs `Compile` against them and
diffs the result against `expected.json` as Go values. `Resolve`/
`Progress` aren't implemented yet, so `conformance/{resolve,progress}/`
aren't exercised by `go test` yet either — extending it to cover them
is the natural next step once those two land.

## Conformance fixtures

`conformance/parse/<case>/` = `source.owl` + (`expected.json` canonical
form, or `expected-error.json`) — `reference/compile_test.go` runs
`Compile` against every case here. `conformance/resolve/` and
`conformance/progress/` similarly pair inputs with an expected output;
these aren't wired into `go test` yet since `Resolve`/`Progress` aren't
implemented. See `conformance/README.md` for the exact directory
contract per kind.

## Changing the language vs. changing an implementation

- Bug fixes to `reference/`, new `conformance/` fixtures for
  *already-spec'd* behavior, docs/examples changes: ordinary PRs.
- Anything touching `spec/grammar.ebnf`, `spec/semantics/*.md`, or
  `spec/canonical-form.schema.json` needs an OWL-EP (see
  `CONTRIBUTING.md`) — a spec change affects every implementation and
  every `.owl` program that already exists. Don't casually "fix" the spec
  while implementing the parser; if the grammar seems wrong or
  underspecified, that's a CONTRIBUTING.md proposal, not a drive-by edit.

## Conventions worth knowing

- Every group-like construct (straight sets, supersets, circuits, EMOMs,
  AMRAPs, for-time) compiles to **one** IR node with four knobs
  (`interleave`, `rest`, `termination`, `atomic`) — see
  `spec/semantics/groups.md`. A new construct should map onto these, not
  add a fifth knob.
- Fixtures are implicitly pinned to the `spec/` version they were written
  against — see `spec/versioning.md`.
