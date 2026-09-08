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
            if tm.squat.weight > 150 then
                progress = double(top_set,8,12,2.5)
            else progress = double(top_set,8,12,5)
        }
        squat;
    }
}
```

`state.json` has a stored `tm.squat.weight` of `100`, so per
[`resolution.md`](../../../spec/semantics/resolution.md) §2 step 2 that
value is used directly (no formula to evaluate here since the binding's
own `expr` is already a plain number). Resolving `top_set`'s
`progression` — a `Conditional`
([`conditionals.md`](../../../spec/semantics/conditionals.md) §3) rather
than a plain `progressionRule` — evaluates `cond`: `100 > 150` is false,
so the **`else`** branch (`double(top_set,8,12,5)`) is the one that
survives into `expected.json`; the `then` branch
(`double(top_set,8,12,2.5)`) and the `Conditional`/`cond` wrapper itself
are both gone. See
[`../../parse/conditional-progression/`](../../parse/conditional-progression/)
for the same program's structural (pre-resolve) canonical form, with
both branches still present.
