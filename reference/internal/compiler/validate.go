package compiler

import (
	"fmt"

	"github.com/open-workout/owl/reference/internal/ir"
)

// owner records who first claimed a state path — spec/semantics/
// parameters.md §2's single-owner rule: each state path may be written
// (i.e. targeted by a progression rule reading it) by at most one set,
// anywhere in the program.
type owner struct{ label, exercise string }

// validateSingleOwner walks every exercise's compiled body (never
// ExerciseDecl.Sets/Groups too — those are the same SetRefs embedded a
// second time, and walking both would double-count, not find new
// violations) and errors on the first state path claimed by two
// different sets.
func validateSingleOwner(prog *ir.Program) error {
	owners := map[string]owner{}
	for _, blk := range prog.Blocks {
		for _, day := range blk.Days {
			for _, ex := range day.Exercises {
				if err := walkGroupForValidation(ex.Body, ex.Name, owners); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func walkGroupForValidation(g ir.Group, exerciseName string, owners map[string]owner) error {
	for _, m := range g.Members {
		if err := walkMemberForValidation(m, exerciseName, owners); err != nil {
			return err
		}
	}
	return nil
}

func walkMemberForValidation(m ir.Member, exerciseName string, owners map[string]owner) error {
	switch v := m.(type) {
	case ir.SetRef:
		return checkSetRef(v, exerciseName, owners)
	case ir.Group:
		return walkGroupForValidation(v, exerciseName, owners)
	case ir.Conditional:
		// Both branches are checked now, not deferred to resolve — the
		// single-owner rule is a compile-time invariant that can't wait
		// on per-athlete state (spec/semantics/conditionals.md §3).
		if err := walkMemberForValidation(v.Then, exerciseName, owners); err != nil {
			return err
		}
		return walkMemberForValidation(v.Else, exerciseName, owners)
	default:
		return nil
	}
}

// checkSetRef registers the state path(s) a progression-bearing SetRef's
// own target/load expression reads — those are exactly the paths its
// scheme will overwrite for the next cycle (stdlib/schemes.md), which is
// what "owns" means here.
func checkSetRef(sr ir.SetRef, exerciseName string, owners map[string]owner) error {
	if sr.Progression == nil {
		return nil
	}
	paths := map[string]bool{}
	collectExprPaths(targetExpr(sr.Target), paths)
	collectExprPaths(exprOfLoad(sr.Load), paths)
	for path := range paths {
		if existing, ok := owners[path]; ok {
			return fmt.Errorf("state path '%s' has more than one progression owner: '%s' (in exercise '%s') and '%s' (in exercise '%s')",
				path, existing.label, existing.exercise, sr.Label, exerciseName)
		}
		owners[path] = owner{label: sr.Label, exercise: exerciseName}
	}
	return nil
}

func targetExpr(t ir.Target) ir.Expr {
	switch v := t.(type) {
	case ir.RepsTarget:
		return v.Expr
	case ir.DistanceTarget:
		return v.Expr
	case ir.DurationTarget:
		return v.Expr
	default:
		return nil
	}
}

func exprOfLoad(l ir.Load) ir.Expr {
	if l == nil {
		return nil
	}
	if wl, ok := l.(ir.WeightLoad); ok {
		return wl.Expr
	}
	return nil
}

func collectExprPaths(e ir.Expr, out map[string]bool) {
	switch v := e.(type) {
	case ir.PathExpr:
		out[v.Path] = true
	case ir.MulExpr:
		collectExprPaths(v.Left, out)
		collectExprPaths(v.Right, out)
	}
}
