Corresponds to compiling
[`conformance/programs/for-time-rounds-groups.owl`](../../programs/for-time-rounds-groups.owl).

Like [`../emom-amrap-groups/`](../emom-amrap-groups/), every `target`/
`load` here is already a plain number, so `state.json` is empty and the
only thing `resolve` actually does is inline `Ref(partA)`/`Ref(partB)`
(nested, not flattened — see `../superset-circuit-groups/`). The point
of this fixture specifically: contrasting the two `for_time` shapes
side by side, post-resolution, so it's unambiguous they're genuinely
different —

- `partA` (bare `for_time { ... }`, no `rounds`): `kind:"for_time"`,
  `interleave:"round_robin"`, its `members` are the flat statement list
  directly (`Row`, `Burpee` — one `distance` target, no `load`, next to
  a bare `reps` target).
- `partB` (`for_time { rounds [21,15,9] as n { ... } }`): the *outer*
  wrapper is also `kind:"for_time"`, but `interleave:"sequential"` (each
  round-value's lap must finish before the next starts —
  `spec/semantics/groups.md` §3.5), and its `members` are three
  `kind:"rounds"` groups, each `round_robin` with `termination:count(1)`
  — one per value substituted into the couplet.

`resolve` doesn't touch either shape's own `kind`/`interleave`/
`termination` — only the `Expr`s nested in `target`/`load` resolve to
plain numbers, and `Ref`s get inlined.
