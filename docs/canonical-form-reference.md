# OWL Canonical JSON — Mobile Client Reference

This document is a self-contained wire-format reference for building a
client (mobile app, or any other consumer) that deserializes OWL's
compiled JSON. It assumes **no other context from the OWL repo** — every
type, enum, and example you need is inlined below.

## 1. Where this fits — read this before writing any code

OWL (Open Workout Language) is a DSL for workout programs. A backend
service compiles `.owl` source into JSON and hands it to the client in
one of two shapes, described fully below. **The client never parses
OWL source, never runs the compiler, and never runs the two "business
logic" passes (`resolve`, `progress`) described in §7-8.** All of that
happens server-side. The client's job is purely:

1. Deserialize the JSON (either shape) into native models.
2. Walk the `Group` tree to drive rendering/execution UI (§5, §9).
3. Collect what the athlete actually did and send it back as a
   `SessionLog` (§8.3) for the server to process.

There are exactly two JSON shapes a client will ever receive:

| Shape | Produced by | Athlete-specific? | Contains unresolved formulas? |
|---|---|---|---|
| **Program** (§2) | compiling `.owl` source | No — same for every athlete | Yes — `state`/catalog-field references, `Conditional` branches not yet picked |
| **Session** (§3) | resolving a `Program` against one athlete's current state, for one day | Yes | No — every number is concrete, except `progress` code (§7), which is carried but never evaluated by the client |

A **`Program`** is the whole compiled workout program (every block,
every day) — athlete-agnostic, useful for showing programme structure,
a calendar/plan overview, "what's coming up," etc. A **`Session`** is
one specific day, resolved for one specific athlete — this is "today's
workout," with every weight/rep/distance number already computed and
ready to render. Expect to consume `Session` for the actual
workout-execution screen, and (optionally) `Program` for
overview/planning screens.

Both shapes share almost all of their building blocks (`Group`,
`SetRef`, `Target`, `Load`, `Duration`...) — the difference is that in
a `Program`, `target`/`load` carry an `expr` (an unresolved formula)
and `Ref`/`Conditional` nodes can appear; in a `Session`, `target`/
`load` carry a plain `value` and every `Ref`/`Conditional` has already
been resolved away (see §3, §6.6).

## 2. `Program` — top level

```
Program {
  owlVersion: string          // e.g. "0.1.0" — see §11
  units:      "kg" | "lb"
  plates?:    number[]        // e.g. [25, 20, 15, 10, 5, 2.5, 1.25]
  state:      StateBinding[]
  blocks:     Block[]
  sequence?:  Group           // top-level ordering/repetition of blocks
}
```

- `owlVersion` — **check this before trusting the rest of the
  document's shape.** See §11.
- `plates` — the available plate sizes for plate-math (rounding a
  resolved weight to what's actually loadable). Optional.
- `state` — declarations of every dotted state path the program
  references (training maxes, etc.) and how to seed them. See §6.7.
  This is program-level metadata; an individual athlete's *current
  values* for these paths live in a separate `State` object (§8.1),
  not here.
- `blocks` — the program's training blocks, in declaration order (a
  "block" = e.g. a 5-week strength cycle, a deload week — whatever
  granularity the program author chose).
- `sequence` — if present, a `Group` (§5) describing how `blocks`
  themselves are ordered/repeated (e.g. `leader*5;` at the top level of
  the source). Uses the same repetition-macro mechanism as `rounds`
  (§5.9).

```
Block { name: string, days: Day[], body: Group }
Day   { name: string, exercises?: ExerciseDecl[], groups?: NamedGroup[], body: Group }
```

- `Block.body` / `Day.body` sequence that block's `days` / that day's
  top-level statements (exercises, supersets, EMOMs, rest, ...) — see
  §5. Always present, even if the source had no explicit sequencing
  (the compiler synthesizes one member per declared day/exercise in
  source order).
- `Day.exercises` — every `exercise` declared directly in this day
  (flat list; NOT nested inside a superset/circuit — those reference
  exercises by name, see `Ref` in §6.6).
- `Day.groups` — named group assignments made directly in this day,
  e.g. `partA = 5*emom(3m){ ... }`. Referenced elsewhere in `body` by
  `Ref{name:"partA"}`.

```
ExerciseDecl {
  name:      string            // local alias, e.g. "back_squat"
  catalog:   string            // exercise-library id, e.g. "BarbellBackSquat"
  sets:      SetRef[]          // flat list of this exercise's own top-level sets
  groups?:   NamedGroup[]      // this exercise's named dropset/restpause declarations
  body:      Group             // this exercise's own set sequence (see §5.1)
  progress?: ProgressionBody   // see §7 — NEVER evaluated client-side
}

NamedGroup { name: string, group: Group }
```

- `name` is the local identifier other parts of the program use to
  refer to this exercise (e.g. as a `superset(...)` argument, or a
  bare `back_squat;` statement) — distinct from `catalog`, the actual
  exercise-library entry (icon, instructions, muscle groups, etc. live
  keyed by `catalog`, not `name`).
