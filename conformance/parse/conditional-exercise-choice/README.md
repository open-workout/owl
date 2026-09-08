Demonstrates a `dayItem`-level conditional
([`spec/semantics/conditionals.md`](../../../spec/semantics/conditionals.md)
§2, `Conditional<Member>`), minimized from
[`if-then-else.owl`](../../programs/if-then-else.owl): `if tm.variable <
15 then: squat; else bench;` compiles to a single `Conditional` node
sitting in `day1.body.members` — both branches (`Ref("squat")` and
`Ref("bench")`) are preserved unevaluated, since `cond` reads `state` and
picking one is `resolve`'s job, not structural compilation's. Note
`day1.body.termination` is `count(1)`: the `if`/`then`/`else` is one
source-level item, so it contributes exactly one member, same as any
other single `stmt`.
