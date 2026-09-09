Demonstrates the single-owner rule
([`spec/semantics/parameters.md`](../../../spec/semantics/parameters.md)
§2): `squatA` and `squatB` each have their own `progress` block, and both
assign `tm.squat.weight` — so two different exercises would write the
same state path. A conforming implementation must reject this program
rather than pick one arbitrarily.