- `sets` lists only this exercise's flat, non-grouped sets. Sets that
  live inside one of this exercise's own named `groups` (a dropset,
  restpause group) are *not* duplicated here — they're only in that
  `NamedGroup`'s own `group.members`.
- `body` embeds `sets`/`groups` directly (never via `Ref`) in
  declaration order — this is "what running this exercise on its own
  looks like."

## 3. `Session` — top level

This is what a `resolve()` call produces for one athlete, one day —
the shape the workout-execution screen should render.

```
Session {
  name:        string             // the day's name
  exercises?:  ExerciseDecl[]     // same shape as Program's, but target/load are resolved (§3.1)
  groups?:     NamedGroup[]       // same shape as Program's, group is resolved
  body:        Group              // resolved — no Ref, no Conditional (see §6.6)
}
```

Structurally identical to a `Program`'s `Day`, with these differences:

- Every `target`/`load`'s `expr` has been replaced by a plain `value`
  (§3.1) — a `Session` never contains an `expr` node **except** inside
  `exerciseDecl.progress`, which is carried completely unresolved (§7)
  — the client should treat `progress` as opaque, display-only data,
  never evaluate it.
- Every `Ref` member has been inlined — replaced by the *whole*
  resolved `Group`/`ExerciseDecl.body` it named, nested at that
  position (not flattened — see §6.6). A `Session`'s `body` (and every
  nested group) never contains a `{"type":"ref",...}` node anywhere.
- Every `Conditional` member has been replaced by whichever of its
  `then`/`else` the server picked; the branch not taken is gone
  entirely — a `Session` never contains a `{"type":"conditional",...}`
  node anywhere.
