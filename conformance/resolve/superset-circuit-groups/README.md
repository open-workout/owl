Corresponds to compiling
[`conformance/programs/superset-circuit-groups.owl`](../../programs/superset-circuit-groups.owl).

Exercises two things not covered by any other `resolve/` fixture yet:

1. **Non-`straight` group `kind`s survive resolution unchanged in shape**
   — `superset`'s and `circuit`'s own `kind`/`interleave`/`rest`/
   `termination`/`atomic` pass through `resolve` untouched; only the
   `Expr`s nested inside their members' `target`/`load` get resolved to
   plain numbers, and only their `Ref` members get inlined. `resolve`
   never re-derives a group's own shape (round count, `two_level` rest)
   — that was decided once, structurally, at compile time
   (`spec/semantics/groups.md` §3.2).
2. **`Ref` inlining nests, it doesn't flatten.** `state.json` has a
   stored `tm.bench.weight` of `65` (so per
   [`resolution.md`](../../../spec/semantics/resolution.md) §2 step 2
   that's used directly, not the `= 60` the state binding would
   otherwise seed); `tm.row.weight` has no stored binding, so its plain
   `= 50` expression is evaluated instead — both ordinary per-path rules.
   The part this fixture is actually for: `superset`'s
   `members: [Ref(bench), Ref(row)]` resolves to `members: [<bench's own
   resolved body Group>, <row's own resolved body Group>]` — each
   `Ref` replaced by the *whole* `Group` the referenced exercise's `body`
   compiled to (`kind:"straight"`, its own `termination:count(3)`),
   nested one level deeper, per `resolution.md`'s "replaced by the
   actual...body" — not spliced flat into `superset`'s own `members`
   list. Same for `circuit`'s `Ref(lunge)`/`Ref(situp)`. `situp` has no
   `load` at all (a bodyweight movement) and resolves with no `load` key,
   same as it's declared.
