// Package owl is the Go reference implementation of the Open Workout
// Language. This file defines the wire types mirroring
// ../spec/canonical-form.schema.json — see ../spec/canonical-form.md for
// the prose version of this same shape.
//
// Nothing parses yet: Compile, Resolve, and Progress (see owl.go) are
// stubs. This file exists so the canonical-form JSON shape has a single,
// checked Go representation to build the lexer/parser/compiler against.
package owl

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
// implementations: NumberExpr, PathExpr, CatalogFieldExpr, MulExpr.
type Expr interface{ isExpr() }

type NumberExpr struct{ Value float64 }

type PathExpr struct{ Path string }

// CatalogFieldExpr is `$Exercise.field`, e.g. `$BarbellBackSquat.e1rm`.
type CatalogFieldExpr struct{ Catalog, Field string }

type MulExpr struct{ Left, Right Expr }

func (NumberExpr) isExpr()       {}
func (PathExpr) isExpr()         {}
func (CatalogFieldExpr) isExpr() {}
func (MulExpr) isExpr()          {}

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

// ProgressionRule is a `progress = scheme(target, ...args)` line
// (spec/semantics/progression.md, spec/stdlib/schemes.md).
type ProgressionRule struct {
	Scheme string    `json:"scheme"`
	Target string    `json:"target"` // label of the SetRef this rule progresses
	Args   []float64 `json:"args,omitempty"`
}

// SetRef is one `set` line, or an anonymous inline `$Catalog(...)` call.
type SetRef struct {
	Label       string // absent for anonymous/inline sets (surface `_`)
	Exercise    string // catalog name, e.g. "BarbellBackSquat"
	Target      Target
	Load        Load // nil for a target-only set (e.g. a cardio distance)
	Progression *ProgressionRule
}

// Ref is a bare-name invocation (grammar `ref ::= IDENT`) — resolved
// against the enclosing scope's declarations (an exercise, a named
// group, a day, or a block) rather than duplicating that declaration's
// body inline. See spec/canonical-form.md.
type Ref struct{ Name string }

// Member is one entry in a Group's Members list.
// Concrete implementations: SetRef, Ref, Group.
type Member interface{ isMember() }

func (SetRef) isMember() {}
func (Ref) isMember()    {}
func (Group) isMember()  {}

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
	Sets    []SetRef `json:"sets"`
	Body    Group    `json:"body"` // this declaration's own straight-set Group
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
