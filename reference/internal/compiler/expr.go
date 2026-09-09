package compiler

import (
	"fmt"
	"strings"

	"github.com/open-workout/owl/reference/internal/ir"
	"github.com/open-workout/owl/reference/internal/parser"
)

// compileExpr lowers a parsed expression into the canonical Expr union.
// A bare NumberLit's Unit (if any) is intentionally dropped here — units
// live on the enclosing Target/Load, not on the number itself; see
// compileTarget/compileLoad, which peek the raw AST node before calling
// this.
func (c *compiler) compileExpr(e parser.Expr, sc *scope) (ir.Expr, error) {
	switch v := e.(type) {
	case *parser.NumberLit:
		return ir.NumberExpr{Value: v.Value}, nil
	case *parser.CatalogFieldRef:
		return ir.CatalogFieldExpr{Catalog: v.Catalog, Field: v.Field}, nil
	case *parser.MulExpr:
		left, right, err := c.compileBinOperands(v.Left, v.Right, sc, c.compileExpr)
		if err != nil {
			return nil, err
		}
		return ir.MulExpr{Left: left, Right: right}, nil
	case *parser.DivExpr:
		left, right, err := c.compileBinOperands(v.Left, v.Right, sc, c.compileExpr)
		if err != nil {
			return nil, err
		}
		return ir.DivExpr{Left: left, Right: right}, nil
	case *parser.AddExpr:
		left, right, err := c.compileBinOperands(v.Left, v.Right, sc, c.compileExpr)
		if err != nil {
			return nil, err
		}
		return ir.AddExpr{Left: left, Right: right}, nil
	case *parser.SubExpr:
		left, right, err := c.compileBinOperands(v.Left, v.Right, sc, c.compileExpr)
		if err != nil {
			return nil, err
		}
		return ir.SubExpr{Left: left, Right: right}, nil
	case *parser.DottedPath:
		return c.resolvePathExpr(v, sc)
	default:
		return nil, fmt.Errorf("compiler: unhandled expr type %T", e)
	}
}

// compileProgressExpr compiles an expr appearing inside a `progress`
// block. It differs from compileExpr only in how it resolves a
// dottedPath: `<label>.<field>` reads the *logged* value for that set
// (LogExpr), not the prescribed target/load substitution
// resolvePathExpr performs — see spec/semantics/progression.md §2 and
// targets-loads.md §4's note. Arithmetic and literals compile
// identically either way.
func (c *compiler) compileProgressExpr(e parser.Expr, sc *scope) (ir.Expr, error) {
	switch v := e.(type) {
	case *parser.NumberLit:
		return ir.NumberExpr{Value: v.Value}, nil
	case *parser.CatalogFieldRef:
		return ir.CatalogFieldExpr{Catalog: v.Catalog, Field: v.Field}, nil
	case *parser.MulExpr:
		left, right, err := c.compileBinOperands(v.Left, v.Right, sc, c.compileProgressExpr)
		if err != nil {
			return nil, err
		}
		return ir.MulExpr{Left: left, Right: right}, nil
	case *parser.DivExpr:
		left, right, err := c.compileBinOperands(v.Left, v.Right, sc, c.compileProgressExpr)
		if err != nil {
			return nil, err
		}
		return ir.DivExpr{Left: left, Right: right}, nil
	case *parser.AddExpr:
		left, right, err := c.compileBinOperands(v.Left, v.Right, sc, c.compileProgressExpr)
		if err != nil {
			return nil, err
		}
		return ir.AddExpr{Left: left, Right: right}, nil
	case *parser.SubExpr:
		left, right, err := c.compileBinOperands(v.Left, v.Right, sc, c.compileProgressExpr)
		if err != nil {
			return nil, err
		}
		return ir.SubExpr{Left: left, Right: right}, nil
	case *parser.DottedPath:
		return c.resolveProgressPathExpr(v, sc)
	default:
		return nil, fmt.Errorf("compiler: unhandled expr type %T", e)
	}
}

func (c *compiler) compileBinOperands(l, r parser.Expr, sc *scope, compileE exprCompiler) (ir.Expr, ir.Expr, error) {
	left, err := compileE(l, sc)
	if err != nil {
		return nil, nil, err
	}
	right, err := compileE(r, sc)
	if err != nil {
		return nil, nil, err
	}
	return left, right, nil
}

// resolveProgressPathExpr is resolvePathExpr's counterpart for a
// progress block: `<label>.<field>` (bare, or dropset-qualified
// `<dropset>.<label>.<field>`) reads what was logged for that set —
// LogExpr — rather than substituting its prescribed formula. A bare
// state path still means the current state value, unchanged from
// resolvePathExpr.
func (c *compiler) resolveProgressPathExpr(dp *parser.DottedPath, sc *scope) (ir.Expr, error) {
	segs := dp.Segments
	full := strings.Join(segs, ".")
	head := segs[0]

	if sc.lookupSet(head) != nil {
		if len(segs) != 2 {
			return nil, fmt.Errorf("%s: %q: expected '<label>.<field>' inside a progress block", dp.Pos, full)
		}
		return ir.LogExpr{Label: head, Field: segs[1]}, nil
	}
	if d := sc.lookupDropset(head); d != nil {
		if len(segs) != 3 {
			return nil, fmt.Errorf("%s: %q: expected '<dropset>.<label>.<field>' inside a progress block", dp.Pos, full)
		}
		if _, ok := d.byLabel[segs[1]]; !ok {
			return nil, fmt.Errorf("%s: dropset %q has no set labeled %q", dp.Pos, head, segs[1])
		}
		return ir.LogExpr{Label: segs[1], Field: segs[2]}, nil
	}
	if c.stateRoots[head] {
		return ir.PathExpr{Path: full}, nil
	}
	return nil, fmt.Errorf("%s: dotted path %q does not resolve to a set label, dropset, or state binding", dp.Pos, full)
}

