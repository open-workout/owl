# Contributing to OWL

OWL is a language, not an application: its most important artifact is
[`spec/`](./spec/), and a change there affects every implementation and
every `.owl` program that already exists. This document describes how to
propose a language change (an "OWL-EP") and how to contribute to the
non-spec parts of the project (implementations, tooling, docs, examples).

## Non-spec contributions

Bug fixes to [`reference/`](./reference/) or the bindings, additions to
[`examples/`](./examples/), improvements to [`docs/`](./docs/) or
[`tools/playground/`](./tools/playground/), and new
[`conformance/`](./conformance/) fixtures for *already-spec'd* behavior
are all ordinary pull requests — open one, no proposal needed. A new
conformance fixture that only tightens coverage of existing, attested
syntax is always welcome.

## Proposing a language change (OWL-EP)

Anything that changes `grammar.ebnf`, adds or changes a construct's
meaning in `semantics/`, changes `canonical-form.schema.json`, or adds a
new standard-library progression scheme needs an OWL-EP (OWL Enhancement
Proposal) before landing. This exists because a spec change is much
harder to walk back than an implementation bug — every `.owl` program
that's ever been written may be relying on the current behavior.

### Process

1. **Open an issue** describing the motivation: what workout structure or
   authoring pattern can't be expressed today, or what's ambiguous/wrong
   in the current spec. Include at least one concrete `.owl` snippet
   showing what you want to be able to write.
2. **Discuss.** A maintainer will tell you early if the direction seems
   right, wrong, or needs a different shape — before you've written a
   full proposal.
3. **Write the proposal** as a PR touching, in this order:
   - `spec/grammar.ebnf` — the new/changed production(s).
   - The relevant `spec/semantics/*.md` — what it compiles to and means.
     If it's a new group-like construct, it should fit the `Group` model
     in `semantics/groups.md` (or explain why it can't).
   - `spec/canonical-form.schema.json` — if the IR shape changes.
   - At least one fixture under `conformance/parse/` (and
     `conformance/resolve/`/`conformance/progress/` if the change touches
     resolution or progression) demonstrating the new behavior end to end.
   - A note in `CHANGELOG.md`.

   Each layer should be backed by the one below it — semantics claims
   should cite the grammar production, fixtures should exercise exactly
   what the semantics doc describes. A proposal that only adds prose
   without a fixture won't be merged.
4. **Status.** While under discussion the PR is Draft; a maintainer marks
   it Accepted (and it can be merged) or asks for changes. There's no
   separate proposal-tracking system yet — the PR *is* the proposal.

### What makes a strong OWL-EP

- It's grounded in a real program someone wants to write, not a
  hypothetical.
- It reuses existing concepts where it can (a new group-like construct
  should map onto `interleave`/`rest`/`termination`/`atomic`, not
  introduce a fifth knob) — see `semantics/groups.md` §5 for constructs
  already flagged as proposed-but-undecided; check there first.
- It comes with a fixture, not just prose — an untestable spec change is
  hard to review and hard to trust.

## Code of conduct

Participation in this project is governed by [`CODE_OF_CONDUCT.md`](./CODE_OF_CONDUCT.md).
