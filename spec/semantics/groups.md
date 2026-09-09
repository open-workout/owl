# Groups — the 4-knob node

This document defines what [`grammar.ebnf`](../grammar.ebnf)'s group
expressions (`superset`, `circuit`, `emom`, `amrap`, `for_time`, `rounds`,
and plain sequential exercise use) compile to in the
[canonical form](../canonical-form.md).

## 1. The core insight: everything is a `Group`

Every construct that sequences or interleaves work — a plain set of an
exercise, a superset, a circuit, a dropset, rest-pause, an EMOM, an AMRAP, a
for-time couplet — compiles to the **same** IR node:

```
Group {
  members:     [Member]       // ordered; a Ref (by name), SetRef, or a nested Group
  interleave:  Pattern         // how members' event-streams merge
  rest:        RestPolicy      // rest at each transition (two-level: intra / inter)
  termination: Termination     // when the group ends
  atomic:      bool            // is this one indivisible event to an outer group?
}
```

- **`interleave`**: `sequential` (finish member A entirely, then B) or
  `round_robin` (members take turns, one lap each, in list order). A
  round-robin group with a single lap is observationally identical to
  sequential — this is how a for-time couplet with no repeated rounds and a
  3-round superset both fall out of the same `round_robin` tag.
- **`rest`**: either `{single}` (the plain rest-between-sets case),
  `{intra, inter}` (two-level: rest *between members within one lap*, rest
  *between laps*), `{ad_lib}` (athlete-paced, no prescribed rest — AMRAP/
  for-time), or `{remainder}` (EMOM: whatever's left of the interval after
  the work is done).
- **`termination`**: `count(n)` (fixed number of sets/rounds/drops/bursts),
  `emom(interval, n)`, `time_cap(t)`, or `for_time` (no cap; the clock runs
  and elapsed time is the recorded result).
- **`atomic`**: `true` for constructs that must never be split up by an
  *outer* group's own interleaving (a dropset's three loads are one
  indivisible unit even if it sits inside a larger circuit).

This structural expansion — including the desugaring rules in §3 below — is
athlete-agnostic and happens once, at parse-to-canonical-form time. It does
not depend on any athlete's `state` — the one exception is `if`/`then`/
`else`, whose `cond` *does* read `state`; see
[`conditionals.md`](./conditionals.md) for the separate `Conditional` node
this produces and why picking its branch is deferred to `resolve`. See
[`resolution.md`](./resolution.md) for the separate, per-athlete pass that
resolves the *values* (catalog fields, fallbacks) inside the members this
document produces.

## 2. The construct → Group mapping

| Surface construct | interleave | rest | termination | atomic |
|---|---|---|---|---|
| plain sets of one exercise ("straight sets") | `sequential` | `{single}` | `count(sets)` | `false` |
| `superset(a, b, …)` | `round_robin` | `{intra:0, inter:R}` | `count(rounds)` | `false` |
| `circuit(a, b, …)` | `round_robin` | `{intra:0, inter:R}` | `count(rounds)` | `false` |
| `dropset NAME = { ... }` / `drop(...)` sugar (§3.8) | `sequential` | `{single: 0s}` | `count(drops)` | `true` |
| rest-pause *(proposed, §5)* | `sequential` | `{single: 15s}` | `count(bursts)` | `true` |
| `N*emom(t){ … }` | `round_robin` | `{inter: remainder}` | `emom(t, N)` | `false` |
| `amrap(t){ … }` | `round_robin` | `{ad_lib}` | `time_cap(t)` | `false` |
| `for_time { … }` | `round_robin` | `{ad_lib}` | `for_time` | `false` |

`circuit` and `superset` are the same IR shape; the surface keyword is
preserved only as a label for the athlete-facing UI, it carries no semantic
difference. Dropset and rest-pause share the same shape too — the only
knobs that differ are `rest`'s `single` duration (`0s` vs. `15s`) and
whether the loads descend across members (dropset) or stay fixed
(rest-pause). Both use `RestPolicy`'s `single` mode, not `two_level` — a
sequential/atomic group has no "lap" concept for `intra` vs. `inter` to
distinguish. `two_level` is reserved for groups where that distinction is
real: `superset`/`circuit` above. (`emom`'s `{inter: remainder}` and
`amrap`/`for_time`'s `{ad_lib}` are their own single-value `RestPolicy`
modes for the same reason — see §1.)

