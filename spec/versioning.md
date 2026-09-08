# Versioning

OWL has two things that version independently: the **surface language**
(`grammar.ebnf` + the semantics that give it meaning) and the
**canonical form** (the JSON interchange shape a compiled program is
serialized as). A source `.owl` file is versionless — it's plain text, run
through whatever compiler the author has; the canonical form it compiles
to is what other tools (the reference implementation, the mobile app, the
playground) actually need to agree on a version for.

## Canonical form version field

Every canonical-form document carries a top-level `owlVersion` field
(SemVer string, e.g. `"0.1.0"`) declaring which version of
[`canonical-form.schema.json`](./canonical-form.schema.json) it was
produced against. A consumer (the app, a conformance runner) must check
this before trusting the rest of the document's shape.

## Policy

While OWL is pre-1.0 (all of `spec/` today), any change to the grammar or
the canonical form is a minor-version bump and may be breaking — there is
no compatibility guarantee yet between `0.x` versions. Once the spec
reaches `1.0`:

- **Patch**: clarifications that don't change what any existing `.owl`
  program compiles to (e.g. fixing an ambiguous grammar comment).
- **Minor**: additive, backward-compatible changes — a new group
  construct, a new `Expr`/`BoolExpr` operator, a new `Target`/`Load`
  variant — that don't change the compiled output of any program valid
  under the previous minor version.
- **Major**: anything that changes what an existing valid `.owl` program
  compiles to, or removes/renames a canonical-form field.

## Conformance corpus versioning

Each fixture under `conformance/` (see [`conformance/README.md`](../conformance/README.md))
is pinned to the spec version it was written against, so a reference
implementation targeting an older spec version isn't expected to pass
fixtures written for a newer one.
