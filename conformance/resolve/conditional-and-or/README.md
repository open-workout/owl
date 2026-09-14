Corresponds to compiling
[`conformance/programs/conditional-and-or.owl`](../../programs/conditional-and-or.owl).

Exercises `and`/`or` combined in one `Conditional.cond`
(`spec/semantics/conditionals.md` §2, `spec/grammar.ebnf`'s `condOr`/
`condAnd`) — the compiled `cond` is
`or(and(weight>=90, readiness>=7), fatigue>8)`, matching `and` binding
tighter than `or`. `state.json` is deliberately chosen so the two
possible (wrong) precedences disagree: `tm.squat.weight` is `80` (fails
`>=90`), `readiness` is `5` (fails `>=7`), `fatigue` is `9` (passes
`>8`).

- With the correct precedence (`and` tighter): `or(and(false, false),
  true)` = **true** → `squat_heavy` runs.
- With the opposite (wrong) precedence, `(weight>=90) and (readiness>=7
  or fatigue>8)`: `and(false, or(false, true)=true)` = **false** →
  `squat_light` would run instead.

So this fixture's `expected.json` picking `squat_heavy` is a real
assertion about precedence, not just about `and`/`or` parsing at all.

Two things worth noting about the shape of `expected.json`:

- **`exercises[]` keeps both `squat_heavy` and `squat_light`**, each
  fully resolved (their own `target`/`load` `Expr`s walked and resolved
  independently of which one the `Conditional` picked) — `exercises[]`
  is the day's flat declared-exercise list (mirrors `groups[]` staying
  present post-resolve, see
  [`../emom-amrap-groups/`](../emom-amrap-groups/)'s README), not a
  reflection of what's actually prescribed today.
- **`body` only ever contains `squat_heavy`.** Per
  `resolution.md` §2 and `conditionals.md` §3, the `Conditional` member
  is replaced by its resolved `then` branch (`Ref(squat_heavy)`,
  inlined — nested, not flattened, same rule as
  `../superset-circuit-groups/`); the `else` branch
  (`Ref(squat_light)`) is dropped and never appears anywhere in `body`.
  `body` is the authoritative "what actually happens" tree;
  `exercises[]`/`groups[]` are reference metadata.
