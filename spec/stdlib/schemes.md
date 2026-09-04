# Standard progression schemes

The canonical definitions each `scheme` name in a `progress = scheme(...)`
line resolves to. See [`semantics/progression.md`](../semantics/progression.md)
for the general contract every scheme satisfies
(`scheme(target, log, ...args, state) -> state`) — this document only
defines the concrete algorithms.

## `double`

```owl
progress = double(top_set, 8, 12, 5)
#                  ^target ^lo ^hi ^increment
```

Double progression: at a fixed weight, climb reps across sessions from
`lo` to `hi`; once `hi` is reached, add `increment` to the weight and
drop back to `lo`. Attested throughout `conformance/programs/` — every
`progress` line in the current corpus uses this scheme.

**Algorithm**, given `target`'s logged performance for the just-completed
session (`performedReps` at `performedWeight`):

1. If `performedReps < lo`: the prescribed rep floor wasn't met. Hold —
   `state` is unchanged (next session prescribes the same weight and the
   same floor `lo` again). *(Some implementations of double progression
   deload here instead of holding; OWL's canonical scheme holds. A
   deloading variant, if wanted, is a different `scheme` name, not a flag
   on `double`.)*
2. If `lo <= performedReps < hi`: reps progress. The `state` path(s) feeding
   `target`'s `reps`/`target` expression are updated so next session's
   prescribed rep count is `performedReps + 1` (still at the same weight).
3. If `performedReps >= hi`: the rep ceiling was met (or exceeded — extra
   reps beyond `hi` don't matter further, the ceiling is a threshold, not a
   cap on what the athlete may do). The `state` path(s) feeding `target`'s
   `load` expression are increased by `increment`, and the rep target
   resets to `lo`.

Step 2/3's "the state path(s) feeding `target`'s expression" is exactly
the single write-owner from
[`parameters.md`](../semantics/parameters.md) §2 — e.g. for
`exercise squat`'s `top_set = tm.squat.reps+ @ tm.squat.weight`, step 2
writes `tm.squat.reps`, step 3 writes `tm.squat.weight` (and resets
`tm.squat.reps` to `lo`).

## `linear`

```owl
progress = linear(top_set, 2.5)
#                  ^target ^increment
```

*Proposed — not yet attested in `conformance/programs/`.* The simpler
sibling of `double`: no rep range, just add `increment` to the load every
time `target` is completed as prescribed, independent of a rep ceiling.
Included here because a corpus this small (single-exercise programs,
straight sets, no rep-range logic) is common enough that `double` alone
would be overkill for it — a program author would want the simpler scheme
available. Needs an attested example before being treated as canon.
