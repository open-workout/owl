Corresponds to compiling
[`conformance/programs/progress-and-or-dropset.owl`](../../programs/progress-and-or-dropset.owl).

Rounds out `progress` coverage beyond
[`double-progression-hit-ceiling/`](../double-progression-hit-ceiling/):
this exercises `and`/`or` combined inside a `progressionIf.cond` (the
same `BoolExpr` shape `Conditional.cond` uses — see
[`../../resolve/conditional-and-or/`](../../resolve/conditional-and-or/)
for the structural-conditional side), reading a **dropset's** logged
value rather than a flat `set`'s.

`legext`'s `progress` reads `ds.top.reps` — dropset-qualified outside
`progress`, but qualification is only needed for *lookup*; the compiled
`LogExpr` is just `{"label":"top","field":"reps"}`, dequalified, per
`spec/semantics/progression.md` §2 and the same rule demonstrated at
parse time in
[`../../parse/dropset-explicit/`](../../parse/dropset-explicit/). The
second, auto-generated drop set in `ds` isn't read by `progress` at all,
so `log.json` only needs an entry for `top`.

`cond` is `or(and(top.reps>=15, top.reps<=20), top.reps>=25)`.
`log.json` records `top` at `17` reps: `and(17>=15, 17<=20)` is true, so
the `or` short-circuits true regardless of the `>=25` arm —
`tm.leg_ext.weight += 5` (`40` → `45`). Choosing `17` (inside the
`15..20` band, not `>=25`) makes the `then` branch depend on the `and`
arm specifically, the same way
[`../../resolve/conditional-and-or/`](../../resolve/conditional-and-or/)'s
`state.json` was chosen to make `and`'s precedence over `or` actually
observable.
