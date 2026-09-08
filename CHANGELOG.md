# Changelog

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
See [`spec/versioning.md`](./spec/versioning.md) for what counts as a
patch/minor/major change once the spec reaches `1.0`; everything below
`1.0` is pre-release and any entry may be breaking.

## [Unreleased]

### Added
- Initial formal grammar (`spec/grammar.ebnf`) covering `state`/`stats`,
  `units`, `plates`, `block`/`day`/`exercise` declarations, `set` lines,
  group expressions (`superset`/`circuit`, `emom`, `amrap`, `for_time`,
  `rounds`), inline catalog calls, and the fallback sentinel (`| AMW`).
- Canonical form (`spec/canonical-form.md`,
  `spec/canonical-form.schema.json`): the `Group` 4-knob IR every
  group-like construct compiles to, `ExerciseDecl`/`NamedGroup`
  declarations with `Ref`-based invocation, and the `Target`/`Load`
  quantity unions.
- Semantics (`spec/semantics/`): `groups.md` (desugaring rules),
  `targets-loads.md`, `parameters.md` (state namespace, single-owner
  rule, onboarding sentinels), `resolution.md` (`resolve`),
  `progression.md` (`progress`).
- Standard library (`spec/stdlib/schemes.md`): the `double` progression
  scheme, formally defined.
- Conformance corpus (`conformance/`): fixture-format schemas
  (`conformance/schema/`), one hand-verified fixture each for
  parse/resolve/progress, and three full real programs under
  `conformance/programs/` (`squat-everyday`, `superset-example`,
  `complicated-crossfit`).
- `if`/`then`/`else` conditionals (`spec/grammar.ebnf`'s `cond*Item`
  family) at the `exerciseItem` (sets/progress), `dayItem` (exercises),
  `blockItem` (days), and `topLevelItem` (blocks) levels, backed by four
  `conformance/programs/` examples (`if-then-else`,
  `if-then-else-for-progress`, `if-then-else-for-days`,
  `if-then-else-for-blocks`). Compiles to a new `Conditional<T>` IR node
  (`spec/semantics/conditionals.md`, `canonical-form.schema.json`'s
  `conditionalMember`/`conditionalProgression`) whose branch is picked by
  `resolve`, not structural compilation, since `cond` reads `state` —
  see `conformance/parse/conditional-exercise-choice/`,
  `conformance/parse/conditional-progression/`, and
  `conformance/resolve/conditional-progress-else-branch/`.
- `relOp` now covers `<=`, `>=`, and `==` in addition to `<`/`>`, backed
  by `conformance/parse/conditional-relops/`.
- `repeatStmt` now accepts an inline, anonymous statement list —
  `(A, B, …)*N` — in addition to a bare declared name, so a repeated
  block doesn't need a name that's never reused elsewhere (`grammar.ebnf`'s
  new `repeatable` production). Desugars identically to the existing
  `name*N` rule (`spec/semantics/groups.md` §3.7a) — no canonical-form
  change. Backed by `conformance/parse/anonymous-group-repeat/`.
- `duration` now also accepts `min` for minutes (`rest(2min)`), matching
  the spelling `unit` already offered — previously only bare `m` worked
  in a `duration` slot.

### Open questions (tracked in the spec, not yet resolved)
- `dropset`/`restpause` surface syntax (`spec/semantics/groups.md` §5).
- Explicit inter-round rest syntax for `superset`/`circuit`
  (`spec/semantics/groups.md` §5).
- Whether `state` and `stats` should stay synonyms or become distinct
  concepts (`spec/semantics/parameters.md` §1).
- `!=`, a `setDecl` conditional branch, `else`-less conditionals, and
  `else if` chaining (`spec/semantics/conditionals.md` §4).
