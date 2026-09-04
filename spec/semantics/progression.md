# Progression — `progress(program, log) → state`

```
progress(program: CanonicalProgram, log: SessionLog) -> State
```

Where [`resolve`](./resolution.md) turns `state` into a concrete session,
`progress` runs the other direction: after an athlete logs what they
actually did, it produces the *updated* `state` that the next `resolve`
call for the same `set` will read.

## 1. Inputs

- **`program`**: the canonical form, specifically the `ProgressionRule`
  attached to each `SetRef` via the surface `progress = scheme(target,
  …args)` line (see [`groups.md`](./groups.md) §3.1).
- **`log`**: what the athlete actually recorded for that session, at
  set granularity — at minimum, reps/distance/duration performed and load
  used, per `SetRef` that was resolved and presented.

`progress` only inspects the `SetRef`s that carry a `ProgressionRule`; sets
without one (e.g. `CardioRower`'s plain distance sets in
`superset-example.owl`) are logged for history but never change `state`.

## 2. Granularity and ownership

A `ProgressionRule`'s `target` names exactly one `set` label (e.g.
`double(top_set, 8, 12, 5)` targets the set labeled `top_set`). Evaluating
it reads that one set's logged performance and writes to *exactly the
`state` path(s) that set's own `load`/`target` expressions read from* —
enforced by the single-owner rule in
[`parameters.md`](./parameters.md) §2: only one `progress` line in the
whole program may write a given `state` path, so there is never a
question of write-ordering or conflicting updates across exercises that
happen to share a training max.

## 3. Aggregation across a program run

`progress` is invoked once per completed session (not per set, not per
block) — it evaluates every progressable `SetRef` that session touched and
returns one updated `State`. A multi-day block (e.g. `squat-everyday.owl`'s
`SE_lower`/`SE_upper`) calls `progress` independently at the end of each
day; a block repeated with `name*N` (`leader*5`) calls it once per
iteration, so training maxes can move between cycles, not just between
days within one cycle.

## 4. Schemes

`scheme` (the identifier before the args, e.g. `double`) selects the
update rule; `args` are its positional parameters, stored generically
(`ProgressionRule.args: number[]`) precisely because different schemes
need different shapes — the grammar's `progressDecl` places no constraint
beyond `IDENT '(' argList ')'`. The concrete, canonical definition of each
built-in scheme's algorithm lives in
[`stdlib/schemes.md`](../stdlib/schemes.md), not here — this document
defines the *contract* every scheme must satisfy:

```
scheme(target: SetRef, log: SetLog, ...args: number[], currentState: State) -> State
```

i.e. a pure function from "what was prescribed, what was logged, the
scheme's own tuning numbers, and the athlete's current state" to "the
athlete's next state" — touching only the `state` path(s) `target`'s
`load`/`reps` expressions read from (§2).
