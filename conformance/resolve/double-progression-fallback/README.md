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
            progress = {
                if top_set.reps >= 12 then
                    tm.squat.weight = tm.squat.weight + 5
                    tm.squat.reps = 8
                else if top_set.reps >= 8 then
                    tm.squat.reps = tm.squat.reps + 1
            }
        }
        squat;
    }
}
```

(`progress` isn't resolve()'s concern — see
[`../conditional-progress-else-branch/`](../conditional-progress-else-branch/)
for the fixture demonstrating it passes through unresolved. It's kept
here to make this a realistic program; nothing about it affects what
this fixture actually asserts.)

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
