# examples/

Real, runnable `.owl` programs, meant for people learning the language —
not test fixtures (see [`../conformance/`](../conformance/) for those).

- [`fran.owl`](./fran.owl) — the CrossFit benchmark "Fran"
  (`for_time { rounds [21, 15, 9] as n { ... } }`), extracted as a
  standalone program from the larger multi-part session in
  [`../conformance/programs/complicated-crossfit.owl`](../conformance/programs/complicated-crossfit.owl).

More programs (a 5/3/1-style block, GZCLP, …) are planned but not yet
written — each one should be a real, correctly-transcribed program rather
than an invented approximation, so they're added deliberately rather than
generated in bulk.
