Demonstrates the `rest_pause(duration, burstCount)` sugar
([`spec/semantics/groups.md`](../../../spec/semantics/groups.md) §3.9):
`set top = 8 @ tm.curl.weight rest_pause(15s, 2)` expands to the
identical `Group{kind:"restpause", interleave: sequential, rest:{single,
duration:15s}, termination:count(3), atomic:true}` as
[`../restpause-explicit/`](../restpause-explicit/)'s explicit form,
with two auto-generated `SetRef`s copying `top`'s own target/load — the
load stays fixed across bursts, unlike `drop(...)`'s descending
per-burst loads.

The group is **not** named, so it gets no entry in `curls`'s `groups`
(empty, same as [`../dropset-sugar/`](../dropset-sugar/)) and is instead
embedded directly, inline, in the exercise's `body` at the position
`top` was declared. `progress = { if top.reps >= 10 then ... }` reads
`top` unqualified — sugar keeps the flat exercise-level namespace; an
explicit `restpause NAME = { ... }` always introduces a nested one (see
`../restpause-explicit/`).
