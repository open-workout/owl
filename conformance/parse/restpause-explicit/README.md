Demonstrates the named `restpause NAME = { ... }` declaration
([`spec/semantics/groups.md`](../../../spec/semantics/groups.md) §3.9):
`burnout`'s three sets compile to a `Group{kind:"restpause", interleave:
sequential, rest:{single, duration:15s}, termination:count(3),
atomic:true}`, held both in `curls`'s `groups` (as
`NamedGroup{name:"burnout", ...}`, since it's named and addressable as
`burnout.top`) and embedded directly in `curls`'s `body`.

Same shape as [`../dropset-explicit/`](../dropset-explicit/)'s
`dropset`, differing only in `rest.single`'s duration (`15s` instead of
omitted/`0s`) and that the two auto-bursts hold the *same* target/load
as `top` rather than a descending percentage of it — rest-pause doesn't
reduce the load across bursts.

`progress = { if burnout.top.reps >= 10 then ... }` reads
`burnout.top.reps` as what was actually logged for the `top` `SetRef`
inside `burnout` — compiled to `{"type":"log","label":"top",...}`, the
qualification resolved away to the bare label, exactly as a dropset's
`progress` read does (`groups.md` §3.8/§3.9, `progression.md` §2).

Compare [`../restpause-sugar/`](../restpause-sugar/) — the same
canonical `Group`, produced by the `rest_pause(...)` sugar instead,
differing only in `groups` being empty (the group is anonymous, so
`top` stays reachable unqualified).
