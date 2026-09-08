# Canonical form

The canonical form is OWL's interchange target: the JSON shape a `.owl`
source file compiles to, and the shape every conformance fixture, the
reference implementation, and the app runtime agree on. Full field-level
detail is in [`canonical-form.schema.json`](./canonical-form.schema.json)
(JSON Schema, draft 2020-12); this document is the prose map of it.

## Shape at a glance

```
Program {
  owlVersion: "0.1.0"
  units:  "kg" | "lb"
  plates: number[]
  state:  StateBinding[]
  blocks: Block[]
  sequence?: Group          // top-level ordering/repetition, e.g. `leader*5;`
}

Block { name, days: Day[], body: Group }       // body sequences the days
Day   { name, exercises: ExerciseDecl[], groups: NamedGroup[], body: Group }

ExerciseDecl { name, catalog, sets: SetRef[], groups: NamedGroup[], body: Group }
                                                // one `exercise` decl; `groups` holds
                                                // named `dropset` declarations (semantics/groups.md §3.8)
NamedGroup   { name, group: Group }            // one `partA = <groupExpr>` assignment

Group { kind, members: Member[], interleave, rest, termination, atomic }
Member = SetRef | Ref | Group | Conditional<Member>   // groups nest
Ref    { name }                                // a bare-name use: exercise, named
                                                // group, day, or block — resolved
                                                // against the enclosing declarations

SetRef { exercise, target, load?, progression? }
Target = Reps | Distance | Duration           // what to do
Load   = Weight                               // what to do it against
progression: ProgressionRule | Conditional<ProgressionRule>

Conditional<T> { cond: BoolExpr, then: T, else: T }   // `if cond then: A else B`;
                                                       // see semantics/conditionals.md
```

A `Block`/`Day`'s `body` is always present, even when the source has no
explicit sequencing statements after its declarations (e.g.
`complicated-crossfit.owl`'s `block main` has a single `day` and nothing
else). In that case the compiler synthesizes `body = Group{ interleave:
sequential, termination: count(n), members: [Ref(d) for each declared day
in source order] }` — a Block/Day's `body` is never itself optional, only
the source's explicit statement list is.

`ExerciseDecl.body` is synthesized the same way, in declaration order —
except its members are the `sets`/`groups` themselves, embedded directly,
never a `Ref` (there's no bare-statement syntax to invoke a single `set`
or named `dropset` from outside its exercise, unlike a day/block name, so
nothing needs to be independently addressable there).

## Two passes, two documents apart

Canonical form is produced by **structural compilation** — parsing a
`.owl` source and expanding every macro (`rounds … as x`, `name*N`) — which
is athlete-agnostic: the same canonical document works for every athlete
running the program. It is *not* the final, ready-to-perform workout: any
`target`/`load` whose `expr` is a dotted `state` path or a
`$Catalog.field` reference is left unresolved, along with its `fallback`
if it has one. `Conditional` is the one construct compilation does *not*
expand — its `cond` reads `state`, so both `then`/`else` branches are kept
as-is and `resolve` (not structural compilation) picks between them; see
[`semantics/conditionals.md`](./semantics/conditionals.md).

Turning one canonical document into a concrete, numbers-filled-in workout
for one athlete on one day is [`resolve`](./semantics/resolution.md)'s
job — a separate, per-athlete pass. See:

- [`semantics/groups.md`](./semantics/groups.md) — how source constructs
  (`superset`, `emom`, `amrap`, `for_time`, `rounds`, …) desugar into the
  `Group` shape above.
- [`semantics/targets-loads.md`](./semantics/targets-loads.md) — the
  `Target`/`Load` unions.
- [`semantics/conditionals.md`](./semantics/conditionals.md) — `if`/
  `then`/`else`, the `Conditional` shape above, and why it's resolved at
  `resolve` time rather than compile time.
- [`semantics/resolution.md`](./semantics/resolution.md) — `resolve`.
- [`semantics/progression.md`](./semantics/progression.md) and
  [`stdlib/schemes.md`](./stdlib/schemes.md) — how a logged session flows
  back into updated `state`.

## Versioning

Every canonical document declares `owlVersion`. See
[`versioning.md`](./versioning.md).
