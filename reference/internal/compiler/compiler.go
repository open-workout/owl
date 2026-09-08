// Package compiler lowers a parser.Program (source-shaped AST) into the
// canonical-form ir.Program, implementing spec/semantics/groups.md's
// desugaring rules, targets-loads.md's local-reference substitution,
// and the single-owner-rule validation from parameters.md.
package compiler

import (
	"strings"

	"github.com/open-workout/owl/reference/internal/ir"
	"github.com/open-workout/owl/reference/internal/parser"
)

const owlVersion = "0.1.0"

type compiler struct {
	programUnits string
	stateRoots   map[string]bool
}

// Compile lowers a parsed Program into canonical form.
func Compile(prog *parser.Program) (*ir.Program, error) {
	c := &compiler{programUnits: prog.Units, stateRoots: map[string]bool{}}
	out := &ir.Program{OWLVersion: owlVersion, Units: prog.Units, Plates: prog.Plates, State: []ir.StateBinding{}}

	// State binding roots must all be known before compiling any
	// binding's expr (or anything else), since a dottedPath's
	// classification depends on the full set of declared roots, not
	// just ones declared earlier in source order.
	for _, b := range prog.State {
		if len(b.Path.Segments) > 0 {
			c.stateRoots[b.Path.Segments[0]] = true
		}
	}
	for _, b := range prog.State {
		expr, err := c.compileExpr(b.Expr, nil)
		if err != nil {
			return nil, err
		}
		out.State = append(out.State, ir.StateBinding{
			Path: strings.Join(b.Path.Segments, "."), Expr: expr, Fallback: compileFallback(b.Fallback),
		})
	}

	sc := newScope(nil)
	var explicitItems []any
	for _, raw := range prog.Items {
		switch item := raw.(type) {
		case *parser.BlockDecl:
			blk, err := c.compileBlockDecl(item, sc)
			if err != nil {
				return nil, err
			}
			out.Blocks = append(out.Blocks, *blk)
			sc.declareName(blk.Name)
			bodyCopy := blk.Body
			sc.declareBody(blk.Name, &bodyCopy)
		default:
			explicitItems = append(explicitItems, raw)
		}
	}

	// Unlike Block.Body/Day.Body (always synthesized — see
	// compileBlockDecl/compileDayDecl), Program.Sequence stays absent
	// when there's no explicit top-level statement: none of the
	// single-block fixtures have a "sequence" key at all.
	if len(explicitItems) > 0 {
		ci, err := c.compileItemList(explicitItems, sc, nil)
		if err != nil {
			return nil, err
		}
		out.Sequence = &ir.Group{
			Kind: "straight", Interleave: "sequential", Rest: restFromDuration(ci.restDuration),
			Termination: ir.CountTermination{N: len(ci.members)}, Atomic: false, Members: ci.members,
		}
	}

	if err := validateSingleOwner(out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *compiler) compileBlockDecl(bd *parser.BlockDecl, parent *scope) (*ir.Block, error) {
	sc := newScope(parent)
	var days []ir.Day
	var groups []ir.NamedGroup
	var explicitItems []any
	var declOrder []string

	for _, raw := range bd.Items {
		switch item := raw.(type) {
		case *parser.DayDecl:
			day, err := c.compileDayDecl(item, sc)
			if err != nil {
				return nil, err
			}
			days = append(days, *day)
			sc.declareName(day.Name)
			bodyCopy := day.Body
			sc.declareBody(day.Name, &bodyCopy)
			declOrder = append(declOrder, day.Name)
		default:
			explicitItems = append(explicitItems, raw)
		}
	}

	members, restDur, err := c.finalizeBody(explicitItems, declOrder, sc, &groups)
	if err != nil {
		return nil, err
	}
	body := ir.Group{
		Kind: "straight", Interleave: "sequential", Rest: restFromDuration(restDur),
		Termination: ir.CountTermination{N: len(members)}, Atomic: false, Members: members,
	}
	return &ir.Block{Name: bd.Name, Days: days, Body: body}, nil
}

func (c *compiler) compileDayDecl(dd *parser.DayDecl, parent *scope) (*ir.Day, error) {
	sc := newScope(parent)
	var exercises []ir.ExerciseDecl
	var groups []ir.NamedGroup
	var explicitItems []any
	var declOrder []string

	for _, raw := range dd.Items {
		switch item := raw.(type) {
		case *parser.ExerciseDecl:
			ex, err := c.compileExerciseDecl(item, sc)
			if err != nil {
				return nil, err
			}
			exercises = append(exercises, *ex)
			sc.declareName(ex.Name)
			bodyCopy := ex.Body
			sc.declareBody(ex.Name, &bodyCopy)
			declOrder = append(declOrder, ex.Name)
		default:
			explicitItems = append(explicitItems, raw)
		}
	}

	members, restDur, err := c.finalizeBody(explicitItems, declOrder, sc, &groups)
	if err != nil {
		return nil, err
	}
	body := ir.Group{
		Kind: "straight", Interleave: "sequential", Rest: restFromDuration(restDur),
		Termination: ir.CountTermination{N: len(members)}, Atomic: false, Members: members,
	}
	return &ir.Day{Name: dd.Name, Exercises: exercises, Groups: groups, Body: body}, nil
}

// finalizeBody implements canonical-form.md's synthesis rule shared by
// Block.Body and Day.Body: if the source has any explicit statement
// after its declarations, the body's members come from those; otherwise
// it's auto-synthesized as one Ref per declared sub-entity, in source
// order.
func (c *compiler) finalizeBody(explicitItems []any, declOrder []string, sc *scope, groupsSink *[]ir.NamedGroup) ([]ir.Member, *ir.Duration, error) {
	if len(explicitItems) > 0 {
		ci, err := c.compileItemList(explicitItems, sc, groupsSink)
		if err != nil {
			return nil, nil, err
		}
		return ci.members, ci.restDuration, nil
	}
	members := make([]ir.Member, len(declOrder))
	for i, name := range declOrder {
		members[i] = ir.Ref{Name: name}
	}
	return members, nil, nil
}
