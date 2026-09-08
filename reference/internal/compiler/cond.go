package compiler

import (
	"fmt"

	"github.com/open-workout/owl/reference/internal/ir"
	"github.com/open-workout/owl/reference/internal/parser"
)

// exprCompiler is either c.compileExpr (structural contexts — a
// Conditional<Member>'s cond) or c.compileProgressExpr (a progressIf's
// cond) — the two differ only in how they resolve a dottedPath (see
// compileProgressExpr's doc comment), so compileCond takes which one to
// use rather than duplicating itself.
type exprCompiler func(parser.Expr, *scope) (ir.Expr, error)

// compileCond lowers a condExpr into a BoolExpr — a comparison, or two
// conditions combined with and/or. Structural compilation never
// evaluates it — `cond` reads per-athlete state (for a
// Conditional<Member>) or the session log (for a progressIf), so
// whichever consumer runs later picks a branch; see
// spec/semantics/conditionals.md §2, §2a.
func (c *compiler) compileCond(cond parser.Cond, sc *scope, compileE exprCompiler) (ir.BoolExpr, error) {
	switch v := cond.(type) {
	case *parser.Comparison:
		left, err := compileE(v.Left, sc)
		if err != nil {
			return nil, err
		}
		right, err := compileE(v.Right, sc)
		if err != nil {
			return nil, err
		}
		return ir.Comparison{Op: v.Op, Left: left, Right: right}, nil
	case *parser.AndCond:
		left, err := c.compileCond(v.Left, sc, compileE)
		if err != nil {
			return nil, err
		}
		right, err := c.compileCond(v.Right, sc, compileE)
		if err != nil {
			return nil, err
		}
		return ir.AndExpr{Left: left, Right: right}, nil
	case *parser.OrCond:
		left, err := c.compileCond(v.Left, sc, compileE)
		if err != nil {
			return nil, err
		}
		right, err := c.compileCond(v.Right, sc, compileE)
		if err != nil {
			return nil, err
		}
		return ir.OrExpr{Left: left, Right: right}, nil
	default:
		return nil, fmt.Errorf("compiler: unhandled cond %T", cond)
	}
}