// resolvePathExpr classifies a dottedPath by scope lookup, per
// grammar.ebnf's note under `dottedPath`: a leading segment naming an
// in-scope `set` label makes it a local set-field reference (only
// `.weight` is attested — spec/semantics/targets-loads.md §4); a
// leading segment naming an in-scope `dropset` qualifies a label inside
// it the same way; otherwise it must name a declared `state` binding
// root. A local reference compiles by substituting the referenced set's
// own Load.Expr — expr trees are immutable once built, so reusing the
// same value here (rather than deep-cloning it) is safe.
func (c *compiler) resolvePathExpr(dp *parser.DottedPath, sc *scope) (ir.Expr, error) {
	segs := dp.Segments
	full := strings.Join(segs, ".")
	head := segs[0]

	if ref := sc.lookupSet(head); ref != nil {
		if len(segs) != 2 || segs[1] != "weight" {
			return nil, fmt.Errorf("%s: local set reference %q: only '<label>.weight' is supported", dp.Pos, full)
		}
		return loadExprOf(ref, full)
	}
	if d := sc.lookupDropset(head); d != nil {
		if len(segs) != 3 || segs[2] != "weight" {
			return nil, fmt.Errorf("%s: dropset-qualified reference %q: expected '<dropset>.<label>.weight'", dp.Pos, full)
		}
		inner, ok := d.byLabel[segs[1]]
		if !ok {
			return nil, fmt.Errorf("%s: dropset %q has no set labeled %q", dp.Pos, head, segs[1])
		}
		return loadExprOf(inner, full)
	}
	if c.stateRoots[head] {
		return ir.PathExpr{Path: full}, nil
	}
	return nil, fmt.Errorf("%s: dotted path %q does not resolve to a set label, dropset, or state binding", dp.Pos, full)
}

func loadExprOf(ref *ir.SetRef, full string) (ir.Expr, error) {
	wl, ok := ref.Load.(ir.WeightLoad)
	if !ok {
		return nil, fmt.Errorf("local set reference %q: set has no load to reference", full)
	}
	return wl.Expr, nil
}

// quantityUnitKind classifies a literal unit suffix (grammar's `unit`
// production, only meaningful when it sits directly on a Quantity's
// top-level factor) into which Target/Load union member it selects.
// "" means no unit was present (or the value wasn't a bare NumberLit) —
// callers default to Reps in that case.
func quantityUnitKind(q parser.Quantity) (kind, unit string, ok bool) {
	lit, isLit := q.Value.(*parser.NumberLit)
	if !isLit || lit.Unit == "" {
		return "", "", false
	}
	switch lit.Unit {
	case "m", "km":
		return "distance", lit.Unit, true
	case "s", "h":
		return "duration", lit.Unit, true
	case "min":
		return "duration", "m", true // canonical DurationTarget unit is "m" for minutes
	case "kg", "lb":
		return "weight", lit.Unit, true
	default:
		return "", "", false
	}
}

func (c *compiler) compileTarget(q parser.Quantity, sc *scope) (ir.Target, error) {
	expr, err := c.compileExpr(q.Value, sc)
	if err != nil {
		return nil, err
	}
	kind, unit, ok := quantityUnitKind(q)
	if ok {
		switch kind {
		case "distance":
			return ir.DistanceTarget{Expr: expr, Unit: unit}, nil
		case "duration":
			return ir.DurationTarget{Expr: expr, Unit: unit}, nil
		case "weight":
			return nil, fmt.Errorf("a weight unit (%q) can't be used as a target — did you mean '@'?", unit)
		}
	}
	return ir.RepsTarget{Expr: expr, Plus: q.Plus}, nil
}

// compileLoad's unit is the literal suffix if the quantity carries a
// weight one, otherwise the program's global `units` declaration —
// every attested load (whether written as "100kg" or composed as
// "0.8 * top.weight" with no literal suffix at all) ends up "kg" this
// way, matching every fixture; a literal "lb" would still win if ever
// used against a "kg" program, since a literal always takes precedence.
func (c *compiler) compileLoad(q parser.Quantity, sc *scope) (ir.Load, error) {
	expr, err := c.compileExpr(q.Value, sc)
	if err != nil {
		return nil, err
	}
	unit := c.programUnits
	if kind, u, ok := quantityUnitKind(q); ok && kind == "weight" {
		unit = u
	}
	return ir.WeightLoad{Expr: expr, Unit: unit}, nil
}

func compileFallback(f *parser.Fallback) *ir.Fallback {
	if f == nil {
		return nil
	}
	if f.Numeric != nil {
		v := f.Numeric.Value
		return &ir.Fallback{Kind: "numeric", Value: &v, Plus: f.Numeric.Plus}
	}
	return &ir.Fallback{Kind: "sentinel", Name: f.Sentinel}
}
