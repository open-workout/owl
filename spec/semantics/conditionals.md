# Conditionals — `if`/`then`/`else`

This document covers `grammar.ebnf`'s `cond*Item` family: what `if cond
then: A else B` compiles to, and — the one thing that makes it different
from every other construct in [`groups.md`](./groups.md) — *when* the
branch is actually chosen.

## 1. Surface syntax, four contexts

The same shape is attested at four grammar levels, each choosing between
two alternatives of whatever list it sits in:

| Context | Chooses between | Attested by |
|---|---|---|
| `exerciseItem` | two `progressDecl`s (same `target`, different args) | [`if-then-else-for-progress.owl`](../../conformance/programs/if-then-else-for-progress.owl) |
| `dayItem` | two `stmt`s (which exercise/group runs) | [`if-then-else.owl`](../../conformance/programs/if-then-else.owl) |
| `blockItem` | two `stmt`s (which day runs) | [`if-then-else-for-days.owl`](../../conformance/programs/if-then-else-for-days.owl) |
| `topLevelItem` | two `stmt`s (which block runs) | [`if-then-else-for-blocks.owl`](../../conformance/programs/if-then-else-for-blocks.owl) |

```owl
if tm.squat.weight > 150 then
    progress = double(top_set,8,12,2.5)
else progress = double(top_set,8,12,5)
```

```owl
if tm.fatigue > 7 then:
    light;
else
    heavy;
```

`cond` (grammar `condExpr`) is a relational comparison between two
[`expr`](../grammar.ebnf)s — `dottedPath relOp expr`, in every attested
example a `state` path compared against a number. `<`, `>`, `<=`, `>=`,
and `==` are all attested (the last three by
[`conformance/parse/conditional-relops/`](../../conformance/parse/conditional-relops/),
since none of the four full example programs happens to use them); `!=`
is not — see §4.

## 2. Compiled shape: `Conditional`

Unlike every construct in `groups.md`, a conditional's `cond` reads
`state` — a per-athlete value, not known until `resolve` runs. So
structural compilation does **not** pick a branch or discard the other
one: it compiles *both* branches (using the same rules `groups.md`
already defines for whatever they are — a `progressDecl` attaches a
`ProgressionRule`, a `stmt` compiles to a `Member`) and wraps them in one
new IR node:

```
Conditional<T> { cond: BoolExpr, then: T, else: T }
BoolExpr        { op: '<' | '>' | '<=' | '>=' | '==', left: Expr, right: Expr }
```

Two instantiations are attested, matching the table in §1:

- **`Conditional<Member>`** — used wherever a `Group.members` entry is
  expected (a `dayItem`/`blockItem`/`topLevelItem`-level conditional
  compiles its `stmt` branches exactly as `groups.md` §3 already
  describes, then wraps the two resulting `Member`s). `if-then-else.owl`'s
  ```owl
  if tm.variable < 15 then:
      superset(bench, row);
  else
      bench;
  ```
  compiles to a `Conditional{ cond: {op: "<", left: path(tm.variable),
  right: number(15)}, then: Group{kind: superset, ...}, else:
  Ref{name: "bench"} }` sitting in the enclosing day's `body.members` at
  the position the `if` appeared — exactly where an ordinary `stmt`
  would sit.
- **`Conditional<ProgressionRule>`** — used as `SetRef.progression`'s
  value, when an `exerciseItem`-level conditional's branches are both
  `progressDecl`s targeting the same set label (the single-owner rule,
  [`parameters.md`](./parameters.md) §2, still applies: both branches
  must target the same `state` paths, since only one `ProgressionRule`
  ever ends up attached once `resolve` picks a branch).

Both shapes carry the literal `"type": "conditional"` tag — see
[`canonical-form.schema.json`](../canonical-form.schema.json)'s
`conditionalMember`/`conditionalProgression` defs. Which one applies is
determined by *where* the node sits (a `members` array vs. a
`progression` field), not by any extra discriminant field.

## 3. Resolution: this is where the branch is picked

[`resolve`](./resolution.md) walks the canonical tree exactly as §2 of
that document describes, with one added rule, run wherever a
`Conditional` node is encountered (in a `members` array, or as a
`progression` value):

1. Resolve `cond.left` and `cond.right` as ordinary `Expr`s (resolution.md
   §2 steps 1-4 — a `state` path may already have progressed a real value,
   a catalog field may need a lookup, a fallback may apply).
2. Evaluate `left <op> right` as a plain numeric comparison.
3. Replace the `Conditional` node with its `then` branch if true, its
   `else` branch otherwise — recursively resolved the same way an inlined
   `Ref` is (resolution.md §2's last bullet). The branch *not* taken is
   dropped; a `Session` handed to the app never contains a `Conditional`
   node, just as it never contains a `Ref`.

This is a pure function of `(cond, state)` for a given cursor, so it
inherits the same idempotence guarantee resolution.md §3 already states
for the rest of `resolve`: re-resolving the same `(program, state)` picks
the same branch every time, and `state` only changes via `progress`
running after a session is logged — never mid-`resolve`.

Because the branch is picked at `resolve` time, not compile time, the
**single-owner rule** ([`parameters.md`](./parameters.md) §2) is checked
against *both* branches of a `Conditional<ProgressionRule>` up front, at
compile time — a compiler can't wait until an athlete's `state` decides
which branch runs to validate that no other `progress` rule collides with
it.

## 4. Open questions (not yet attested in `conformance/programs/`)

Flagging these here rather than asserting them as canon, per this
project's convention (compare `groups.md` §5):

- **`!=`** — the remaining obvious completion of `relOp`; no conformance
  fixture uses it yet.
- **A `setDecl` conditional branch** — `condExerciseItem`'s branches parse
  as any `exerciseItem`, so `if cond then: set _ = 10 @ 100kg else set _ =
  8 @ 100kg` is grammatically valid, but no example attests it and this
  document doesn't define its canonical-form shape. A conditional `Member`
  can't represent it directly (a `set` line isn't a `Member`, it's an
  entry in `ExerciseDecl.sets`/`body.members` alongside other sets), so it
  likely needs its own `Conditional<SetRef>` variant rather than reusing
  `conditionalMember` — an open design question, not a settled shape.
- **`else`-less conditionals** — every attested example has an `else`
  branch; there's no attested meaning for "do nothing" when `cond` is
  false (would `Conditional.else` become optional, or is a no-op
  expressed some other way — an empty `Group`?).
- **Nested/chained conditionals (`else if`)** — not attested; today an
  `else` branch that is itself a conditional item parses fine (branches
  recurse through the same item nonterminal) but no example demonstrates
  it, so the multi-way-branch reading is unconfirmed.