- `exercises`/`groups` still list **every** declared exercise/named
  group in the day, even ones not reachable from the (now-pruned)
  `body` because a `Conditional` picked the other branch — `body` is
  the authoritative "what to actually do today" tree; `exercises`/
  `groups` are reference metadata (useful for looking up a set's own
  declared context, e.g. rendering an exercise's demo video by
  `catalog` even if it's inside a nested group).

### 3.1 Resolved `target`/`load` shape

Same tags as the `Program` shapes (§6.3-6.4), `expr` → `value`, no
`fallback` field (it's already been applied or wasn't needed):

```
// Program (unresolved):
{ "kind": "reps", "expr": { "type": "path", "path": "tm.squat.reps" } }

// Session (resolved):
{ "kind": "reps", "value": 5 }
```

## 4. Minimal worked example (straight sets)

**Program** (compiled from a simple `set top = 5 @ tm.squat.weight`):

```json
{
  "owlVersion": "0.1.0",
  "units": "kg",
  "plates": [25, 20, 15, 10, 5, 2.5, 1.25],
  "state": [
    { "path": "tm.squat.weight", "expr": { "type": "number", "value": 100 } }
  ],
  "blocks": [
    {
      "name": "main",
      "days": [
        {
          "name": "day1",
          "exercises": [
            {
              "name": "squat",
              "catalog": "BarbellBackSquat",
              "sets": [
                {
                  "type": "setRef",
                  "label": "top_set",
                  "exercise": "BarbellBackSquat",
                  "target": { "kind": "reps", "expr": { "type": "number", "value": 5 } },
                  "load": { "kind": "weight", "expr": { "type": "path", "path": "tm.squat.weight" }, "unit": "kg" }
                }
              ],
              "groups": [],
              "body": {
                "type": "group", "kind": "straight",
                "interleave": "sequential", "rest": { "mode": "single" },
                "termination": { "mode": "count", "n": 1 }, "atomic": false,
                "members": [
                  {
                    "type": "setRef", "label": "top_set", "exercise": "BarbellBackSquat",
                    "target": { "kind": "reps", "expr": { "type": "number", "value": 5 } },
                    "load": { "kind": "weight", "expr": { "type": "path", "path": "tm.squat.weight" }, "unit": "kg" }
                  }
                ]
              }
            }
          ],
          "groups": [],
          "body": {
            "type": "group", "kind": "straight",
            "interleave": "sequential", "rest": { "mode": "single" },
            "termination": { "mode": "count", "n": 1 }, "atomic": false,
            "members": [ { "type": "ref", "name": "squat" } ]
          }
        }
      ],
      "body": {
        "type": "group", "kind": "straight",
        "interleave": "sequential", "rest": { "mode": "single" },
        "termination": { "mode": "count", "n": 1 }, "atomic": false,
        "members": [ { "type": "ref", "name": "day1" } ]
      }
    }
  ]
}
```

**Session** (resolved for an athlete whose `tm.squat.weight` is stored
as `100`; note `body`'s `Ref{name:"squat"}` inlined to squat's own
resolved `body` group, nested one level, per §6.6):

```json
{
  "name": "day1",
  "exercises": [
    {
      "name": "squat",
      "catalog": "BarbellBackSquat",
      "sets": [
        {
          "type": "setRef", "label": "top_set", "exercise": "BarbellBackSquat",
          "target": { "kind": "reps", "value": 5 },
          "load": { "kind": "weight", "value": 100, "unit": "kg" }
        }
      ],
      "groups": [],
      "body": {
        "type": "group", "kind": "straight",
        "interleave": "sequential", "rest": { "mode": "single" },
        "termination": { "mode": "count", "n": 1 }, "atomic": false,
        "members": [
          {
            "type": "setRef", "label": "top_set", "exercise": "BarbellBackSquat",
            "target": { "kind": "reps", "value": 5 },
            "load": { "kind": "weight", "value": 100, "unit": "kg" }
          }
        ]
      }
    }
  ],
  "groups": [],
  "body": {
    "type": "group", "kind": "straight",
    "interleave": "sequential", "rest": { "mode": "single" },
    "termination": { "mode": "count", "n": 1 }, "atomic": false,
    "members": [
      {
        "type": "group", "kind": "straight",
        "interleave": "sequential", "rest": { "mode": "single" },
        "termination": { "mode": "count", "n": 1 }, "atomic": false,
        "members": [
          {
            "type": "setRef", "label": "top_set", "exercise": "BarbellBackSquat",
            "target": { "kind": "reps", "value": 5 },
            "load": { "kind": "weight", "value": 100, "unit": "kg" }
          }
        ]
      }
    ]
  }
}
```

## 5. `Group` — the one node every workout structure compiles to

**Every** sequencing/interleaving construct — straight sets,
supersets, circuits, dropsets, rest-pause, EMOMs, AMRAPs, for-time,
rep-scheme rounds — compiles to the *same* node shape, distinguished
only by `kind` and four "knobs":

```
Group {
  type:        "group"
  kind:        "straight" | "superset" | "circuit" | "dropset" | "restpause"
             | "emom" | "amrap" | "for_time" | "rounds"
  members:     Member[]
  interleave:  "sequential" | "round_robin"
  rest:        RestPolicy
  termination: Termination
  atomic:      boolean
}
```

- **`interleave`**: `"sequential"` = finish member 0 entirely, then
  member 1, etc. `"round_robin"` = members take turns, one "lap" each,
  in list order (a superset: do bench, do row, repeat — not do all of
  bench's sets then all of row's).
- **`rest`**: when/how long to rest at each transition — see §5's
  `RestPolicy` table below.
- **`termination`**: when this group ends — see the `Termination` table
  below.
- **`atomic`**: `true` means this group is one indivisible unit from an
  *outer* group's point of view — e.g. a dropset's three descending
  sets are always run back-to-back even if the dropset itself sits
  inside a larger circuit; the outer circuit's own interleaving never
  splits the dropset's members apart.

`Member` (an entry in `Group.members`) is a **tagged union**:

```
Member =
  | SetRef                              // a single set to perform (§6.1)
  | Ref            { type: "ref", name }           // Program only — never in a Session
  | Group                                          // a nested group (groups nest freely)
  | Conditional<Member> { type: "conditional", ... }  // Program only — never in a Session
```

### `RestPolicy` — tagged by `mode`

| `mode` | Fields | Meaning |
|---|---|---|
| `"single"` | `duration?: Duration` | Flat rest before the next member (e.g. rest 90s between straight sets). `duration` omitted = no prescribed rest (athlete-paced). |
| `"two_level"` | `intra: Duration`, `inter: Duration` | `intra` = rest between members *within* one lap (e.g. bench→row, 20s); `inter` = rest between laps (e.g. finishing a full round, 90s). Used by `superset`/`circuit`. |
| `"ad_lib"` | *(none)* | No prescribed rest at all — athlete-paced (AMRAP, for-time). |
| `"remainder"` | *(none)* | Whatever time is left in a fixed interval after the work is done (EMOM). |

```json
{ "mode": "single" }
{ "mode": "single", "duration": { "value": 90, "unit": "s" } }
{ "mode": "two_level", "intra": { "value": 20, "unit": "s" }, "inter": { "value": 90, "unit": "s" } }
{ "mode": "ad_lib" }
{ "mode": "remainder" }
```

### `Termination` — tagged by `mode`

| `mode` | Fields | Meaning |
|---|---|---|
| `"count"` | `n: integer` | Fixed number of sets/rounds/drops/bursts. |
| `"emom"` | `interval: Duration`, `n: integer` | Every `interval`, for `n` total intervals. |
| `"time_cap"` | `cap: Duration` | Stop when the clock hits `cap` (AMRAP) — record completed rounds + partial reps. |
| `"for_time"` | *(none)* | No cap — the clock runs and elapsed time *is* the result. |

```json
{ "mode": "count", "n": 3 }
{ "mode": "emom", "interval": { "value": 3, "unit": "m" }, "n": 5 }
{ "mode": "time_cap", "cap": { "value": 12, "unit": "m" } }
{ "mode": "for_time" }
```

### `Duration`

```
Duration { value: number, unit: "s" | "m" | "h" | "d" }
```
`"m"` here always means **minutes** (never meters — a `Duration` is
never used for distance; see `Target.distance`'s separate `unit` enum
in §6.3, which uses `"m"`/`"km"` for meters/kilometers instead. The two
never appear in the same field, so there's no runtime ambiguity — only
know which one a given field is by its type in the tables above).

### 5.1-5.9 — every `kind`, fully worked

Every row below shows the exact `Group` shape (minus `members`, which
depends on the exercises involved) for that construct.

**5.1 `straight`** — plain sets of one exercise, or "run this bare
name once" wrappers (a `Block`/`Day`/`ExerciseDecl`'s own `body` when
it's just sequencing its children with no special interleaving).
```json
{ "kind": "straight", "interleave": "sequential", "rest": { "mode": "single" }, "termination": { "mode": "count", "n": 3 }, "atomic": false }
```

**5.2 `superset`** / **5.3 `circuit`** — structurally identical; the
`kind` string is the only difference (surface-syntax label for UI,
e.g. "Superset" vs "Circuit" badge — no other semantic difference).
Round count is however many sets the member exercises each have (they
must agree). `intra` rest is overridable per-pair in source
(`superset(bench, rest(20s), row)`); `inter` defaults to 90s and has
no override syntax yet.
```json
{ "kind": "superset", "interleave": "round_robin",
  "rest": { "mode": "two_level", "intra": { "value": 20, "unit": "s" }, "inter": { "value": 90, "unit": "s" } },
  "termination": { "mode": "count", "n": 3 }, "atomic": false }
```
```json
{ "kind": "circuit", "interleave": "round_robin",
  "rest": { "mode": "two_level", "intra": { "value": 0, "unit": "s" }, "inter": { "value": 90, "unit": "s" } },
  "termination": { "mode": "count", "n": 2 }, "atomic": false }
```

**5.4 `dropset`** — a top set plus its descending-load drop sets, run
as one atomic unit (`atomic: true` — an outer group never splits these
apart). `rest.single`'s duration is always omitted (no rest between
drops).
```json
{ "kind": "dropset", "interleave": "sequential", "rest": { "mode": "single" }, "termination": { "mode": "count", "n": 3 }, "atomic": true }
```

