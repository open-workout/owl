Corresponds to compiling:

```owl
state = {
    tm.squat.weight = 0.85 * $BarbellBackSquat.e1rm | 60
    tm.squat.reps = 8
}
units = "kg"
plates = [25, 20, 15, 10, 5, 2.5, 1.25]

block main = {
    day day1 = {
        exercise squat = $BarbellBackSquat {
            set top_set = tm.squat.reps @ tm.squat.weight
            progress = double(top_set, 8, 12, 5)
        }
        squat;
    }
}
```

`state.json` has no recorded `BarbellBackSquat.e1rm` and no stored
`tm.squat.weight` binding, so `resolve` must take the numeric fallback
path: `tm.squat.weight` resolves to `60` (the fallback value itself, not
`0.85 * 60` — see `resolution.md` step 4: fallback replaces the whole
expression's value, not just the missing leaf). `tm.squat.reps` resolves
from its own literal `expr` (`8`) since it needs no lookup at all.

`expected.json` also exercises `Ref` inlining (`resolution.md` §2): both
`day1`'s `body` (`Ref("squat")`) and `squat`'s own body resolve down to
the same concrete `top_set` `SetRef` — a conforming `resolve` must not
leave any `{"type":"ref",...}` node in its output.
