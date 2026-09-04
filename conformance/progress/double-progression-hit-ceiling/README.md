Same program as
[`conformance/resolve/double-progression-fallback/`](../../resolve/double-progression-fallback/README.md),
but here the athlete already has a stored `tm.squat.weight` (100) and
`tm.squat.reps` (12) from a previous cycle — `progress` reads and writes
`state.bindings` directly and never touches the `e1rm` formula (see
`resolution.md` §2's precedence rule: a stored binding is authoritative
once one exists).

`log.json` records the top set completed at 12 reps @ 100kg — 12 is the
rule's `hi` bound (`double(top_set, 8, 12, 5)`), so per
[`stdlib/schemes.md`](../../../spec/stdlib/schemes.md)'s `double`
algorithm step 3 (rep ceiling met): weight increases by the rule's
`increment` (5) to 105, and reps reset to `lo` (8).
