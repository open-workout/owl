# Targets and Loads

Every `set` line and inline catalog call in the grammar
(`quantity ('@' quantity)?`, see [`grammar.ebnf`](../grammar.ebnf)) carries
one or two `quantity` values. This document splits that single surface
production into the two semantically distinct unions the canonical form
actually stores: **Target** (what the athlete is asked to do) and
**Load** (what they're asked to do it against).

```owl
set work = 10 @ 0.8 * tm.squat.weight   #  target=10 reps        load=0.8*tm.squat.weight
set _     = 100m                        #  target=100m distance  load=(none)
$KettlebellSwing( 15 @ 24kg )           #  target=15 reps        load=24kg
```

## 1. Target

`Target` is a tagged union of the three kinds attested in
`conformance/programs/`:

```
Target = Reps     { count: Expr, plus: bool }   // "10", "tm.squat.reps+"
       | Distance { amount: Expr, unit: "m"|"km" }   // "100m"
       | Duration { amount: Expr, unit: "s"|"m"|"h" }  // attested only as group-level
                                                        // durations (emom/amrap/rest),
                                                        // not yet as a per-set target —
                                                        // included for e.g. a plank hold
```

`plus` is the surface `+` marker (`tm.squat.reps+`): "this many reps *or
more*" — an open-ended/AMRAP-style target rather than a fixed count. It is
only meaningful on `Reps`. The same marker on a nominal floor (`1+`, as in
a dropset's auto-generated sets — [`groups.md`](./groups.md) §3.8) reads
as "to failure" — it's the same open-ended semantics, just anchored at
the minimum instead of a working number.

A `set` with only one `quantity` and no `@` (e.g. `CardioRower`'s
`set _ = 100m`) has a `Target` and no `Load` at all — some movements are
measured purely by output (distance, calories, time), with no separate
resistance to prescribe.

## 2. Load

`Load` is, in every attested example, a weight expression:

```
Load = Weight { amount: Expr, unit: "kg"|"lb" }
```

where `Expr` is the same recursive arithmetic node used in `state`
(numeric literals, `*`, dotted `state` paths like `tm.squat.weight`, and
catalog field references like `$BarbellBackSquat.e1rm`) — see
[`resolution.md`](./resolution.md) for how `Expr` is resolved to a
concrete number for a specific athlete.

No example yet expresses a load as "percentage of 1RM" independent of a
`state` binding (`0.8 * tm.squat.weight` always goes through a named `tm.*`
value, never `0.8 * $BarbellBackSquat.e1rm` directly inside a `set` line —
that composition only appears inside `state`). Nor is there an attested
bodyweight-only load marker (as opposed to simply omitting `Load`
entirely, which is what `CardioRower` does). Both would be natural
additions to the `Load` union if a program needs them; treat the union
above as *what's proven*, not a closed set.

## 3. Why the split matters

The grammar's `quantity ('@' quantity)?` doesn't distinguish the two
slots — both are the same `quantity` production. The canonical form does,
because they mean different things downstream:

- Only `Target.plus` participates in [progression](./progression.md)'s
  rep-range logic (`double(top_set, 8, 12, 5)` reads the *logged* reps
  against the target's rep range).
- Only `Load` participates in plate-math (rounding a resolved weight to
  the nearest achievable value from the `plates` array) — a `Distance` or
  `Duration` target is never run through the plate calculator.
- A `Target`-only set (no `Load`) is a valid, common case (cardio,
  bodyweight-for-reps); a `Load`-only set with no `Target` is not
  meaningful and should be rejected at compile time.

See [`canonical-form.schema.json`](../canonical-form.schema.json)'s
`setRef.target` / `setRef.load` fields for the JSON shape.

## 4. Local set references

A `Load`'s `expr` can reference an earlier `set` in the same exercise or
`dropset`, not just a `state` path — this is how a dropset expresses "80%
of my top set" without hardcoding a number:

```owl
set top = 12 @ tm.leg_ext.weight
set _   = 1+ @ 0.8 * top.weight     # 80% of top
```

Syntactically `top.weight` is nothing new — it's the same `dottedPath`
production as `tm.squat.weight` (`grammar.ebnf`'s `dottedPath ::= IDENT
('.' IDENT)*`). What makes it a *local set reference* rather than a
`state` path is purely name resolution: the compiler resolves a
`dottedPath`'s leading segment by checking, in order, whether it names an
earlier `set` label in scope, a `dropset` name in scope, or a `state`
binding root — see `grammar.ebnf`'s note under `dottedPath` for the full
rule, and [`groups.md`](./groups.md) §3.8 for how this is used inside
`dropset`.

Structurally, `X.weight` compiles by substituting a **copy of `X`'s own
`load.expr`** wherever it appears — the same compile-time substitution
`groups.md` §3.5 uses for `rounds ... as n`, not a new runtime-resolved
reference. `0.8 * top.weight` where `top`'s load is `tm.squat.weight`
compiles to the same `Expr` tree as if you'd written `0.8 *
tm.squat.weight` directly — no new node type in
`canonical-form.schema.json`'s `expr` def, and no dependency on `top`'s
*actual logged* performance (OWL has no live mid-session re-resolution;
see `resolution.md` §3). Only `.weight` is attested; a `.reps`/`.target`
variant (substituting the referenced set's `target.expr` instead) would
follow the same rule but isn't demonstrated anywhere yet.
