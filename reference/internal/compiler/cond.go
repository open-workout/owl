package compiler

import (
	"github.com/open-workout/owl/reference/internal/ir"
	"github.com/open-workout/owl/reference/internal/parser"
)

// compileCond lowers a condExpr into a BoolExpr. Structural compilation
// never evaluates it — `cond` reads per-athlete state, so both branches
// of whatever it guards are compiled and kept; resolve picks one. See
// spec/semantics/conditionals.md §2.
func (c *compiler) compileCond(cond parser.Cond, sc *scope) (ir.BoolExpr, error) {
	left, err := c.compileExpr(cond.Left, sc)
	if err != nil {
		return ir.BoolExpr{}, err
	}
	right, err := c.compileExpr(cond.Right, sc)
	if err != nil {
		return ir.BoolExpr{}, err
	}
	return ir.BoolExpr{Op: cond.Op, Left: left, Right: right}, nil
}
