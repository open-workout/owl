# Progression — `progress(program, log) → state`

```
progress(program: CanonicalProgram, log: SessionLog) -> State
```

Where [`resolve`](./resolution.md) turns `state` into a concrete session,
`progress` runs the other direction: after an athlete logs what they
actually did, it produces the *updated* `state` that the next `resolve`
call for the same `set`s will read.

Unlike every other pass in this spec, `progress` doesn't evaluate a
static shape against inputs — it **runs code**: each exercise's
`progress = { … }` block (grammar `progressDecl`/`progressBody`) is a
small imperative program, not a call into a closed set of named
stdlib schemes. There is no scheme registry; a program author writes
exactly the update rule they want, as ordinary `state`-path assignments
guarded by `if`/`then`/`else`.

## 1. Inputs

- **`program`**: the canonical form, specifically each `ExerciseDecl`'s
  optional `progress` field (`ProgressionBody`) — at most one per
  exercise (§2).
- **`log`**: what the athlete actually recorded for that session, at set
  granularity (`SessionLog`, keyed by `exercise`+`label`) — reps/
  distance/duration performed and load used, per `SetRef` that was
  resolved and presented.

An exercise with no `progress` block is simply never updated by this
pass — its sets are logged for history but `state` doesn't change
because of them (e.g. `CardioRower`'s plain distance sets in
`superset-example.owl`).

## 2. What a `progress` block is

```owl
exercise squat = $BarbellBackSquat {
    set top_set = tm.squat.reps @ tm.squat.weight
    progress = {
        if top_set.reps >= 8 then           # floor gate, on the LOGGED value —
                                             # if false, nothing below runs at
                                             # all: true hold, state untouched.
            tm.squat.reps = top_set.reps
            tm.squat.weight = top_set.weight
            if tm.squat.reps >= 12 then
                tm.squat.weight += 5
                tm.squat.reps = 8
            else
                tm.squat.reps += 1
    }
}
```

A `progressBody` is a statement list (`ProgressionStmt[]`) of two kinds:

- **`assign`** (`path op= expr`, `op` one of `=`, `+=`, `-=`, `*=`,
  `/=`): writes one `state` path. A compound `op=` is sugar, desugared
  at parse time into `path = path op expr` — `tm.squat.weight += 5`
  compiles to the exact same `Assign` node `tm.squat.weight =
  tm.squat.weight + 5` would, reading the path's own current value as
  the left operand of an ordinary `add`/`sub`/`mul`/`div` `Expr`. No new
  canonical-form shape; see `grammar.ebnf`'s `assignOp` production.
  Compound assignment is scoped to `progressAssign` only — a `state`
  binding's `=` seeds a value once and a named-group `=` names a group
  expression, neither has a "current value" to mutate.
- **`if`** (`progressionIf`): picks one of two statement lists to run,
  based on `cond` — evaluated once, against *this session's* data (§3),
  not deferred to a later pass the way [`conditionals.md`](./conditionals.md)'s
  `conditionalMember` is. `else` is optional here (nowhere else in the
  language is it); an omitted or untaken branch simply performs no
  assignments — this is how "hold, don't update" (`double`'s old step 1)
  is expressed: a branch (or, as in the outer `if` above, an entire
  gated block) that assigns nothing.

Every `dottedPath` inside a `progressBody`'s `expr`s means something
slightly different than it does everywhere else in the language — see
grammar.ebnf's note under `dottedPath` and the walkthrough below.

**At most one `progress` per exercise.** Unlike the old scheme-call
model, a `progress` block isn't attached to one `set` — it belongs to
the whole exercise and can read/write across all of that exercise's
sets and `state` paths in one place. Two sets in the same exercise
needing different update rules is expressed as one `progress` block
with an `if`/`else` distinguishing them (by whichever `state`/log values
tell them apart), not two separate `progress` lines.

## 3. Reading values inside a `progress` block

Two different things can appear on the right-hand side of `=` or inside
a `cond`:

- **`tm.squat.weight`** (a bare `state` path) — the athlete's *current*
  value for that path, exactly as `state` holds it right now (before
  this `progress` call updates anything). Reading it doesn't re-run its
  defining `expr`/fallback — same rule as [`resolution.md`](./resolution.md)
  §2: once a path has a stored value, that value is authoritative.
- **`top_set.reps`** / **`top_set.weight`** (a `set` label, bare or
  dropset-qualified, dotted with a field) — what was *actually logged*
  for that specific set this session (from `log`), **not** the
  prescribed target/load formula that same spelling would mean inside a
  `set` line's own `target`/`load` (`targets-loads.md` §4). `top_set` here
  must name a `set` declared somewhere in the same exercise (bare, or as
  `dropsetName.label` if it's inside a named `dropset`); the field (e.g.
  `reps`, `weight`) picks which of the log entry's `performedTarget`/
  `performedLoad` to read. Compiles to a dedicated `LogExpr` node
  (`canonical-form.schema.json`'s `expr`'s `"type": "log"` variant) —
  not a `PathExpr` — precisely so it's never confusable with a `state`
  read once compiled, even though the source spelling can look similar
  to a local set-field reference.

`double`'s old three-way branch, reimplemented as ordinary code, reads
exactly as it did in prose: "if the logged reps are below the floor,
hold; else copy the log into `state` and, if it hit the ceiling, bump
the weight and reset reps — otherwise bump reps" — no scheme name, no
positional `args` tuple, just `if`/`then`/`else`, assignment, and
arithmetic over named values (§2's example is the canonical shape every
`double`-style reimplementation in `conformance/` follows).

The floor check has to test the **logged** value (`top_set.reps`), not
`tm.squat.reps`, and has to gate the copy-down assignments rather than
follow them: were it written the other way — copy first, then check
`tm.squat.reps` — a below-floor session would already have overwritten
`state` with the sub-floor logged values by the time the floor check
ran, which isn't "hold" at all, just a training max that silently drops
to whatever the athlete happened to log. Testing the log value first,
before any assignment, is what makes "false → skip the whole block →
`state` never touched" an actual hold.

## 4. Ownership

Since a `progress` block writes `state` paths directly (`assign`'s
`path`), there's no more inference step (no more "the paths its target's
expression reads") — the **single-owner rule**
([`parameters.md`](./parameters.md) §2) now just walks every `assign` in
every exercise's `progress` block and rejects the program if the same
path is assigned by two different exercises. Multiple `assign`s to the
same path within *one* exercise's own block (e.g. across its own
`if`/`else` branches) are fine — at most one of them ever executes for
a given session, since they're mutually exclusive branches.

## 5. Aggregation across a program run

`progress` is invoked once per completed session (not per set, not per
block) — it runs every exercise's `progress` block against that
session's log and returns one updated `State`. A multi-day block (e.g.
`squat-everyday.owl`'s `SE_lower`/`SE_upper`) calls `progress`
independently at the end of each day; a block repeated with `name*N`
(`leader*5`) calls it once per iteration, so training maxes can move
between cycles, not just between days within one cycle.
