# spec/

This directory is OWL's source of truth: prose and formal grammar, not
code. Nothing here executes — [`reference/`](../reference/) is the
implementation that does, and [`conformance/`](../conformance/) is the
test corpus that checks a given implementation against this spec.

## Reading order

If you're new to OWL, read in this order:

1. [`grammar.ebnf`](./grammar.ebnf) — the surface syntax. Every production
   is backed by an example in [`conformance/programs/`](../conformance/programs/).
2. [`canonical-form.md`](./canonical-form.md) (+ [`canonical-form.schema.json`](./canonical-form.schema.json))
   — what a `.owl` program compiles to.
3. `semantics/`, in order:
   - [`groups.md`](./semantics/groups.md) — the one IR node every
     superset/circuit/EMOM/AMRAP/for-time construct reduces to.
   - [`targets-loads.md`](./semantics/targets-loads.md) — what a `set`
     line's two `quantity` slots actually mean.
   - [`parameters.md`](./semantics/parameters.md) — the `state`
     namespace, the single-owner rule, onboarding sentinels (`AMW`).
   - [`resolution.md`](./semantics/resolution.md) — turning a compiled
     program into one athlete's concrete session.
   - [`progression.md`](./semantics/progression.md) — turning a logged
     session back into updated `state`.
4. [`stdlib/schemes.md`](./stdlib/schemes.md) — the concrete progression
   algorithms (`double`, …) that `progression.md` defines the contract for.
5. [`versioning.md`](./versioning.md) — how the spec and canonical form
   version.

## Where to look for X

| Question | Document |
|---|---|
| "Does `X` parse?" | `grammar.ebnf` |
| "What does `superset(...)` actually mean?" | `semantics/groups.md` |
| "What's the difference between reps and load?" | `semantics/targets-loads.md` |
| "How does a training max get updated after a workout?" | `semantics/progression.md`, `stdlib/schemes.md` |
| "What happens for a brand-new athlete with no history?" | `semantics/parameters.md` § Onboarding |
| "What JSON does the app actually consume?" | `canonical-form.md`, `canonical-form.schema.json` |
| "Is this construct real or proposed?" | Check the construct's doc for a "Proposed extensions" section — anything not backed by `conformance/programs/` is flagged, not asserted as canon. |

## Proposing a change

See [`../CONTRIBUTING.md`](../CONTRIBUTING.md) for the OWL-EP process. In
short: a spec change should touch `grammar.ebnf` (if it's new syntax),
the relevant `semantics/*.md`, `canonical-form.schema.json` (if the IR
shape changes), and add a fixture under `conformance/` — in that order, so
each layer's claims are backed by the one below it.
