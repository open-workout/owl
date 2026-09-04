Demonstrates the `drop(f1, f2, …)` sugar
([`spec/semantics/groups.md`](../../../spec/semantics/groups.md) §3.8):
`set top = 12 @ tm.leg_ext.weight drop(0.8, 0.6)` desugars to the exact
same `Group{kind:"dropset", ...}` shape as
[`../dropset-explicit/`](../dropset-explicit/)'s named form — this
fixture's `expected.json` is byte-for-byte identical to that one's
*except* `legext.groups` is empty here (the dropset has no name, so it
gets no `NamedGroup` entry — it's only ever embedded inline in `body`)
and `progress` targets the unqualified `top` directly rather than
`ds.top`, since there's no name to qualify with.
