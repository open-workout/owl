// Package ir defines the canonical-form wire types mirroring
// ../../../spec/canonical-form.schema.json — see
// ../../../spec/canonical-form.md for the prose version of this same
// shape. It's a leaf package (imports nothing else in this module) so
// that both internal/compiler (which builds these values) and the
// public owl package (which re-exports them as type aliases — see
// ../../types.go) can import it without an import cycle.
package ir

// Program is the top-level canonical-form document.
type Program struct {
	OWLVersion string         `json:"owlVersion"`
	Units      string         `json:"units"`
	Plates     []float64      `json:"plates,omitempty"`
	State      []StateBinding `json:"state"`
	Blocks     []Block        `json:"blocks"`
	// Sequence is the top-level ordering/repetition of blocks, e.g.
	// `leader*5;`. See spec/semantics/groups.md §3.7.
	Sequence *Group `json:"sequence,omitempty"`
}

// StateBinding is one `state`/`stats` entry, e.g.
// `tm.squat.weight = 0.85 * $BarbellBackSquat.e1rm | AMW`.
type StateBinding struct {
	Path     string    `json:"path"`
	Expr     Expr      `json:"expr"`
	Fallback *Fallback `json:"fallback,omitempty"`
}

// Fallback models surface `| fallback` (spec/semantics/parameters.md §3):
// either a plain numeric default (Kind == "numeric") or an onboarding
// sentinel such as "AMW" (Kind == "sentinel").
type Fallback struct {
	Kind  string   `json:"kind"`            // "numeric" | "sentinel"
	Value *float64 `json:"value,omitempty"` // set iff Kind == "numeric"
	Plus  bool     `json:"plus,omitempty"`  // set iff Kind == "numeric"
	Name  string   `json:"name,omitempty"`  // set iff Kind == "sentinel", e.g. "AMW"
}

// Expr is the recursive arithmetic expression node. Concrete
// implementations: NumberExpr, PathExpr, CatalogFieldExpr, MulExpr,
// AddExpr, SubExpr, DivExpr, LogExpr.
type Expr interface{ isExpr() }

type NumberExpr struct{ Value float64 }

type PathExpr struct{ Path string }

// CatalogFieldExpr is `$Exercise.field`, e.g. `$BarbellBackSquat.e1rm`.
type CatalogFieldExpr struct{ Catalog, Field string }

type MulExpr struct{ Left, Right Expr }
type AddExpr struct{ Left, Right Expr }
type SubExpr struct{ Left, Right Expr }
type DivExpr struct{ Left, Right Expr }

// LogExpr reads what was actually logged for a set this session — e.g.
// surface `top_set.reps` inside a `progress` block. Label is the set's
// own label (already resolved away from any dropset-qualified source
// spelling — see spec/semantics/progression.md §2). Never appears
// outside a ProgressionBody.
type LogExpr struct{ Label, Field string }

func (NumberExpr) isExpr()       {}
func (PathExpr) isExpr()         {}
func (CatalogFieldExpr) isExpr() {}
func (MulExpr) isExpr()          {}
func (AddExpr) isExpr()          {}
func (SubExpr) isExpr()          {}
func (DivExpr) isExpr()          {}
func (LogExpr) isExpr()          {}

// Target is what the athlete is asked to do (spec/semantics/targets-loads.md
// §1). Concrete implementations: RepsTarget, DistanceTarget, DurationTarget.
type Target interface{ isTarget() }

type RepsTarget struct {
	Expr     Expr
	Plus     bool
	Fallback *Fallback
}

type DistanceTarget struct {
	Expr     Expr
	Unit     string // "m" | "km"
	Fallback *Fallback
}

type DurationTarget struct {
	Expr     Expr
	Unit     string // "s" | "m" | "h"
	Fallback *Fallback
}

func (RepsTarget) isTarget()     {}
func (DistanceTarget) isTarget() {}
func (DurationTarget) isTarget() {}

// Load is what a Target is performed against (spec/semantics/targets-loads.md
// §2). Only WeightLoad is attested today.
type Load interface{ isLoad() }

type WeightLoad struct {
	Expr     Expr
	Unit     string // "kg" | "lb"
	Fallback *Fallback
}

func (WeightLoad) isLoad() {}

// BoolExpr is the condition of an `if`/`then`/`else` (grammar's
// `condExpr`) — a relational comparison, or two conditions combined with
// `and`/`or`. Shared verbatim by Conditional.Cond and ProgressionIf.Cond
// (spec/semantics/conditionals.md §2, §2a). Concrete implementations:
// Comparison, AndExpr, OrExpr.
type BoolExpr interface{ isBoolExpr() }

type Comparison struct {
	Op    string // "<" | ">" | "<=" | ">=" | "==" | "!="
	Left  Expr
	Right Expr
}

type AndExpr struct{ Left, Right BoolExpr }
type OrExpr struct{ Left, Right BoolExpr }

func (Comparison) isBoolExpr() {}
func (AndExpr) isBoolExpr()    {}
func (OrExpr) isBoolExpr()     {}

// ProgressionBody is an exercise's `progress = { ... }` block — at most
// one per ExerciseDecl. See spec/semantics/progression.md.
type ProgressionBody struct {
	Stmts []ProgressionStmt
}

