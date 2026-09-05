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
| [`squat-everyday.owl`](./squat-everyday.owl) | Multi-day block repeated with `block*N`, `state`/training-max formulas with catalog-field references and both fallback shapes (numeric and `AMW` sentinel), `double` progression across several exercises. |
| [`superset-example.owl`](./superset-example.owl) | `superset(...)`, including the "accepts any number of params" case (3rd exercise run straight, not part of the superset), a distance-only set (`CardioRower`, no `@ load`). |
| [`complicated-crossfit.owl`](./complicated-crossfit.owl) | All three CrossFit-style group constructs in one day: `N*emom(t){...}`, `for_time { rounds [...] as n {...} }` ("Fran"), `amrap(t){...}` — plus `rest(...)` sequencing between them. |
| [`if-then-else.owl`](./if-then-else.owl) | `if`/`then`/`else` at the `dayItem` level, choosing between a `superset(...)` and a bare exercise ref based on a `state` comparison. |
| [`if-then-else-for-progress.owl`](./if-then-else-for-progress.owl) | `if`/`then`/`else` at the `exerciseItem` level, choosing between two `progress = double(...)` rules (same target, different args) based on a `state` comparison. |
| [`if-then-else-for-days.owl`](./if-then-else-for-days.owl) | `if`/`then`/`else` at the `blockItem` level, choosing which of two declared `day`s runs. |
| [`if-then-else-for-blocks.owl`](./if-then-else-for-blocks.owl) | `if`/`then`/`else` at the `topLevelItem` level, choosing which of two declared `block`s runs. |
