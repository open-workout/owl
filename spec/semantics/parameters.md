# Parameters — namespace, ownership, onboarding

This document covers the `state`/`stats` section itself: what its dotted
paths mean, the rule that keeps multiple `progress` rules from stepping on
each other, and what happens when an athlete has no history to resolve a
value from at all.

## 1. Namespace

```owl
state = {
    tm.squat.weight = 0.85 * $BarbellBackSquat.e1rm | AMW
    tm.squat.reps   = 5
}
```

Every binding is a dotted path (`dottedPath` in `grammar.ebnf`). By
convention (attested throughout `conformance/programs/`) paths are
namespaced `tm.<exercise>.<field>` — "training max" — but the grammar
places no restriction on the path shape beyond `IDENT ('.' IDENT)*`; `tm`
is a convention, not a keyword.

`state` and `stats` are accepted as the same section (both appear across
the conformance files — `state` where bindings exist,
`superset-example.owl`'s empty `stats = {}` where there are none). Until
there's a reason to treat them as two distinct concepts (e.g. mutable
training state vs. read-only historical stats), the grammar and this spec
treat them as synonyms.

## 2. The single-owner rule

Every dotted path bound in `state` may be written by **at most one**
exercise's `progress` block anywhere in the program. An exercise's
`progress = { ... }` writes whatever `state` paths appear as the `path`
of one of its `assign` statements ([`progression.md`](./progression.md)
§2-4) — e.g. `tm.squat.weight += 5` inside `squat`'s `progress` block
claims `tm.squat.weight` — and no *other* exercise's
`progress` block may assign that same path.

This is a **validation rule enforced at compile time**, not a grammar
constraint: a compiler must reject a program where two different
exercises' `progress` blocks assign the same `state` path. It exists so
`progress`'s per-session update is always unambiguous — there is never a
"which assignment wins" question across exercises, because only one
exercise's code can ever write a given path. Within *one* exercise's own
`progress` block, the same path may appear as the `path` of `assign`
statements in multiple, mutually exclusive `if`/`else` branches (that's
the normal way to express "update this path differently depending on
what happened") — the rule only concerns *different exercises*
colliding on the same path.

A `state` path with **no** `progress` block assigning it anywhere is
valid and common — it's simply never updated automatically; it stays at
whatever value onboarding or a manual edit gave it (e.g. `tm.squat.reps =
5` in the examples is a fixed rep-scheme parameter that a `progress`
block may choose not to touch at all).

## 3. Onboarding: the sentinel fallback

```owl
tm.squat.weight = 0.85 * $BarbellBackSquat.e1rm | AMW
```

`| fallback` (see `grammar.ebnf`'s `fallback` production) has two shapes:

- **Numeric** (`| 60`, `| 20+`): used directly as the resolved value when
  the athlete has no recorded history yet. Resolution needs nothing extra
  from the app — see [`resolution.md`](./resolution.md) §2.
- **Sentinel** (`| AMW`): an uppercase identifier naming a value
  [`resolve`](./resolution.md) cannot manufacture on its own. `AMW`
  ("Assumed Max Weight") means: when this athlete has no `$Exercise.e1rm`
  recorded, the app must *ask the athlete directly* for a working number
  before this session can be resolved — a one-time onboarding prompt per
  exercise ("What's your assumed max weight for Back Squat?"), rather than
  a formula-derived default.

Once an athlete answers that prompt, the app should record it as if it
were the resolved `$BarbellBackSquat.e1rm` (or store it directly as the
`tm.squat.weight` binding's resolved value — an implementation choice not
yet fixed here), so subsequent cycles' `progress` calls have real history
to work from and the sentinel is never consulted again for that
athlete/exercise pair.

`AMW` is the only sentinel attested so far. The grammar treats `sentinel`
as a plain `IDENT` (any all-caps name), so new onboarding sentinels (e.g. a
future `BW` for "ask for bodyweight") don't require a grammar change —
just a new entry in whatever registry the app/runtime uses to map sentinel
names to onboarding prompts. That registry doesn't exist yet; it's an open
question for the reference implementation, not part of the language spec.
