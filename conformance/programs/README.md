# conformance/programs/

Full, real workout programs, used as parse-only conformance fixtures (see
[`../README.md`](../README.md)): a reference implementation must parse
every file here to a valid [canonical form](../../spec/canonical-form.md)
without error. None of these currently ships a golden `expected.json` —
hand-computing the full canonical output for a program this size is
significant, error-prone manual work; see the small hand-verified fixtures
under [`../parse/`](../parse/), [`../resolve/`](../resolve/), and
[`../progress/`](../progress/) for the format golden output would follow.
Pairing these with golden output is tracked as follow-up work.

| Program | Demonstrates |
|---|---|
| [`squat-everyday.owl`](./squat-everyday.owl) | Multi-day block repeated with `block*N`, `state`/training-max formulas with catalog-field references and both fallback shapes (numeric and `AMW` sentinel), a `double`-style `progress` code block (floor gate, log-value copy-down, ceiling check with compound assignment — see `spec/semantics/progression.md` §2) reimplemented across several exercises. |
| [`superset-example.owl`](./superset-example.owl) | `superset(...)`, including the "accepts any number of params" case (3rd exercise run straight, not part of the superset), a distance-only set (`CardioRower`, no `@ load`). |
| [`complicated-crossfit.owl`](./complicated-crossfit.owl) | All three CrossFit-style group constructs in one day: `N*emom(t){...}`, `for_time { rounds [...] as n {...} }` ("Fran"), `amrap(t){...}` — plus `rest(...)` sequencing between them. |
| [`if-then-else.owl`](./if-then-else.owl) | `if`/`then`/`else` at the `dayItem` level, choosing between a `superset(...)` and a bare exercise ref based on a `state` comparison. |
| [`if-then-else-for-progress.owl`](./if-then-else-for-progress.owl) | `progress`'s own `if`/`then`/`else` (a `progressionIf`, not a structural `Conditional`): a single `progress` block branches on the current `tm.squat.weight` to pick a different weight increment. (An `exerciseItem`-level structural conditional choosing between two whole `progress` blocks was the old shape here — that's gone along with the `double` scheme call; `condExerciseItem` still parses but the compiler rejects it, see `spec/semantics/conditionals.md` §2's note.) |
| [`if-then-else-for-days.owl`](./if-then-else-for-days.owl) | `if`/`then`/`else` at the `blockItem` level, choosing which of two declared `day`s runs. |
| [`if-then-else-for-blocks.owl`](./if-then-else-for-blocks.owl) | `if`/`then`/`else` at the `topLevelItem` level, choosing which of two declared `block`s runs. |
| [`multiple-blocks.owl`](./multiple-blocks.owl) | `squat-everyday.owl`'s program plus a second, unrelated `block pause`, both sequenced at the top level with the anonymous-body form of the repeat macro — `(leader*5, pause)*5` — nesting a `name*N` repeat (`leader*5`) inside a parenthesized statement list (`spec/semantics/groups.md` §3.7a); also `(cardio_rower, rest(2min))*5` and `(cardio, rest(1d))*3` for the same construct with `rest(...)` as a list member. |
