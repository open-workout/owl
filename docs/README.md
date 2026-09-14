# docs/

User-facing documentation: tutorials and how-tos for people *writing*
OWL programs, as opposed to [`spec/`](../spec/) (the formal definition,
for implementers) or [`examples/`](../examples/) (programs to read).

One implementer-facing reference lives here too, since it's consumption
docs rather than language-defining prose:
[`canonical-form-reference.md`](./canonical-form-reference.md) — a
self-contained wire-format guide to the compiled JSON (`Program`) and
resolved JSON (`Session`) shapes, written for whoever builds a client
(e.g. the mobile app) that deserializes them.

The tutorials below are **not written yet.** Planned outline:

- **Getting started** — install/open the playground, write and run your
  first program.
- **Writing your first program** — `state`, `units`/`plates`,
  `block`/`day`/`exercise`, a single progressing lift.
- **Groups** — supersets, circuits, EMOMs, AMRAPs, for-time; when to
  reach for each.
- **Progression** — how `state` and `progress` fit together, what
  happens for a first-time athlete (onboarding sentinels like `AMW`).
- **Authoring a real program** — walking through one of
  `conformance/programs/` end to end.
