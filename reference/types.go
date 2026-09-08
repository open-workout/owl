// Package owl is the Go reference implementation of the Open Workout
// Language. This file re-exports internal/ir's canonical-form types
// (mirroring ../spec/canonical-form.schema.json — see
// ../spec/canonical-form.md for the prose version) as type aliases, so
// the public API stays `owl.Program`, `owl.Group`, etc. exactly as
// before. The types themselves live in internal/ir, not here, because
// internal/compiler needs to construct them and internal/compiler is
// imported by this package (to wire Compile in owl.go) — putting the
// types in this package too would make that an import cycle. A type
// alias (`type X = ir.X`) is the identical underlying type, including
// its MarshalJSON/UnmarshalJSON methods, so this is a pure reorg with no
// change in behavior.
package owl

import "github.com/open-workout/owl/reference/internal/ir"

type (
	Program                  = ir.Program
	StateBinding             = ir.StateBinding
	Fallback                 = ir.Fallback
	Expr                     = ir.Expr
	NumberExpr               = ir.NumberExpr
	PathExpr                 = ir.PathExpr
	CatalogFieldExpr         = ir.CatalogFieldExpr
	MulExpr                  = ir.MulExpr
	Target                   = ir.Target
	RepsTarget               = ir.RepsTarget
	DistanceTarget           = ir.DistanceTarget
	DurationTarget           = ir.DurationTarget
	Load                     = ir.Load
	WeightLoad               = ir.WeightLoad
	ProgressionRule          = ir.ProgressionRule
	BoolExpr                 = ir.BoolExpr
	ProgressionOrConditional = ir.ProgressionOrConditional
	ConditionalProgression   = ir.ConditionalProgression
	SetRef                   = ir.SetRef
	Ref                      = ir.Ref
	Member                   = ir.Member
	Conditional              = ir.Conditional
	RestPolicy               = ir.RestPolicy
	SingleRest               = ir.SingleRest
	TwoLevelRest             = ir.TwoLevelRest
	AdLibRest                = ir.AdLibRest
	RemainderRest            = ir.RemainderRest
	Duration                 = ir.Duration
	Termination              = ir.Termination
	CountTermination         = ir.CountTermination
	EmomTermination          = ir.EmomTermination
	TimeCapTermination       = ir.TimeCapTermination
	ForTimeTermination       = ir.ForTimeTermination
	Group                    = ir.Group
	ExerciseDecl             = ir.ExerciseDecl
	NamedGroup               = ir.NamedGroup
	Day                      = ir.Day
	Block                    = ir.Block
)
