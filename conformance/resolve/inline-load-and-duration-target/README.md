Corresponds to compiling
[`conformance/programs/inline-load-and-duration-target.owl`](../../programs/inline-load-and-duration-target.owl).

Two load/target shapes not covered by any other `resolve/` fixture:

1. **An inline %1RM-style load** — `squat`'s `load` composes
   `0.8 * $BarbellBackSquat.e1rm` directly in the `set` line (a
   `mul(number, catalogField)` `Expr`), not via a `state` binding.
   `state.json`'s `catalogFields` has an entry for
   `BarbellBackSquat.e1rm` (`140`), so per
   [`resolution.md`](../../../spec/semantics/resolution.md) §2 step 1
   that's looked up and the arithmetic evaluated once both operands are
   resolved (step 3): `0.8 * 140 = 112`.
2. **A per-set `DurationTarget`** — `plank`'s `set hold = 60s` has no
   `load` at all (an isometric bodyweight hold); its `target` is
   `{kind:"duration", value:60, unit:"s"}` post-resolve, the first
   `resolve/` fixture to exercise a duration (rather than reps/distance)
   target.

`state.json` has no `bindings` at all — this program declares no `state`
section, so there's nothing to look up or fall back on for either set's
`target`; only `squat`'s `load` needs the one `catalogFields` lookup
above. `body`'s two `Ref`s (`squat`, `plank`) inline to their own
resolved `body` Groups, nested per the rule in
[`../superset-circuit-groups/`](../superset-circuit-groups/).
