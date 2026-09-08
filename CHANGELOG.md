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
- Conformance corpus (`conformance/`): fixture-format schemas
  (`conformance/schema/`), one hand-verified fixture each for
  parse/resolve/progress, and three full real programs under
  `conformance/programs/` (`squat-everyday`, `superset-example`,
  `complicated-crossfit`).
- `if`/`then`/`else` conditionals (`spec/grammar.ebnf`'s `cond*Item`
  family) at the `dayItem` (exercises), `blockItem` (days), and
  `topLevelItem` (blocks) levels, backed by three `conformance/programs/`
  examples (`if-then-else`, `if-then-else-for-days`,
  `if-then-else-for-blocks`). Compiles to a new `Conditional<Member>` IR
  node (`spec/semantics/conditionals.md`, `canonical-form.schema.json`'s
  `conditionalMember`) whose branch is picked by `resolve`, not
  structural compilation, since `cond` reads `state` — see
  `conformance/parse/conditional-exercise-choice/` and
  `conformance/parse/conditional-relops/`. (`condExerciseItem` parses,
  for symmetry, but the compiler rejects it — see the progression
  redesign below for why.)
- `relOp` now covers `<=`, `>=`, `==`, and `!=` in addition to `<`/`>`,
  backed by `conformance/parse/conditional-relops/`. `condExpr` also now
  supports `and`/`or` combining two conditions (`and` binds tighter than
  `or`; parenthesize to mix), shared by both `Conditional.cond` and the
  new `progressionIf.cond` below — not yet backed by a fixture combining
  them.
- `repeatStmt` now accepts an inline, anonymous statement list —
  `(A, B, …)*N` — in addition to a bare declared name, so a repeated
  block doesn't need a name that's never reused elsewhere (`grammar.ebnf`'s
  new `repeatable` production). Desugars identically to the existing
  `name*N` rule (`spec/semantics/groups.md` §3.7a) — no canonical-form
  change. Backed by `conformance/parse/anonymous-group-repeat/`.
- `duration` now also accepts `min` for minutes (`rest(2min)`), matching
  the spelling `unit` already offered — previously only bare `m` worked
  in a `duration` slot.

### Changed
- **`progress` is now code, not a call into a closed set of stdlib
  schemes — `double` is gone.** Previously `progress = double(top_set,
  8, 12, 5)` named a set label and positional tuning numbers, resolved
  against a fixed algorithm defined in `spec/stdlib/schemes.md` (now
  deleted). Now `progress = { ... }` is a small code block — at most one
  per exercise, not per set — of `state`-path assignments guarded by its
  own `if`/`then`/`else` (grammar's new `progressBody`/`progressAssign`/
  `progressIf`; `canonical-form.schema.json`'s new `progressionBody`/
  `assign`/`progressionIf` defs, replacing `progressionRule` and
  `conditionalProgression` entirely). `ExerciseDecl` gained an optional
  `progress` field; `SetRef` lost its `progression` field. See
  `spec/semantics/progression.md`.
- `expr` gained binary `+`, `-` (new precedence level above `term`) and
  `term` gained `/`, alongside the existing `*` — needed to write
  `tm.squat.weight = tm.squat.weight + 5` directly, now that update
  rules are ordinary code rather than a scheme's positional args.
  `canonical-form.schema.json`'s `expr` def gained matching `add`/`sub`/
  `div` variants.
- Inside a `progress` block specifically, `<label>.<field>`
  (`top_set.reps`, or dropset-qualified `ds.top.weight`) reads what was
  *actually logged* for that set this session, compiling to a new
  `LogExpr` (`expr`'s `"type": "log"` variant) — a different meaning
  from the *same spelling* everywhere else in the language (a
  `set` line's own target/load local reference, which stays the
  prescribed-formula substitution it always was; see
  `spec/semantics/targets-loads.md` §4's note). The single-owner rule
  (`spec/semantics/parameters.md` §2) is now checked directly against
  `progress` blocks' `assign` targets rather than inferred from a
  `SetRef`'s target/load expression.

### Open questions (tracked in the spec, not yet resolved)
- `dropset`/`restpause` surface syntax (`spec/semantics/groups.md` §5).
- Explicit inter-round rest syntax for `superset`/`circuit`
  (`spec/semantics/groups.md` §5).
- Whether `state` and `stats` should stay synonyms or become distinct
  concepts (`spec/semantics/parameters.md` §1).
- A `setDecl` conditional branch, `else`-less conditionals, and
  `else if` chaining for `Conditional<Member>`'s four levels
  (`spec/semantics/conditionals.md` §4) — `progressionIf` already has
  both an optional `else` and attested `else if` chaining by
  construction, so these only remain open for the older
  `dayItem`/`blockItem`/`topLevelItem`/`exerciseItem` family.
