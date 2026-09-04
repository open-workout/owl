Demonstrates the named `dropset NAME = { ... }` declaration
([`spec/semantics/groups.md`](../../../spec/semantics/groups.md) §3.8):
`ds`'s three sets compile to a `Group{kind:"dropset", interleave:
sequential, rest:{single}, termination:count(3), atomic:true}`, held both
in `legext`'s `groups` (as `NamedGroup{name:"ds", ...}`, since it's
named and addressable as `ds.top`) and embedded directly in `legext`'s
`body`. The two auto-percentage sets' `load.expr` are produced by the
local-set-reference substitution in `targets-loads.md` §4 — `0.8 *
top.weight` compiles to `0.8 * tm.leg_ext.weight`, not a new node type.

`progress = double(ds.top, 8, 15, 5)` resolves `ds.top` to the `top`
`SetRef` inside `ds` and attaches there; the stored
`progression.target` is the bare label `"top"` (the qualification was
only needed for lookup — see `groups.md` §3.8).

Compare [`../dropset-sugar/`](../dropset-sugar/) — the same canonical
`Group`, produced by the `drop(...)` sugar instead, differing only in
`groups` being empty (the dropset is anonymous, so `top` stays reachable
unqualified).
