# Resolution — `resolve(program, state) → session`

Structural compilation ([`groups.md`](./groups.md)) produces the
[canonical form](../canonical-form.md): a fully expanded `Group` tree that
is the same for every athlete running the program. `resolve` is the
second, per-athlete pass that turns one specific occurrence of that tree
into concrete numbers the app can render and the athlete can actually
perform today.

```
resolve(program: CanonicalProgram, state: State) -> Session
```

## 1. Inputs

- **`program`**: the canonical form — already macro-expanded (`rounds`,
  `name*N` unrolled per [`groups.md`](./groups.md) §3), still holding
  unresolved `Expr` nodes wherever a `set` line referenced a `state` path
  or a catalog field.
- **`state`**: this athlete's current values for every dotted path
  declared in the program's `state`/`stats` section (see
  [`parameters.md`](./parameters.md)), plus a position cursor — which
  block/day the athlete is next due for. `resolve` does not decide
  *which* day to run; it is handed the day (or block, for a `superset`/
  `emom`/etc. spanning multiple exercises) already selected by that
  cursor, and resolves it.

## 2. What resolution does

Walk the selected subtree and, for every `Expr` node encountered inside a
`target` or `load` (see [`targets-loads.md`](./targets-loads.md)):

1. **Catalog field reference** (`$Exercise.field`, e.g.
   `$BarbellBackSquat.e1rm`) — look up this athlete's most recent recorded
   value for that field against that catalog entry.
2. **Dotted `state` path** (`tm.squat.weight`) — **if `state` already holds
   a stored value for this exact path (written by a previous
   [`progress`](./progression.md) call), use it directly — do not
   re-evaluate the binding's defining expression.** The `state` section's
   expression for a path is only ever evaluated to produce that path's
   *first* value, before `progress` has written to it for the first time;
   from then on the stored value is authoritative and the formula is not
   consulted again. This matches how real progression schemes work: `0.85
   * $BarbellBackSquat.e1rm` seeds a starting training max, but every
   session after that tracks its own trajectory independent of the
   athlete's e1rm.
3. **Arithmetic** (`*`) — evaluate once both sides are resolved to numbers.
4. **Fallback** (`expr | fallback`) — if step 1 or 2 finds no recorded
   value for this athlete yet, substitute `fallback` instead of failing.
   Two fallback shapes exist (see `grammar.ebnf`'s `fallback` production):
   - a numeric literal (`| 60`, `| 20+`) — used directly as the resolved
     value;
   - a sentinel identifier (`| AMW`) — resolution cannot produce a number
     on its own; see [`parameters.md`](./parameters.md) § Onboarding for
     how the app must collect it instead.

The output, `Session`, is structurally identical to the input subtree
except:

- every `Expr` has been replaced by a plain number (with its original
  unit) — `Target.plus`/fallback provenance is preserved as metadata (the
  UI needs to know a target was "8-12 reps, top set" rather than just "10
  reps" — see `progression.md`'s use of the rep range at log time);
- every `Ref` member has been **inlined**: replaced by the actual
  (recursively resolved) declaration it names — an `ExerciseDecl`'s `body`,
  a `NamedGroup`'s `group`, another `Day`'s `body`, etc. A `Session` is
  meant to be handed straight to the app for execution, so it carries no
  further name lookups — `superset(bench, row)`'s `Ref(bench)`/`Ref(row)`
  become the fully resolved sets themselves.

## 3. Idempotence and caching

`resolve` is a pure function of `(program, state)` for a given cursor
position — running it twice with the same inputs produces the same
`Session`. `state` only changes via [`progress`](./progression.md), which
runs strictly *after* a session is logged, never during `resolve` itself.
This means the app can safely resolve a session ahead of time (e.g. to
show tomorrow's workout in a preview) without side effects.
