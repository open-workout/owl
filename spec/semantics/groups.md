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
not depend on any athlete's `state`; see
[`resolution.md`](./resolution.md) for the separate, per-athlete pass that
resolves the *values* (catalog fields, fallbacks) inside the members this
document produces.

## 2. The construct → Group mapping

| Surface construct | interleave | rest | termination | atomic |
|---|---|---|---|---|
| plain sets of one exercise ("straight sets") | `sequential` | `{single}` | `count(sets)` | `false` |
| `superset(a, b, …)` | `round_robin` | `{intra:0, inter:R}` | `count(rounds)` | `false` |
| `circuit(a, b, …)` | `round_robin` | `{intra:0, inter:R}` | `count(rounds)` | `false` |
| dropset *(proposed, §5)* | `sequential` | `{intra:0}` | `count(drops)` | `true` |
| rest-pause *(proposed, §5)* | `sequential` | `{intra:15s}` | `count(bursts)` | `true` |
| `N*emom(t){ … }` | `round_robin` | `{inter: remainder}` | `emom(t, N)` | `false` |
| `amrap(t){ … }` | `round_robin` | `{ad_lib}` | `time_cap(t)` | `false` |
| `for_time { … }` | `round_robin` | `{ad_lib}` | `for_time` | `false` |

`circuit` and `superset` are the same IR shape; the surface keyword is
preserved only as a label for the athlete-facing UI, it carries no semantic
difference. Dropset and rest-pause share the same shape too — the only
knobs that differ are `rest.intra` (`0` vs. `15s`) and whether the loads
descend across members (dropset) or stay fixed (rest-pause).

## 3. Desugaring rules

### 3.1 `exercise` declaration → `straight` Group

```owl
exercise back_squat = $BarbellBackSquat {
    set work = tm.squat.reps @ 0.8 * tm.squat.weight
    progress = double(work, 3, 5, 5)
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
declarations). A `progress` line does not itself produce a `Group` — it
attaches a `ProgressionRule` (see
[`progression.md`](./progression.md)) to the `SetRef` it names.

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

- **`dropset(a, b, c)`** / **`restpause(a, b, c)`** — proposed symmetric
  with `supersetExpr` (`dropsetExpr`/`restpauseExpr` productions reusing
  `supersetArg`), compiling to the rows in §2's mapping table
  (`interleave: sequential`, `atomic: true`, differing only in
  `rest.intra`). An alternative would be modeling these as a modifier on a
  single `exercise { }` block's consecutive `set` lines rather than a
  multi-member call — needs a decision before it's added to the grammar.
- **Explicit inter-round rest for `superset`/`circuit`** — the `rest(d)`
  *between* two `ref` args sets `intra` (§3.2), but there's no attested way
  to override the `inter` (between-lap) default. A trailing `rest(d)` after
  the last member, or a keyword arg, are both plausible — needs a decision.
