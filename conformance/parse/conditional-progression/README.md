Demonstrates `progress`'s own `if`/`then`/`else`
([`spec/semantics/progression.md`](../../../spec/semantics/progression.md)
§2, `progressionIf`), minimized from
[`if-then-else-for-progress.owl`](../../programs/if-then-else-for-progress.owl):
`top_set.reps` reads what was actually logged for `top_set` this session
(compiles to `{"type":"log","label":"top_set","field":"reps"}`, distinct
from `top_set`'s own prescribed `load` formula), and each branch
assigns `tm.squat.weight` by a different amount. Unlike the old
`Conditional<ProgressionRule>` shape this replaces, there is no
`Conditional` node here at all — `progressIf` is evaluated directly by
`progress()` against the session log, never deferred to `resolve()` the
way a `Conditional<Member>` is (see conditionals.md §2's note on why
`condExerciseItem` itself is no longer used for this).
