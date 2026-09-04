Demonstrates the single-owner rule
([`spec/semantics/parameters.md`](../../../spec/semantics/parameters.md)
§2): `squatA`'s set `a` and `squatB`'s set `b` both read
`tm.squat.weight` and both carry a `progress` rule, so two rules would
write the same state path. A conforming implementation must reject this
program rather than pick one rule arbitrarily.
