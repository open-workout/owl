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
		left, err := c.compileExpr(v.Left, sc)
		if err != nil {
			return nil, err
		}
		right, err := c.compileExpr(v.Right, sc)
		if err != nil {
			return nil, err
		}
		return ir.MulExpr{Left: left, Right: right}, nil
	case *parser.DottedPath:
		return c.resolvePathExpr(v, sc)
	default:
		return nil, fmt.Errorf("compiler: unhandled expr type %T", e)
	}
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

// resolveLabelRef resolves a progressDecl's first argument — a bare set
// label or a dropset-qualified `ds.label` — to both the *ir.SetRef it
// names (so a progression rule can attach to it) and the plain label
// string ProgressionRule.Target stores. This is a different resolution
// than resolvePathExpr: the result is a name/reference, not a value
// expression.
func resolveLabelRef(dp *parser.DottedPath, sc *scope) (*ir.SetRef, string, error) {
	segs := dp.Segments
	full := strings.Join(segs, ".")
	switch len(segs) {
	case 1:
		ref := sc.lookupSet(segs[0])
		if ref == nil {
			return nil, "", fmt.Errorf("%s: progress target %q: no such set in scope", dp.Pos, full)
		}
		return ref, segs[0], nil
	case 2:
		d := sc.lookupDropset(segs[0])
		if d == nil {
			return nil, "", fmt.Errorf("%s: progress target %q: %q is not an in-scope dropset", dp.Pos, full, segs[0])
		}
		ref, ok := d.byLabel[segs[1]]
		if !ok {
			return nil, "", fmt.Errorf("%s: progress target %q: dropset %q has no set labeled %q", dp.Pos, full, segs[0], segs[1])
		}
		return ref, segs[1], nil
	default:
		return nil, "", fmt.Errorf("%s: progress target %q: expected a set label or dropset-qualified label", dp.Pos, full)
	}
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