**5.5 `restpause`** — a top set plus fixed-load rest-pause bursts, same
shape as `dropset` but `rest.single.duration` is `15s` (explicit-form
default) or whatever the source's `rest_pause(duration, ...)` sugar
specified, and each burst holds the *same* target/load as the top set
(no descending load).
```json
{ "kind": "restpause", "interleave": "sequential",
  "rest": { "mode": "single", "duration": { "value": 15, "unit": "s" } },
  "termination": { "mode": "count", "n": 3 }, "atomic": true }
```

**5.6 `emom`** — "every minute on the minute"-style clock cadence.
`termination.interval` is the interval length, `termination.n` is the
total number of intervals.
```json
{ "kind": "emom", "interleave": "round_robin", "rest": { "mode": "remainder" },
  "termination": { "mode": "emom", "interval": { "value": 1, "unit": "m" }, "n": 5 }, "atomic": false }
```

**5.7 `amrap`** — "as many rounds as possible" in a time cap.
```json
{ "kind": "amrap", "interleave": "round_robin", "rest": { "mode": "ad_lib" },
  "termination": { "mode": "time_cap", "cap": { "value": 10, "unit": "m" } }, "atomic": false }
```

**5.8 `for_time`** — two distinct shapes share this `kind`:

- **Bare** (no rep-scheme rounds): `interleave: "round_robin"`,
  `members` is the flat statement list directly.
  ```json
  { "kind": "for_time", "interleave": "round_robin", "rest": { "mode": "ad_lib" },
    "termination": { "mode": "for_time" }, "atomic": false }
  ```
