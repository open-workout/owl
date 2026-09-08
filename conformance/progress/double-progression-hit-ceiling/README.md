Same program as
[`conformance/resolve/double-progression-fallback/`](../../resolve/double-progression-fallback/README.md),
but here the athlete already has a stored `tm.squat.weight` (100) and
`tm.squat.reps` (12) from a previous cycle — `progress` reads and writes
`state.bindings` directly and never touches the `e1rm` formula (see
`resolution.md` §2's precedence rule: a stored binding is authoritative
once one exists).

`log.json` records the top set completed at 12 reps @ 100kg. `squat`'s
`progress` block (see `canonical.json`) reads that as
`top_set.reps >= 12` — true, the rep-ceiling branch — so it assigns
`tm.squat.weight = tm.squat.weight + 5` (105) and
`tm.squat.reps = 8`, the reset floor. This is the same "hold / bump reps
/ bump weight and reset" logic the old built-in `double` scheme used to
encode positionally; see
[`spec/semantics/progression.md`](../../../spec/semantics/progression.md)
§2 for why it's ordinary code now instead.
