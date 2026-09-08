Demonstrates an `exerciseItem`-level conditional
([`spec/semantics/conditionals.md`](../../../spec/semantics/conditionals.md)
§2, `Conditional<ProgressionRule>`), minimized from
[`if-then-else-for-progress.owl`](../../programs/if-then-else-for-progress.owl):
both branches are `progress = double(top_set, ...)` rules targeting the
same set (`top_set`), differing only in the `increment` argument. This
compiles to a single `Conditional` node — not a plain `ProgressionRule` —
at `top_set`'s `progression` field, with both `double(...)` calls
preserved as its `then`/`else`. Which one applies is decided by `resolve`
(`tm.squat.weight > 150`), never at parse/compile time.
