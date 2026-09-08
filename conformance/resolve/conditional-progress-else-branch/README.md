Corresponds to compiling:

```owl
state = {
    tm.squat.weight = 100
}
units = "kg"
plates = [25, 20, 15, 10, 5, 2.5, 1.25]

block main = {
    day day1 = {
        exercise squat = $BarbellBackSquat {
            set top_set = 5 @ tm.squat.weight
            progress = {
                if tm.squat.weight > 150 then
                    tm.squat.weight = tm.squat.weight + 2.5
                else
                    tm.squat.weight = tm.squat.weight + 5
            }
        }
        squat;
    }
}
```

`state.json` has a stored `tm.squat.weight` of `100`, so per
[`resolution.md`](../../../spec/semantics/resolution.md) §2 step 2 that
value is used directly. The interesting case this fixture exercises:
`squat`'s `progress` block — including its own `if`/`then`/`else` — is
**not** touched by `resolve` at all. Per `resolution.md`'s note and
[`progression.md`](../../../spec/semantics/progression.md) §2,
`progress`'s `if` is evaluated by `progress()` against a session log
that doesn't exist yet at resolve time, never by `resolve`. So
`expected.json`'s `exercises[0].progress` is **byte-identical** to
`canonical.json`'s — same unresolved `path`/`comparison` nodes, neither
branch dropped — while `target`/`load` elsewhere in the same document
resolve to plain numbers as usual. See
[`../../parse/conditional-progression/`](../../parse/conditional-progression/)
for the same shape at parse time.
