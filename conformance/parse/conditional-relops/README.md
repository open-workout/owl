Attests the three `relOp` alternatives
([`spec/grammar.ebnf`](../../../spec/grammar.ebnf)) beyond `<`/`>` that
none of the full `conformance/programs/*.owl` examples happens to use:
`<=`, `>=`, and `==`. Three independent `dayItem`-level conditionals
(same shape as
[`../conditional-exercise-choice/`](../conditional-exercise-choice/), one
per operator) sit back-to-back in `day1`'s body, so
`day1.body.termination` is `count(3)` rather than `count(1)` — three
source-level `if`/`then`/`else` items, three `Conditional` members, each
with the corresponding `cond.op` and both branches preserved unevaluated
(see [`spec/semantics/conditionals.md`](../../../spec/semantics/conditionals.md)
§2). `!=` remains unattested — see that document's §4.
