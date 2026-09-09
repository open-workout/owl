package compiler

import (
	"fmt"

	"github.com/open-workout/owl/reference/internal/ir"
)

// validateSingleOwner implements spec/semantics/parameters.md §2: a
// state path may be assigned by at most one exercise's `progress`
// block. Ownership is now read directly off `assign` statements' `path`
// (spec/semantics/progression.md §4) rather than inferred from a
// SetRef's target/load expression — multiple assigns to the same path
// within *one* exercise's own progress block (e.g. across its own
// if/else branches) are fine, since at most one branch ever executes;
// only different exercises colliding is an error.
func validateSingleOwner(prog *ir.Program) error {
	owners := map[string]string{} // state path -> owning exercise name
	for _, blk := range prog.Blocks {
		for _, day := range blk.Days {
			for _, ex := range day.Exercises {
				if ex.Progress == nil {
					continue
				}
				paths := map[string]bool{}
				collectAssignPaths(ex.Progress.Stmts, paths)
				for path := range paths {
					if existing, ok := owners[path]; ok && existing != ex.Name {
						return fmt.Errorf("state path '%s' has more than one progression owner: exercise '%s' and exercise '%s'",
							path, existing, ex.Name)
					}
					owners[path] = ex.Name
				}
			}
		}
	}
	return nil
}

func collectAssignPaths(stmts []ir.ProgressionStmt, out map[string]bool) {
	for _, s := range stmts {
		switch v := s.(type) {
		case ir.Assign:
			out[v.Path] = true
		case ir.ProgressionIf:
			collectAssignPaths(v.Then, out)
			collectAssignPaths(v.Else, out)
		}
	}
}
