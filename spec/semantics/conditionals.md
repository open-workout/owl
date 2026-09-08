# Conditionals — `if`/`then`/`else`

This document covers `grammar.ebnf`'s `cond*Item` family: what `if cond
then: A else B` compiles to, and — the one thing that makes it different
from every other construct in [`groups.md`](./groups.md) — *when* the
branch is actually chosen.

## 1. Surface syntax, three contexts

The same shape is attested at three grammar levels, each choosing
between two alternatives of whatever list it sits in:

| Context | Chooses between | Attested by |
|---|---|---|
| `dayItem` | two `stmt`s (which exercise/group runs) | [`if-then-else.owl`](../../conformance/programs/if-then-else.owl) |
| `blockItem` | two `stmt`s (which day runs) | [`if-then-else-for-days.owl`](../../conformance/programs/if-then-else-for-days.owl) |
| `topLevelItem` | two `stmt`s (which block runs) | [`if-then-else-for-blocks.owl`](../../conformance/programs/if-then-else-for-blocks.owl) |

```owl
if tm.fatigue > 7 then:
    light;
else
    heavy;
```

A fourth level, `condExerciseItem`, still parses (grammar symmetry with
the other three) but has no attested use and is rejected by the
compiler — see §2's note.

`cond` (grammar `condExpr`) is a relational comparison between two
[`expr`](../grammar.ebnf)s — `dottedPath relOp expr`, in every attested
example a `state` path compared against a number, optionally combined
with `and`/`or` (§2a). `<`, `>`, `<=`, `>=`, `==`, and `!=` are all
attested (the last four by
[`conformance/parse/conditional-relops/`](../../conformance/parse/conditional-relops/),
since none of the four full example programs happens to use them).

## 2. Compiled shape: `Conditional<Member>`

Unlike every construct in `groups.md`, a conditional's `cond` reads
`state` — a per-athlete value, not known until `resolve` runs. So
structural compilation does **not** pick a branch or discard the other
one: it compiles *both* branches (a `stmt` compiles to a `Member`, per
`groups.md` §3) and wraps them in one new IR node:

```
Conditional<Member> { cond: BoolExpr, then: Member, else: Member }
```

`if-then-else.owl`'s

```owl
if tm.variable < 15 then:
    superset(bench, row);
else
    bench;
```

compiles to a `Conditional{ cond: {type: "comparison", op: "<", left:
path(tm.variable), right: number(15)}, then: Group{kind: superset,
...}, else: Ref{name: "bench"} }` sitting in the enclosing day's
`body.members` at the position the `if` appeared — exactly where an
ordinary `stmt` would sit. Carries the literal `"type": "conditional"`
tag — see [`canonical-form.schema.json`](../canonical-form.schema.json)'s
`conditionalMember` def.

**Note on `exerciseItem`-level conditionals.** Before `progress` became
a single code block per exercise ([`progression.md`](./progression.md)),
the one attested use of a conditional at this level was choosing between
two whole `progressDecl`s. That's gone: a `progressBody` has its own
`if`/`then`/`else` (`progressionIf`, §2a) for exactly this, so there's no
longer a reason to wrap a whole exerciseItem in a `Conditional`. A
conditional `setDecl`/`dropsetDecl` branch remains unattested (§4) — so
today the compiler rejects any `condExerciseItem` outright, rather than
picking a shape for a case nothing demonstrates.

## 2a. `BoolExpr`: comparisons, `and`, `or`

`BoolExpr` is shared verbatim between `Conditional.cond` here and
`progressionIf.cond` ([`progression.md`](./progression.md) §2) — the same
grammar `condExpr`, the same compiled shape, used by two different
consumers (`resolve` for one, `progress` for the other):

```
BoolExpr = Comparison | And | Or
Comparison { op: '<' | '>' | '<=' | '>=' | '==' | '!=', left: Expr, right: Expr }
And        { left: BoolExpr, right: BoolExpr }
Or         { left: BoolExpr, right: BoolExpr }
```

`and` binds tighter than `or` (standard precedence); parenthesize
(`condAtom ::= '(' condExpr ')'`) to mix them unambiguously. Both are new
— no attested example combines conditions yet, but the grammar and
compiler support them uniformly with the base relational case.

## 3. Resolution: this is where the branch is picked

[`resolve`](./resolution.md) walks the canonical tree exactly as §2 of
that document describes, with one added rule, run wherever a
`Conditional` node is encountered in a `members` array:

1. Resolve `cond` recursively: for a `Comparison`, resolve `left`/`right`
   as ordinary `Expr`s (resolution.md §2 steps 1-4 — a `state` path may
   already have progressed a real value, a catalog field may need a
   lookup, a fallback may apply) and evaluate `left <op> right`; for
   `And`/`Or`, resolve both sides the same way and combine with the
   ordinary boolean operator.
2. Replace the `Conditional` node with its `then` branch if `cond` is
   true, its `else` branch otherwise — recursively resolved the same way
   an inlined `Ref` is (resolution.md §2's last bullet). The branch *not*
   taken is dropped; a `Session` handed to the app never contains a
   `Conditional` node, just as it never contains a `Ref`.

This is a pure function of `(cond, state)` for a given cursor, so it
inherits the same idempotence guarantee resolution.md §3 already states
for the rest of `resolve`: re-resolving the same `(program, state)` picks
the same branch every time, and `state` only changes via `progress`
running after a session is logged — never mid-`resolve`.

A `progressionIf` inside a `progress` block ([`progression.md`](./progression.md)
§2) is evaluated the same way, but by `progress`, against that session's
log and the athlete's pre-update `state` — never by `resolve`, and never
turned into a `Conditional<Member>`; it's a different consumer of the
same `BoolExpr` shape, not an instance of this section's mechanism.

## 4. Open questions (not yet attested in `conformance/programs/`)

Flagging these here rather than asserting them as canon, per this
project's convention (compare `groups.md` §5):

- **A `setDecl` conditional branch** — `condExerciseItem`'s branches parse
  as any `exerciseItem`, so `if cond then: set _ = 10 @ 100kg else set _ =
  8 @ 100kg` is grammatically valid, but no example attests it, this
  document doesn't define its canonical-form shape, and the compiler
  rejects `condExerciseItem` outright today (§2's note) — including this
  case. A conditional `Member` can't represent it directly (a `set` line
  isn't a `Member`, it's an entry in `ExerciseDecl.sets`/`body.members`
  alongside other sets), so it would likely need its own
  `Conditional<SetRef>` variant rather than reusing `conditionalMember`.
- **`else`-less conditionals** — every attested example at the
  `dayItem`/`blockItem`/`topLevelItem` levels has an `else` branch;
  there's no attested meaning for "do nothing" when `cond` is false at
  those levels (would `Conditional.else` become optional, or is a no-op
  expressed some other way — an empty `Group`?). `progressionIf`, by
  contrast, already has an attested-by-design optional `else` — see
  `progression.md` §2.
- **Nested/chained conditionals (`else if`)** — attested for
  `progressionIf` by construction (§2a), but not yet for
  `Conditional<Member>`'s four levels; an `else` branch that is itself a
  conditional item parses fine there too (branches recurse through the
  same item nonterminal) but no `conformance/programs/` example
  demonstrates it, so the multi-way-branch reading is unconfirmed at
  those levels.
