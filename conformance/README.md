# conformance/

An implementation-independent test corpus. Nothing here is OWL source for
its own sake (that's [`examples/`](../examples/)) or prose spec (that's
[`spec/`](../spec/)) — every file here is a fixture: a fixed input and a
fixed expected output, meant to be run against any implementation that
claims to compile/resolve/progress OWL.

## Layout

```
conformance/
├── schema/    JSON Schema for the fixture files themselves
├── parse/     source.owl        → expected.json  (canonical form) | expected.error.json
├── resolve/   canonical.json + state.json  → expected.json  (a Session)
├── progress/  canonical.json + log.json    → expected.json  (a State)
└── programs/  full real programs (no expected output paired yet — see programs/README.md)
```

## `parse/<case>/`

Each case is a directory: `source.owl` (the input) plus either
`expected.json` (the canonical form it must compile to — validates against
[`../spec/canonical-form.schema.json`](../spec/canonical-form.schema.json))
or `expected-error.json` (for cases that must be rejected — see
[`schema/parse-error.schema.json`](./schema/parse-error.schema.json)).
Exactly one of the two must be present.

## `resolve/<case>/`

Each case is a directory: `canonical.json` (a full canonical-form
document) + `state.json` (an athlete's current `state` values — see
[`schema/state.schema.json`](./schema/state.schema.json)) + `cursor.json`
(which block/day/iteration to resolve — see
[`schema/cursor.schema.json`](./schema/cursor.schema.json)), and
`expected.json`, the `Session` [`resolve`](../spec/semantics/resolution.md)
must produce from them.

## `progress/<case>/`

Each case is a directory: `canonical.json` + `state.json` (state *before*)
+ `log.json` (what was actually performed — see
[`schema/session-log.schema.json`](./schema/session-log.schema.json)), and
`expected.json`, the updated `state` [`progress`](../spec/semantics/progression.md)
must produce.

## `programs/`

Full, real programs — see [`programs/README.md`](./programs/README.md).
These are currently parse-only fixtures without a golden `expected.json`
(hand-computing the full canonical form for a program this size is
significant, error-prone manual work); a reference implementation should
still be able to parse every one of them without error. Pairing them with
golden output is tracked as follow-up work, not blocking.

## Versioning

Every fixture is implicitly pinned to the `spec/` version it was written
against — see [`../spec/versioning.md`](../spec/versioning.md).
