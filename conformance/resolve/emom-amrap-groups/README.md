Corresponds to compiling
[`conformance/programs/emom-amrap-groups.owl`](../../programs/emom-amrap-groups.owl).

`state.json`/`state`'s `bindings`/`catalogFields` are all empty — every
`target`/`load` in this program is a plain number (inline catalog calls
with literal reps/weights), so there's nothing for `resolve` to look up
or fall back on; every `Expr` here is already a `number` node and
resolves to itself unchanged (`resolution.md` §2 step 3's base case).

The interesting case: `partA`/`partB` are **day-level** named groups
(`Day.groups`, the `partA = 5*emom(1min){...}` assignment form — mirrors
`ExerciseDecl.groups` for a named `dropset`/`restpause`, but at the day
level for `emom`/`amrap`/`for_time`/`superset` assignments), referenced
from `day1.body` as `Ref(partA)`/`Ref(partB)`. Per the same "nest, don't
flatten" rule demonstrated in
[`../superset-circuit-groups/`](../superset-circuit-groups/), each `Ref`
resolves to the *whole* named group's own `Group` value (its `kind`,
`rest`, `termination` all intact — `emom`'s `RemainderRest`/
`EmomTermination`, `amrap`'s `AdLibRest`/`TimeCapTermination`), nested as
`day1.body.members[i]`, not spliced flat. `day1.groups` itself survives
resolution unchanged alongside `body` — the named-group list is metadata
the app may still want (e.g. to label "Part A"/"Part B" in the UI) even
after every `Ref` elsewhere has been inlined.
