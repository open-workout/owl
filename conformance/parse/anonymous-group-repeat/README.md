Demonstrates the anonymous-body form of the repeat macro
([`spec/semantics/groups.md`](../../../spec/semantics/groups.md) §3.7a,
`spec/grammar.ebnf`'s `repeatable` production): `(squat, rest(1m))*3`
desugars to the same `Group{kind:"straight", termination:{mode:"count","n":3}}`
shape `name*N` (§3.7) produces, just wrapping an inline statement list
instead of a declared name's body. Also exercises the "a bare `rest(d)`
statement sets its enclosing sequential group's `rest` to
`{single, duration: d}`" rule (§3.7a) — each of the 3 per-iteration
groups carries `rest: {mode: "single", duration: {value: 1, unit: "m"}}`
rather than `rest(1m)` appearing as a member of its own.
