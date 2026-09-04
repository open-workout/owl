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

### Open questions (tracked in the spec, not yet resolved)
- `dropset`/`restpause` surface syntax (`spec/semantics/groups.md` §5).
- Explicit inter-round rest syntax for `superset`/`circuit`
  (`spec/semantics/groups.md` §5).
- Whether `state` and `stats` should stay synonyms or become distinct
  concepts (`spec/semantics/parameters.md` §1).