// ProgressionStmt is one progressBody statement. Concrete
// implementations: Assign, ProgressionIf.
type ProgressionStmt interface{ isProgressionStmt() }

// Assign is `<state path> = <expr>` inside a progress block.
type Assign struct {
	Path string
	Expr Expr
}

// ProgressionIf is progress's own if/then/else — evaluated directly by
// progress() against the session log and current state, never deferred
// the way Conditional is. Else may be nil: the only place in the
// language `else` is optional (spec/semantics/progression.md §2).
type ProgressionIf struct {
	Cond BoolExpr
	Then []ProgressionStmt
	Else []ProgressionStmt
}

func (Assign) isProgressionStmt()        {}
func (ProgressionIf) isProgressionStmt() {}

// SetRef is one `set` line, or an anonymous inline `$Catalog(...)` call.
type SetRef struct {
	Label    string // absent for anonymous/inline sets (surface `_`)
	Exercise string // catalog name, e.g. "BarbellBackSquat"
	Target   Target
	Load     Load // nil for a target-only set (e.g. a cardio distance)
}

// Ref is a bare-name invocation (grammar `ref ::= IDENT`) — resolved
// against the enclosing scope's declarations (an exercise, a named
// group, a day, or a block) rather than duplicating that declaration's
// body inline. See spec/canonical-form.md.
type Ref struct{ Name string }

// Member is one entry in a Group's Members list.
// Concrete implementations: SetRef, Ref, Group, Conditional.
type Member interface{ isMember() }

func (SetRef) isMember() {}
func (Ref) isMember()    {}
func (Group) isMember()  {}

// Conditional is the compiled shape of `if cond then: A else B` when A/B
// are Members — e.g. a dayItem/blockItem/topLevelItem-level choice
// between two exercises, groups, or refs. Both branches are compiled and
// kept; resolve picks one. See spec/semantics/conditionals.md §2.
type Conditional struct {
	Cond BoolExpr
	Then Member
	Else Member
}

func (Conditional) isMember() {}

// RestPolicy is the rest at each transition within a Group
// (spec/semantics/groups.md §1). Concrete implementations: SingleRest,
// TwoLevelRest, AdLibRest, RemainderRest.
type RestPolicy interface{ isRestPolicy() }

type SingleRest struct{ Duration *Duration }

type TwoLevelRest struct{ Intra, Inter Duration }

type AdLibRest struct{}

type RemainderRest struct{}

func (SingleRest) isRestPolicy()    {}
func (TwoLevelRest) isRestPolicy()  {}
func (AdLibRest) isRestPolicy()     {}
func (RemainderRest) isRestPolicy() {}

type Duration struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"` // "s" | "m" | "h" | "d"
}

// Termination is when a Group ends (spec/semantics/groups.md §1).
// Concrete implementations: CountTermination, EmomTermination,
// TimeCapTermination, ForTimeTermination.
type Termination interface{ isTermination() }

type CountTermination struct{ N int }

type EmomTermination struct {
	Interval Duration
	N        int
}

type TimeCapTermination struct{ Cap Duration }

type ForTimeTermination struct{}

func (CountTermination) isTermination()   {}
func (EmomTermination) isTermination()    {}
func (TimeCapTermination) isTermination() {}
func (ForTimeTermination) isTermination() {}

// Group is the single IR node every superset/circuit/EMOM/AMRAP/for-time
// construct compiles to. See spec/semantics/groups.md.
type Group struct {
	Kind        string // "straight" | "superset" | "circuit" | "dropset" | "restpause" | "emom" | "amrap" | "for_time" | "rounds"
	Members     []Member
	Interleave  string // "sequential" | "round_robin"
	Rest        RestPolicy
	Termination Termination
	Atomic      bool
}

// ExerciseDecl is an `exercise <name> = $Catalog { set ... }` declaration.
type ExerciseDecl struct {
	Name    string   `json:"name"`    // local alias, e.g. "back_squat" — distinct from Catalog
	Catalog string   `json:"catalog"` // e.g. "BarbellBackSquat"
	Sets    []SetRef `json:"sets"`    // flat, top-level sets only — not ones nested in Groups
	// Groups holds named `dropset` declarations (spec/semantics/groups.md
	// §3.8), mirroring Day.Groups for `partA`-style assignments.
	Groups []NamedGroup `json:"groups,omitempty"`
	Body   Group        `json:"body"` // this declaration's own straight-set Group
	// Progress is this exercise's progression code, at most one per
	// exercise. See spec/semantics/progression.md.
	Progress *ProgressionBody `json:"progress,omitempty"`
}

// NamedGroup is a `partA = <groupExpr>` assignment.
type NamedGroup struct {
	Name  string `json:"name"`
	Group Group  `json:"group"`
}

type Day struct {
	Name      string         `json:"name"`
	Exercises []ExerciseDecl `json:"exercises,omitempty"`
	Groups    []NamedGroup   `json:"groups,omitempty"`
	Body      Group          `json:"body"`
}

type Block struct {
	Name string `json:"name"`
	Days []Day  `json:"days"`
	Body Group  `json:"body"`
}