## 3. Desugaring rules

### 3.1 `exercise` declaration → `straight` Group

```owl
exercise back_squat = $BarbellBackSquat {
    set work = tm.squat.reps @ 0.8 * tm.squat.weight
    progress = {
        if work.reps >= 3 then
            tm.squat.reps = work.reps
            tm.squat.weight = work.weight
            if tm.squat.reps >= 5 then
                tm.squat.weight += 5
                tm.squat.reps = 3
            else
                tm.squat.reps += 1
    }
}
```

compiles to an `ExerciseDecl{ name: "back_squat", catalog: "BarbellBackSquat",
sets: [SetRef, …] }` — one `SetRef{ target, load }` per `set` line (see
[`targets-loads.md`](./targets-loads.md) for the `target`/`load` shapes).
The declaration's own `body` is precomputed as `Group{ interleave:
sequential, rest: {single}, termination: count(N sets), atomic: false,
members: sets }`, so that a later bare-name statement (`back_squat;`)
compiles to a `Ref{ name: "back_squat" }` member that resolves to this
`body` — the sets are declared once, not duplicated at every use site
(e.g. inside `superset(bench, row)`, `bench`/`row` are `Ref`s into the same
declarations). A `progress` line does not itself produce a `Group` or
attach to any one `SetRef` — it compiles to the exercise declaration's
own `progress` field (see [`progression.md`](./progression.md)), a
single code block covering the whole exercise.

### 3.2 `superset(a, b, …)` → `superset` Group

```owl
superset(bench, row);
```

→ `Group{ kind: superset, members: [Ref(bench), Ref(row)],
interleave: round_robin, rest: {intra: 0s, inter: 90s}, termination:
count(rounds), atomic: false }`, where `bench`/`row` resolve against the
enclosing day's declared `exercise`s rather than duplicating their `set`
lines inline.

Per the comment in `superset-example.owl` ("superset can accept any number
of params, also accepts rest blocks as params"), a `rest(d)` argument
appearing between two `ref` arguments sets the `intra` rest between exactly
that adjacent pair, overriding the `0s` default:

```owl
superset(bench, rest(20s), row);   # 20s between bench and row, still 0 default elsewhere
```

`termination.count(rounds)` is inferred from the member exercises' own
declared set counts (all members of one superset must agree on round
count — a validation rule, not a syntax rule). The `inter` (between-lap)
rest value has no attested surface syntax yet — see §5.

### 3.3 `N*emom(t){ body }` → `emom` Group

```owl
partA = 5*emom(3m){ back_squat }
```

→ `Group{ kind: emom, members: [<compiled body>], interleave: round_robin,
rest: {inter: remainder}, termination: emom(3m, 5), atomic: false }`. The
body is compiled as an ordinary statement list (§3.1-style); "remainder"
means whatever time is left in each 3-minute interval after the body
finishes becomes rest before the next interval starts.

### 3.4 `amrap(t){ body }` → `amrap` Group

Same shape as EMOM but `rest: {ad_lib}` (no forced interval — go
immediately into the next lap) and `termination: time_cap(t)` (stop when
the clock hits `t`, recording completed rounds + partial reps).

### 3.5 `for_time { body }` → `for_time` Group

```owl
partB = for_time {
    rounds [21, 15, 9] as n {
        $BarbellThruster( n @ 42.5)
        $PullUp( n )
    }
}
```

`for_time` itself contributes `rest: {ad_lib}` and `termination: for_time`
(no cap — the clock records elapsed time to completion) to whatever
grouping structure is inside its braces. Two shapes can appear inside:

- **A flat statement list** (no `rounds`): compiles directly to the
  `for_time` Group with `interleave: round_robin` over those statements,
  one lap.
- **A `rounds [v1, …, vk] as x { body }` block**: this is a *repetition
  macro*, resolved at compile time, not its own IR node. It expands to `k`
  member groups, one per value, each `body` with every free occurrence of
  `x` substituted by that value:

  ```
  Group {
    kind: for_time,
    interleave: sequential,          // laps of the rounds list run strictly one after another
    rest: {ad_lib},
    termination: for_time,
    atomic: false,
    members: [
      Group{ members:[Thruster(21@42.5), PullUp(21)], interleave: round_robin, rest:{ad_lib}, termination: count(1), atomic:false },
      Group{ members:[Thruster(15@42.5), PullUp(15)], interleave: round_robin, rest:{ad_lib}, termination: count(1), atomic:false },
      Group{ members:[Thruster( 9@42.5), PullUp( 9)], interleave: round_robin, rest:{ad_lib}, termination: count(1), atomic:false },
    ]
  }
  ```

  Each per-value member is itself the ordinary "flat statement list"
  compilation from the bullet above (round-robin with a single lap, which
  is why "Fran" reads as thruster-then-pullup rather than alternating reps
  — a 1-lap round-robin degenerates to sequential order). The outer
  `rounds` wrapper is `sequential` because value `v1`'s entire lap must
  finish before `v2` starts.

`rounds … as x { }` is not specific to `for_time` — the same macro applies
anywhere a group body appears (e.g. nesting it inside an `amrap`), it just
happens to be attested only under `for_time` in the conformance files.

### 3.6 Inline catalog call → anonymous `SetRef`

```owl
$KettlebellSwing( 15 @ 24kg )
```

compiles to a bare `SetRef{ exercise: $KettlebellSwing, target: 15 reps,
load: 24kg }` member — the anonymous-single-set equivalent of a full
`exercise { set … }` declaration, used when the movement doesn't need a
name for later reference or progression.

### 3.7 `name*N` → repeated sequential Group

```owl
leader*5;
```

→ `Group{ interleave: sequential, rest: {single}, termination: count(5),
atomic: false, members: [copy of `leader`'s compiled body] × 5 }`. Same
repetition-macro treatment as `rounds`, just with a fixed body instead of a
per-lap substitution.

### 3.7a `(stmt, stmt, …)*N` → repeated sequential Group, anonymous body

```owl
(cardio_rower, rest(2min))*5
```

Same macro as §3.7, just with an inline, anonymous statement list in
place of a declared name — there is nothing to name when the body is
only ever used at this one repeat site. Compiles to `Group{ interleave:
sequential, rest: {single}, termination: count(5), atomic: false,
members: [copy of the compiled `(cardio_rower, rest(2min))` body] × 5 }`,
where the inner body is compiled exactly as any other statement list
(§3.1-style): `cardio_rower` → `Ref{name: "cardio_rower"}`, and the bare
`rest(2min)` — since it's not a `supersetArg`'s per-pair rest override
(§3.2) — sets that inner Group's own `rest` to `{single: 2min}` rather
than becoming a member of its own. The same bare-`rest(d)`-sets-the-
enclosing-group's-`rest` reading applies wherever a `restStmt` appears
directly in a sequential statement list, e.g. `SE_lower; rest(1d);
SE_upper;` at block level.

### 3.8 `dropset` declaration → `dropset` Group

```owl
exercise legext = $LegExtension {
    dropset ds = {
        set top = 12 @ tm.leg_ext.weight
        set _   = 1+ @ 0.8 * top.weight     # to failure, 80% of top
        set _   = 1+ @ 0.6 * top.weight     # to failure, 60% of top
    }
    progress = {
        if ds.top.reps >= 15 then
            tm.leg_ext.weight += 5
        else if ds.top.reps >= 8 then
            tm.leg_ext.weight += 2.5
    }
}
```

→ `NamedGroup{ name: "ds", group: Group{ kind: dropset, interleave:
sequential, rest: {single: 0s}, termination: count(3), atomic: true,
members: [SetRef(top), SetRef(_), SetRef(_)] } }`, held in the exercise
declaration's `groups` list (`ExerciseDecl.groups`, mirroring `Day.groups`
for `partA`-style assignments — see [`canonical-form.md`](../canonical-form.md)).
Each `_`-labeled set's `load.expr` is produced by the local-set-reference
substitution in [`targets-loads.md`](./targets-loads.md) §4 — `0.8 *
top.weight` compiles to `0.8 * tm.leg_ext.weight`, a plain composed `Expr`
with no new node type.

Naming the dropset (`ds`) makes it a scope: `top` is only reachable from
outside as `ds.top` (as `progress`'s `ds.top.reps` read does above) — see
`grammar.ebnf`'s note under `dottedPath`. That same `ds.top.reps` inside
`progress` means something different than `top.weight` would mean inside
one of `ds`'s own `set` lines: there it's the *logged* value for `top`,
not a prescribed-formula substitution — see
[`targets-loads.md`](./targets-loads.md) §4's note and
[`progression.md`](./progression.md) §2. A dropset changes where a set
sits structurally, not which sets `progress` can read from — the
exercise's one `progress` block can reference any of its sets, dropset
member or not.

**Sugar**: `drop(f1, f2, …)` on a `set` line desugars to the same
`dropset` `Group` shape, anonymously:

```owl
exercise legext = $LegExtension {
    set top = 12 @ tm.leg_ext.weight drop(0.8, 0.6)   # top + two auto-drops to failure
    progress = {
        if top.reps >= 15 then
            tm.leg_ext.weight += 5
    }
}
```

expands to the identical `Group{ kind: dropset, ... }` as the explicit
form, with one `SetRef{ target: {kind: reps, expr: 1, plus: true}, load:
{expr: fᵢ * <annotated set>.load.expr} }` per factor `fᵢ`, in order —
but the dropset is **not** named, so it gets no entry in
`ExerciseDecl.groups` and is instead embedded directly, inline, in the
exercise's `body` at the position `top` was declared (the same treatment
a flat `set` already gets — see §3.1). Because there's no name to
qualify with, `top` stays reachable directly (`top.reps`, not
`ds.top.reps`) — sugar keeps the flat exercise-level namespace; an
explicit name always introduces a nested one.

`drop(...)`'s factors are always multipliers on the annotated set's own
`load` — there's no attested syntax for drop-setting a load-less
(distance/duration-only) set, and a compiler should reject `drop(...)` on
one. `1+` (open-ended reps) is being used here specifically to mean "to
failure" — see `targets-loads.md` §1's note on `plus`.

## 4. Where structural compilation ends

Everything above runs once per program, independent of any athlete. The
result — a fully expanded `Group` tree, still holding unresolved `expr`
nodes (catalog field references, dotted `state` paths, fallbacks) inside
its `target`/`load` quantities — *is* the canonical form
([`canonical-form.md`](../canonical-form.md)). Turning that tree into
concrete numbers for a specific athlete on a specific day is
[`resolve`](./resolution.md)'s job, not this pass's.

## 5. Proposed extensions (not yet attested in `conformance/programs/`)

These follow directly from the group model above, but no `.owl` file
demonstrates surface syntax for them yet. Flagging them here rather than
asserting them as canon:

- **`restpause`** — proposed as the same shape `dropset` uses (§3.8: a
  named `restpause NAME = { ... }` declaration inside an `exercise { }`,
  or a `rest_pause(15s)`-style sugar modifier on a `set` line), differing
  only in `rest`'s `single` duration (`15s` instead of `0`) and that the
  load stays fixed across bursts rather than descending. Not added to the grammar
  yet — no attested example to confirm the sugar's exact spelling.
- **Explicit inter-round rest for `superset`/`circuit`** — the `rest(d)`
  *between* two `ref` args sets `intra` (§3.2), but there's no attested way
  to override the `inter` (between-lap) default. A trailing `rest(d)` after
  the last member, or a keyword arg, are both plausible — needs a decision.