- **Wrapping `rounds [...] as x`** (a decreasing/increasing rep scheme,
  e.g. CrossFit's "21-15-9"): the *outer* wrapper is
  `interleave: "sequential"` (each rep-count's lap must fully finish
  before the next starts) and its `members` are `kind: "rounds"`
  groups, one per value (§5.9).
  ```json
  { "kind": "for_time", "interleave": "sequential", "rest": { "mode": "ad_lib" },
    "termination": { "mode": "for_time" }, "atomic": false }
  ```

**5.9 `rounds`** — only ever appears as a member of the `rounds`-macro
`for_time` wrapper above (5.8's second shape) — never at the top level.
One of these per rep-scheme value, `round_robin` with exactly one lap
(so e.g. "21 thrusters, 21 pull-ups" reads as thruster-then-pullup, not
alternating single reps).
```json
{ "kind": "rounds", "interleave": "round_robin", "rest": { "mode": "ad_lib" },
  "termination": { "mode": "count", "n": 1 }, "atomic": false }
```
So a full "21-15-9" `for_time` looks like (shown resolved/`Session`-shaped,
with plain `value` fields, for readability — in a `Program` each `target`/
`load` would instead carry an unresolved `expr`, per §3.1/§6.3-6.4):
```json
{ "type": "group", "kind": "for_time", "interleave": "sequential", "rest": { "mode": "ad_lib" }, "termination": { "mode": "for_time" }, "atomic": false,
  "members": [
    { "type": "group", "kind": "rounds", "interleave": "round_robin", "rest": { "mode": "ad_lib" }, "termination": { "mode": "count", "n": 1 }, "atomic": false,
      "members": [
        { "type": "setRef", "exercise": "BarbellThruster", "target": { "kind": "reps", "value": 21 }, "load": { "kind": "weight", "value": 42.5, "unit": "kg" } },
        { "type": "setRef", "exercise": "PullUp", "target": { "kind": "reps", "value": 21 } }
      ] },
    { "type": "group", "kind": "rounds", "interleave": "round_robin", "rest": { "mode": "ad_lib" }, "termination": { "mode": "count", "n": 1 }, "atomic": false,
      "members": [
        { "type": "setRef", "exercise": "BarbellThruster", "target": { "kind": "reps", "value": 15 }, "load": { "kind": "weight", "value": 42.5, "unit": "kg" } },
        { "type": "setRef", "exercise": "PullUp", "target": { "kind": "reps", "value": 15 } }
      ] },
    { "type": "group", "kind": "rounds", "interleave": "round_robin", "rest": { "mode": "ad_lib" }, "termination": { "mode": "count", "n": 1 }, "atomic": false,
      "members": [
        { "type": "setRef", "exercise": "BarbellThruster", "target": { "kind": "reps", "value": 9 }, "load": { "kind": "weight", "value": 42.5, "unit": "kg" } },
        { "type": "setRef", "exercise": "PullUp", "target": { "kind": "reps", "value": 9 } }
      ] }
  ]
}
```

## 6. Shared leaf/expression types

### 6.1 `SetRef` — one set to perform

```
SetRef {
  type:     "setRef"
  label?:   string     // omitted for anonymous/inline sets (surface `_`)
  exercise: string      // catalog id, e.g. "BarbellBackSquat"
  target:   Target      // §6.3
  load?:    Load        // §6.4 — omitted entirely for a load-less set (cardio, bodyweight reps)
}
```

### 6.2 `Ref` — a bare-name invocation (**Program only**)

```
Ref { type: "ref", name: string }
```
Points at a declared exercise, named group, day, or block by its local
`name`. **Never appears in a `Session`** — always inlined by `resolve`
(§6.6).

### 6.3 `Target` — what the athlete is asked to do

Tagged by `kind`. In a `Program`, carries `expr` (§6.5) + optional
`fallback` (§6.7); in a `Session`, carries a plain `value`.

| `kind` | Fields (Program) | Fields (Session) |
|---|---|---|
| `"reps"` | `expr`, `plus?: bool`, `fallback?` | `value: number`, `plus?: bool` |
| `"distance"` | `expr`, `unit: "m"\|"km"`, `fallback?` | `value: number`, `unit: "m"\|"km"` |
| `"duration"` | `expr`, `unit: "s"\|"m"\|"h"`, `fallback?` | `value: number`, `unit: "s"\|"m"\|"h"` |

`plus` (only meaningful on `reps`) marks an open-ended target — "10 or
more" / "to failure" (e.g. a dropset's auto-generated bursts are
`{kind:"reps", value:1, plus:true}` — "at least 1, to failure").

```json
{ "kind": "reps", "value": 8, "plus": true }
{ "kind": "distance", "value": 500, "unit": "m" }
{ "kind": "duration", "value": 60, "unit": "s" }
```

A `SetRef` with a `target` and no `load` is valid and common (cardio,
bodyweight reps, an isometric hold). A `load` with no `target` never
happens.

### 6.4 `Load` — what the target is performed against

Tagged by `kind`. Today the **only** variant is `weight` — there is no
bodyweight-load or percent-of-1RM-as-its-own-kind marker; a set with no
resistance simply omits `load` entirely (see §6.3's last paragraph).

| `kind` | Fields (Program) | Fields (Session) |
|---|---|---|
| `"weight"` | `expr`, `unit: "kg"\|"lb"`, `fallback?` | `value: number`, `unit: "kg"\|"lb"` |

```json
{ "kind": "weight", "value": 102.5, "unit": "kg" }
```

### 6.5 `Expr` — unresolved arithmetic (**Program only**, and inside `progress`, §7)

Tagged by `type`, recursive:

| `type` | Fields | Meaning |
|---|---|---|
| `"number"` | `value: number` | A literal. |
| `"path"` | `path: string` | A dotted `state` path, e.g. `"tm.squat.weight"` — resolves to the athlete's current stored value for that path, or its seed formula/fallback if none exists yet. |
| `"catalogField"` | `catalog: string`, `field: string` | e.g. `$BarbellBackSquat.e1rm` — resolves to the athlete's most recent recorded value of that field for that catalog exercise. |
| `"mul"` / `"add"` / `"sub"` / `"div"` | `left: Expr`, `right: Expr` | Binary arithmetic. |
| `"log"` | `label: string`, `field: string` | **Only appears inside `progress`** (§7) — reads what was actually logged for a named set this session. Never appears in a `target`/`load`. |

```json
{ "type": "mul",
  "left": { "type": "number", "value": 0.85 },
  "right": { "type": "catalogField", "catalog": "BarbellBackSquat", "field": "e1rm" } }
```

**The client never evaluates an `Expr`.** In a `Program`, it's replaced
by a `value` once `resolve()` runs server-side; the only place a client
ever *sees* an `Expr` in practice is inside a `Session`'s
`exerciseDecl.progress` (§7), which it should treat as opaque.

### 6.6 `Conditional<Member>` — deferred branch (**Program only**)

```
Conditional {
  type: "conditional"
  cond: BoolExpr    // §6.8
  then: Member
  else: Member
}
```

Appears wherever an `if cond then: A else B` was written at the
day/block/top level. `cond` reads athlete `state`, so it can't be
decided until resolve time — both branches are compiled and kept.
**Never appears in a `Session`** — `resolve()` evaluates `cond` and
replaces this node with whichever branch it picked; the other branch
is dropped entirely and never reaches the client.

> **`Ref`/`Conditional` inlining rule** (matters if you ever walk a
> `Program` directly, e.g. for a pre-resolve preview): when the server
> inlines a `Ref` or a `Conditional`'s chosen branch, it substitutes
> the **whole nested `Group`/`ExerciseDecl.body` value**, not just its
> `members` spliced flat into the parent. So `Ref(bench)` inside a
> `superset`'s `members` becomes a full nested `{"type":"group",
> "kind":"straight", ...}` node (bench's own body, with its own
> `termination.n` = however many sets bench has), not bench's sets
> spliced directly into the superset's member list. Don't assume a
> `Session`'s `body` is exactly as "flat" as the source read — walk it
> recursively.

### 6.7 `Fallback` (**Program only**)

Appears on `stateBinding`, `target`, and `load` (`Program` shapes only
— never in a resolved `Session`, since by definition a fallback has
already been applied or wasn't needed by the time resolution finishes).

Tagged by `kind`:

| `kind` | Fields | Meaning |
|---|---|---|
| `"numeric"` | `value: number`, `plus?: bool` | A plain default used when the athlete has no recorded history yet. |
| `"sentinel"` | `name: string` | An onboarding marker — the app must **ask the athlete directly** for this value; resolution cannot derive it. |

```json
{ "kind": "numeric", "value": 60 }
{ "kind": "sentinel", "name": "AMW" }
```

`"AMW"` ("Assumed Max Weight") is the only sentinel in use today. The
set of valid sentinel names isn't fixed by the schema (any all-caps
identifier is legal) — treat an unrecognized sentinel as "ask the
athlete for a starting number for this exercise" as a safe generic
fallback behavior, and add a proper onboarding flow for `AMW`
specifically. **A `Fallback` never appears in a `Session`** — you'll
only ever see this while walking a raw `Program` before it's resolved,
or conceptually understanding why resolve might need onboarding input
first.

### 6.8 `BoolExpr` — conditions

Shared verbatim by `Conditional.cond` (§6.6) and `progressionIf.cond`
(§7). Tagged by `type`, recursive:

```
BoolExpr =
  | Comparison { type: "comparison", op: "<"|">"|"<="|">="|"=="|"!=", left: Expr, right: Expr }
  | And        { type: "and", left: BoolExpr, right: BoolExpr }
  | Or         { type: "or",  left: BoolExpr, right: BoolExpr }
```
`and` binds tighter than `or` (already resolved by the compiler into
this tree shape — you never need to parse precedence yourself, just
walk the tree as given).

```json
{ "type": "or",
  "left": { "type": "and",
    "left":  { "type": "comparison", "op": ">=", "left": { "type": "path", "path": "tm.squat.weight" }, "right": { "type": "number", "value": 90 } },
    "right": { "type": "comparison", "op": ">=", "left": { "type": "path", "path": "readiness" }, "right": { "type": "number", "value": 7 } } },
  "right": { "type": "comparison", "op": ">", "left": { "type": "path", "path": "fatigue" }, "right": { "type": "number", "value": 8 } } }
```

## 7. `progress` — carried, never evaluated client-side

`ExerciseDecl.progress`, if present, describes how the athlete's
training-max `state` should update after this session — but **this
logic always runs server-side**, after the client submits a
`SessionLog` (§8.3). The client receives it (unchanged, whether from a
`Program` or a `Session` — resolve never touches it) purely as
carried-along data; there's no product requirement to do anything with
it beyond, optionally, displaying human-readable progression rules
("weight goes up 5kg if you hit 12 reps") if you want to reverse-engineer
UI copy from it. **Never evaluate it, never resolve its `path`/`log`
expressions client-side.**

```
ProgressionBody { stmts: ProgressionStmt[] }
ProgressionStmt = Assign | ProgressionIf

Assign { type: "assign", path: string, expr: Expr }

ProgressionIf {
  type: "if"
  cond: BoolExpr           // §6.8
  then: ProgressionStmt[]
  else?: ProgressionStmt[] // optional — the only place `else` is optional
}
```

Inside a `progress` block specifically, an `Expr`'s `"type":"log"`
variant (§6.5) reads what was *actually logged* for a named set this
session (`label`+`field`, matching a `SessionLog` entry, §8.3) — this
is the one place `"log"` nodes appear.

```json
{
  "stmts": [
    {
      "type": "if",
      "cond": { "type": "comparison", "op": ">=", "left": { "type": "log", "label": "top_set", "field": "reps" }, "right": { "type": "number", "value": 8 } },
      "then": [
        { "type": "assign", "path": "tm.squat.reps", "expr": { "type": "log", "label": "top_set", "field": "reps" } },
        { "type": "assign", "path": "tm.squat.weight", "expr": { "type": "log", "label": "top_set", "field": "weight" } },
        {
          "type": "if",
          "cond": { "type": "comparison", "op": ">=", "left": { "type": "path", "path": "tm.squat.reps" }, "right": { "type": "number", "value": 12 } },
          "then": [
            { "type": "assign", "path": "tm.squat.weight", "expr": { "type": "add", "left": { "type": "path", "path": "tm.squat.weight" }, "right": { "type": "number", "value": 5 } } },
            { "type": "assign", "path": "tm.squat.reps", "expr": { "type": "number", "value": 8 } }
          ],
          "else": [
            { "type": "assign", "path": "tm.squat.reps", "expr": { "type": "add", "left": { "type": "path", "path": "tm.squat.reps" }, "right": { "type": "number", "value": 1 } } }
          ]
        }
      ]
    }
  ]
}
```

## 8. Supporting types — state, cursor, log

These aren't part of the `Program`/`Session` documents themselves, but
are the surrounding request/response shapes a client needs to know
about to request a `Session` and submit results. (Exact request/
response API shape is a backend concern outside this doc's scope —
these are the JSON *payload* shapes involved.)

### 8.1 `State` — an athlete's current values

What the server holds per athlete and consults when resolving. A
client only needs this shape if it ever needs to *display* an
athlete's current training maxes, or (in an onboarding flow) submit an
initial value for a sentinel fallback (§6.7).

```
State {
  bindings:      { path: string, value: number }[]
  catalogFields: { catalog: string, field: string, value: number }[]
}
```

```json
{
  "bindings": [ { "path": "tm.squat.weight", "value": 102.5 } ],
  "catalogFields": [ { "catalog": "BarbellBackSquat", "field": "e1rm", "value": 140 } ]
}
```

### 8.2 Cursor — which day to resolve

What a client sends (or a server derives from an athlete's own
progression cursor) to say "resolve *this* occurrence of the program":

```
Cursor {
  block:      string
  day:        string
  iteration?: integer  // 1-based; which pass through a repeated block (e.g. `leader*5`'s 3rd time)
}
```

### 8.3 `SessionLog` — what the client submits after a workout

One entry per `SetRef` the athlete actually performed and logged:

```
SessionLog = LogEntry[]

LogEntry {
  exercise:        string   // catalog id, matching the SetRef's `exercise`
  label:           string   // the set's label — matches `progress`'s `<label>.<field>` reads
  performedTarget: number   // reps completed / distance covered / duration held
  performedLoad?:  number   // weight actually used, if the set had a `load`
}
```

```json
[
  { "exercise": "BarbellBackSquat", "label": "top_set", "performedTarget": 12, "performedLoad": 100 }
]
```

Only sets with a `label` can be logged this way (anonymous/inline
sets, i.e. `SetRef`s with no `label`, aren't individually addressable
for progression purposes — they can still be displayed/checked-off in
the UI, just not referenced by a `progress` block).

## 9. Rendering/execution guide per `kind`

Practical mapping from the wire shapes above to what the workout
screen should actually do:

| `kind` | UI behavior |
|---|---|
| `straight` | Show sets one at a time in order; standard rest timer (`rest.single.duration`, if present) between them. |
| `superset` / `circuit` | Round-robin: show member 0's next set, then member 1's next set, etc., using `rest.two_level.intra` between them; after one full lap, use `rest.two_level.inter`; repeat for `termination.n` laps. |
| `dropset` / `restpause` | One continuous atomic sequence — typically no "rest screen" between members beyond `rest.single.duration` (0 for dropset, 15s+ for restpause); never interrupted by an outer group's own interleaving. |
| `emom` | Clock-driven: start an interval timer of `termination.interval`; athlete performs the member(s) as fast as possible, `rest.remainder` = whatever's left of the interval is automatic rest; repeat for `termination.n` intervals. |
| `amrap` | Countdown timer for `termination.cap`; athlete cycles through `members` back-to-back (`rest.ad_lib` — no prescribed rest) until time expires; record completed rounds + partial progress into the last round. |
| `for_time` (bare) | Stopwatch counts up; athlete performs `members` back-to-back, `rest.ad_lib`; stop the clock on completion — elapsed time is the result. |
| `for_time` (wrapping `rounds`) | Same stopwatch behavior, but sequence through each `rounds`-kind member (each is one full lap of the couplet at that rep count) strictly in order. |
| `rounds` | Not rendered as its own concept — it's always one lap inside a `for_time` wrapper; render its `members` as the couplet for that rep-scheme value. |

## 10. Common mistakes to avoid

- **Don't try to evaluate any `Expr`/`BoolExpr` client-side.** If
  you're looking at a `Program` (not a `Session`) and need concrete
  numbers, that means you need to call `resolve()` server-side — there
  is no client-side resolution logic to write, ever.
- **Don't assume `Session.body`'s nesting depth matches the source
  code's visual nesting.** `Ref` inlining nests the *whole* referenced
  `Group`, so a bare `squat;` statement at the day level resolves to a
  `Group` wrapping a `Group` wrapping the actual `SetRef`s — walk
  recursively, don't assume a fixed depth.
- **`superset` and `circuit` are the same shape** — don't build
  separate rendering logic for them; the `kind` string is purely a UI
  label.
- **A missing `load` is meaningful, not an error** — it means the set
  has no resistance to display (cardio, bodyweight, a duration hold).
  Don't treat it as malformed data.
- **`exercises`/`groups` on a `Session` can list more than `body`
  actually uses** (§3) — when a `Conditional` picked one branch, the
  untaken exercise/group is still listed in the flat metadata arrays.
  Always drive the actual workout from `body`, never from iterating
  `exercises` directly.
- **`"m"` means different things in different fields** — minutes in
  `Duration.unit`, meters in a distance `Target.unit`. They never
  collide within one field (each field's enum is fixed), so this only
  matters if you're writing a shared "format unit" helper — check
  which field you're formatting.
- **A fallback `sentinel` (e.g. `"AMW"`) means "go ask the athlete,"
  not "treat as zero/null."** A `Session` will never contain an
  unresolved fallback — if you see one, you're looking at a `Program`
  that hasn't been resolved yet (or your resolve request is missing
  required onboarding data the server needs first).

## 11. Versioning

Every `Program`/`Session` document is (or, for a `Session`, was
resolved from a `Program`) tagged `owlVersion` (SemVer, e.g. `"0.1.0"`)
on the `Program`. **Check it before trusting the document shape.**
While pre-`1.0`, any shape change may be breaking between versions —
pin your client's parser to a known-good `owlVersion` range and fail
loudly (not silently) on a mismatch, rather than attempting best-effort
parsing of an unknown version.

## Appendix A — compact type index

```
Program        { owlVersion, units, plates?, state, blocks, sequence? }
Block          { name, days, body }
Day            { name, exercises?, groups?, body }
ExerciseDecl   { name, catalog, sets, groups?, body, progress? }
NamedGroup     { name, group }

Session        { name, exercises?, groups?, body }   // resolved Day

Group          { type:"group", kind, members, interleave, rest, termination, atomic }
Member         = SetRef | Ref | Group | Conditional   // Ref/Conditional: Program only
SetRef         { type:"setRef", label?, exercise, target, load? }
Ref            { type:"ref", name }                   // Program only
Conditional    { type:"conditional", cond, then, else }  // Program only

Target         = { kind:"reps", expr|value, plus?, fallback? }
               | { kind:"distance", expr|value, unit:"m"|"km", fallback? }
               | { kind:"duration", expr|value, unit:"s"|"m"|"h", fallback? }
Load           = { kind:"weight", expr|value, unit:"kg"|"lb", fallback? }

RestPolicy     = { mode:"single", duration? }
               | { mode:"two_level", intra, inter }
               | { mode:"ad_lib" }
               | { mode:"remainder" }
Termination    = { mode:"count", n }
               | { mode:"emom", interval, n }
               | { mode:"time_cap", cap }
               | { mode:"for_time" }
Duration       { value, unit:"s"|"m"|"h"|"d" }

Expr           = { type:"number", value }
               | { type:"path", path }
               | { type:"catalogField", catalog, field }
               | { type:"mul"|"add"|"sub"|"div", left, right }
               | { type:"log", label, field }          // progress only
Fallback       = { kind:"numeric", value, plus? } | { kind:"sentinel", name }
BoolExpr       = { type:"comparison", op, left, right }
               | { type:"and"|"or", left, right }

ProgressionBody { stmts }
Assign          { type:"assign", path, expr }
ProgressionIf   { type:"if", cond, then, else? }

StateBinding    { path, expr, fallback? }              // Program.state entries

State           { bindings: [{path,value}], catalogFields: [{catalog,field,value}] }
Cursor          { block, day, iteration? }
SessionLog      = [{ exercise, label, performedTarget, performedLoad? }]
```

`kind` enum (Group): `straight | superset | circuit | dropset | restpause | emom | amrap | for_time | rounds`
`interleave` enum: `sequential | round_robin`
`units` enum (weight): `kg | lb`
`op` enum (comparison): `< | > | <= | >= | == | !=`
