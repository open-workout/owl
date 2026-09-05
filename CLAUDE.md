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
├── types.go, codec.go  Go structs mirroring canonical-form.schema.json, JSON codec
├── internal/
│   ├── lexer/           tokenizer for spec/grammar.ebnf        (not written yet)
│   ├── parser/          recursive-descent parser → AST         (not written yet)
│   └── compiler/        AST → canonical form (desugaring)      (not written yet)
└── cmd/owlc/            CLI: parse | validate | validate-fixtures
```

**Current status**: `types.go`/`codec.go` and `ParseCanonicalJSON` work.
`Compile`, `Resolve`, `Progress` in `owl.go` are stubs that return
`ErrNotImplemented`. The lexer/parser/compiler is the next real chunk of
work — that's where new parser/compiler code belongs, wired up through
`Compile` in `owl.go` as the only public entry point.

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
```

This is exactly what CI (`.github/workflows/ci.yml`) runs. Note
`validate-fixtures` only checks fixtures are well-formed against the
schemas — it does not yet run a real `Compile`/`Resolve`/`Progress` and
diff against `expected.json`, because those aren't implemented. Once they
are, that CI step should be replaced (see the TODO in `ci.yml`) with one
that actually exercises `conformance/{parse,resolve,progress}/` end to
end — that's a natural signal for "the parser/compiler is far enough
along to wire into CI."

## Conformance fixtures

`conformance/parse/<case>/` = `source.owl` + (`expected.json` canonical
form, or `expected-error.json`). `conformance/resolve/` and
`conformance/progress/` similarly pair inputs with an expected output.
When the parser/compiler lands, these are the fixtures it must satisfy —
see `conformance/README.md` for the exact directory contract per kind.

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
