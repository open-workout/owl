# OWL — Open Workout Language

OWL is a domain-specific language for describing workout programs — from
a plain 5x5 squat routine to multi-part CrossFit metcons with EMOMs,
AMRAPs, and for-time couplets — as structured, executable data rather than
prose. A compiled OWL program is a self-contained description of what to
do, how it's grouped and timed, and how it should progress over time,
built to be picked up and run by an app (the reference target is the
OpenWorkout mobile app).

```owl
state = {
    tm.squat.weight = 0.85 * $BarbellBackSquat.e1rm | AMW
    tm.squat.reps   = 5
}

units  = "kg"
plates = [25, 20, 15, 10, 5, 2.5, 1.25]

block leader = {
    day SE_lower = {
        exercise squat = $BarbellBackSquat {
            set top_set = tm.squat.reps+ @ tm.squat.weight
            progress    = double(top_set, 8, 12, 5)
        }
        squat;
    }
    SE_lower;
}

leader*5;
```

Every group-like construct — straight sets, supersets, circuits, EMOMs,
AMRAPs, for-time — compiles down to one IR node with four knobs
(`interleave`, `rest`, `termination`, `atomic`); see
[`spec/semantics/groups.md`](./spec/semantics/groups.md) for the full
picture.

## Where to go next

- **New to OWL?** Start with [`spec/README.md`](./spec/README.md) — a
  reading guide through the grammar, canonical form, and semantics, in
  the order that actually makes sense.
- **Want to see real programs?** [`conformance/programs/`](./conformance/programs/)
  has a plain progressive-overload program, a superset day, and a
  multi-part CrossFit session, each annotated with what it demonstrates.
- **Want to try it interactively?** The web playground lives under
  [`tools/playground/`](./tools/playground/) (not yet built — see that
  directory's README).
- **Want to contribute?** See [`CONTRIBUTING.md`](./CONTRIBUTING.md) —
  language changes go through a lightweight proposal process since they
  affect every implementation and every existing `.owl` program.

## Repository layout

| Path | What's there |
|---|---|
| [`spec/`](./spec/) | The source of truth: grammar, canonical form, semantics. Prose, not code. |
| [`conformance/`](./conformance/) | Implementation-independent test corpus: fixtures and real programs. |
| [`reference/`](./reference/) | The reference implementation. |
| [`bindings/`](./bindings/) | Per-language wrappers around the reference implementation. |
| [`examples/`](./examples/) | Real, runnable `.owl` programs, for reading rather than testing. |
| [`docs/`](./docs/) | User-facing tutorials and how-tos. |
| [`tools/playground/`](./tools/playground/) | Web REPL for pasting and running OWL programs. |

## License

Dual-licensed under [MIT](./LICENSE-MIT) or [Apache-2.0](./LICENSE-APACHE),
at your option — see [`LICENSE`](./LICENSE).
