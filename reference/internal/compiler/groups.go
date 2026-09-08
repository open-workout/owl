package compiler

import (
	"fmt"

	"github.com/open-workout/owl/reference/internal/ir"
	"github.com/open-workout/owl/reference/internal/parser"
)

// defaultSupersetInterRest is the constant `inter` (between-lap) rest
// superset/circuit use — spec/semantics/groups.md §3.2 notes there's no
// attested surface syntax to override it yet.
var defaultSupersetInterRest = ir.Duration{Value: 90, Unit: "s"}

func toIRDuration(d parser.Duration) ir.Duration { return ir.Duration{Value: d.Value, Unit: d.Unit} }

// restFromDuration builds the "plain rest-between-sets" RestPolicy —
// {single}, optionally carrying an explicit duration when a bare
// rest(d) statement set one (see compileItemList).
func restFromDuration(d *ir.Duration) ir.RestPolicy { return ir.SingleRest{Duration: d} }

// compiledItems is the result of compiling a statement list: the
// members it produces, plus any bare rest(d) duration seen along the
// way (not itself a member — see compileItem).
type compiledItems struct {
	members      []ir.Member
	restDuration *ir.Duration
}

// compileItemList compiles a stmt-list body (a groupBody, or the
// explicit non-declaration items of a day/block/program). groupsSink,
// when non-nil, is where an AssignStmt's NamedGroup gets recorded
// (Day.Groups / ExerciseDecl.Groups / Block.Groups have one; a bare
// group-expr body like an amrap's does not, so groupsSink is nil there
// and an AssignStmt inside one is rejected).
func (c *compiler) compileItemList(items []any, sc *scope, groupsSink *[]ir.NamedGroup) (compiledItems, error) {
	var out compiledItems
	for _, raw := range items {
		m, restDur, err := c.compileItem(raw, sc, groupsSink)
		if err != nil {
			return compiledItems{}, err
		}
		if restDur != nil {
			// A bare rest(d) sets the enclosing group's own rest
			// duration rather than becoming a member — see
			// spec/semantics/groups.md §3.7a. If more than one
			// distinct duration appears (not attested with a golden
			// fixture; RestPolicy has only one scalar per group), the
			// last one wins — a documented simplification, not a
			// silently-wrong default.
			out.restDuration = restDur
			continue
		}
		if m != nil {
			out.members = append(out.members, m)
		}
	}
	return out, nil
}

func (c *compiler) compileItem(raw any, sc *scope, groupsSink *[]ir.NamedGroup) (ir.Member, *ir.Duration, error) {
	switch item := raw.(type) {
	case *parser.Ref:
		if !sc.hasName(item.Name) {
			return nil, nil, fmt.Errorf("%s: undefined reference %q", item.Pos, item.Name)
		}
		return ir.Ref{Name: item.Name}, nil, nil

	case *parser.RestStmt:
		d := toIRDuration(item.Duration)
		return nil, &d, nil

	case *parser.InlineCall:
		sr, err := c.compileInlineCall(item, sc)
		if err != nil {
			return nil, nil, err
		}
		return sr, nil, nil

	case *parser.RepeatStmt:
		g, err := c.compileRepeatStmt(item, sc)
		if err != nil {
			return nil, nil, err
		}
		return g, nil, nil

	case *parser.AssignStmt:
		if groupsSink == nil {
			return nil, nil, fmt.Errorf("%s: a named group assignment isn't allowed here", item.Pos)
		}
		g, err := c.compileGroupExpr(item.Group, sc)
		if err != nil {
			return nil, nil, err
		}
		sc.declareName(item.Name)
		gCopy := g
		sc.declareBody(item.Name, &gCopy)
		*groupsSink = append(*groupsSink, ir.NamedGroup{Name: item.Name, Group: g})
		// The assignment itself only declares — it doesn't sequence.
		// Every attested example separately sequences the name with a
		// later bare ref (e.g. `partA = ...` then `partA;`).
		return nil, nil, nil

	case parser.GroupExpr:
		g, err := c.compileGroupExpr(item, sc)
		if err != nil {
			return nil, nil, err
		}
		return g, nil, nil

	case *parser.CondItem:
		thenMember, _, err := c.compileItem(item.Then, sc, groupsSink)
		if err != nil {
			return nil, nil, err
		}
		elseMember, _, err := c.compileItem(item.Else, sc, groupsSink)
		if err != nil {
			return nil, nil, err
		}
		if thenMember == nil || elseMember == nil {
			return nil, nil, fmt.Errorf("%s: conditional branches must both produce a schedulable item", item.Pos)
		}
		cond, err := c.compileCond(item.Cond, sc, c.compileExpr)
		if err != nil {
			return nil, nil, err
		}
		return ir.Conditional{Cond: cond, Then: thenMember, Else: elseMember}, nil, nil

	default:
		return nil, nil, fmt.Errorf("compiler: unexpected statement %T", raw)
	}
}

func (c *compiler) compileInlineCall(ic *parser.InlineCall, sc *scope) (ir.SetRef, error) {
	target, err := c.compileTarget(ic.Target, sc)
	if err != nil {
		return ir.SetRef{}, err
	}
	var load ir.Load
	if ic.Load != nil {
		load, err = c.compileLoad(*ic.Load, sc)
		if err != nil {
			return ir.SetRef{}, err
		}
	}
	return ir.SetRef{Exercise: ic.Catalog, Target: target, Load: load}, nil
}

// compileRepeatStmt implements groups.md §3.7 (`name*N`) and §3.7a
// (`(A, B, …)*N`) — both the same repetition macro, differing only in
// whether the body comes from a declared name's own compiled Group or
// an inline, anonymous statement list.
func (c *compiler) compileRepeatStmt(rs *parser.RepeatStmt, sc *scope) (ir.Group, error) {
	var inner ir.Group
	switch rep := rs.Repeatable.(type) {
	case *parser.Ref:
		if !sc.hasName(rep.Name) {
			return ir.Group{}, fmt.Errorf("%s: undefined reference %q", rep.Pos, rep.Name)
		}
		body := sc.lookupBody(rep.Name)
		if body == nil {
			return ir.Group{}, fmt.Errorf("%s: %q has no repeatable body", rep.Pos, rep.Name)
		}
		inner = *body
	case *parser.StmtList:
		ci, err := c.compileItemList(rep.Stmts, sc, nil)
		if err != nil {
			return ir.Group{}, err
		}
		inner = ir.Group{
			Kind: "straight", Interleave: "sequential", Rest: restFromDuration(ci.restDuration),
			Termination: ir.CountTermination{N: len(ci.members)}, Atomic: false, Members: ci.members,
		}
	default:
		return ir.Group{}, fmt.Errorf("compiler: unhandled repeatable %T", rs.Repeatable)
	}

	// N copies of the same compiled Group value. Group/Member trees are
	// never mutated after construction (see expr.go's note on
	// loadExprOf), so sharing the underlying Members slice's backing
	// array across copies is safe — no deep clone needed.
	members := make([]ir.Member, rs.N)
	for i := range members {
		members[i] = inner
	}
	return ir.Group{
		Kind: "straight", Interleave: "sequential", Rest: ir.SingleRest{},
		Termination: ir.CountTermination{N: rs.N}, Atomic: false, Members: members,
	}, nil
}

func (c *compiler) compileGroupExpr(ge parser.GroupExpr, sc *scope) (ir.Group, error) {
	switch v := ge.(type) {
	case *parser.SupersetExpr:
		return c.compileSupersetExpr(v, sc)
	case *parser.CadenceExpr:
		return c.compileCadenceExpr(v, sc)
	case *parser.AmrapExpr:
		return c.compileAmrapExpr(v, sc)
	case *parser.ForTimeExpr:
		return c.compileForTimeExpr(v, sc)
	default:
		return ir.Group{}, fmt.Errorf("compiler: unhandled group expr %T", ge)
	}
}

// compileSupersetExpr implements groups.md §3.2. Round count is
// inferred from the referenced exercises' own member counts, which must
// all agree. A rest(d) argument between refs sets `intra` — per the
// spec text this is meant to apply to "exactly that adjacent pair", but
// RestPolicy's `intra` is one scalar for the whole group, so (as with
// the bare-rest simplification in compileItemList) the last one found
// wins; not exercised by any golden fixture.
func (c *compiler) compileSupersetExpr(se *parser.SupersetExpr, sc *scope) (ir.Group, error) {
	var members []ir.Member
	roundCount := -1
	intra := ir.Duration{Value: 0, Unit: "s"}
	for _, arg := range se.Args {
		if arg.Rest != nil {
			intra = toIRDuration(arg.Rest.Duration)
			continue
		}
		if !sc.hasName(arg.Ref.Name) {
			return ir.Group{}, fmt.Errorf("%s: undefined reference %q", arg.Ref.Pos, arg.Ref.Name)
		}
		body := sc.lookupBody(arg.Ref.Name)
		if body == nil {
			return ir.Group{}, fmt.Errorf("%s: %q has no members to use in a %s", arg.Ref.Pos, arg.Ref.Name, se.Kind)
		}
		n := len(body.Members)
		if roundCount == -1 {
			roundCount = n
		} else if roundCount != n {
			return ir.Group{}, fmt.Errorf("%s: %s members disagree on round count (%d vs %d)", se.Pos, se.Kind, roundCount, n)
		}
		members = append(members, ir.Ref{Name: arg.Ref.Name})
	}
	if roundCount == -1 {
		roundCount = 0
	}
	return ir.Group{
		Kind: se.Kind, Interleave: "round_robin",
		Rest:        ir.TwoLevelRest{Intra: intra, Inter: defaultSupersetInterRest},
		Termination: ir.CountTermination{N: roundCount}, Atomic: false, Members: members,
	}, nil
}

func (c *compiler) compileCadenceExpr(ce *parser.CadenceExpr, sc *scope) (ir.Group, error) {
	ci, err := c.compileItemList(ce.Body, sc, nil)
	if err != nil {
		return ir.Group{}, err
	}
	return ir.Group{
		Kind: "emom", Interleave: "round_robin", Rest: ir.RemainderRest{},
		Termination: ir.EmomTermination{Interval: toIRDuration(ce.Duration), N: ce.N},
		Atomic:      false, Members: ci.members,
	}, nil
}

func (c *compiler) compileAmrapExpr(ae *parser.AmrapExpr, sc *scope) (ir.Group, error) {
	ci, err := c.compileItemList(ae.Body, sc, nil)
	if err != nil {
		return ir.Group{}, err
	}
	return ir.Group{
		Kind: "amrap", Interleave: "round_robin", Rest: ir.AdLibRest{},
		Termination: ir.TimeCapTermination{Cap: toIRDuration(ae.Duration)}, Atomic: false, Members: ci.members,
	}, nil
}

func (c *compiler) compileForTimeExpr(fe *parser.ForTimeExpr, sc *scope) (ir.Group, error) {
	if fe.Rounds != nil {
		return c.compileRoundsExpr(fe.Rounds, sc)
	}
	ci, err := c.compileItemList(fe.Body, sc, nil)
	if err != nil {
		return ir.Group{}, err
	}
	return ir.Group{
		Kind: "for_time", Interleave: "round_robin", Rest: ir.AdLibRest{},
		Termination: ir.ForTimeTermination{}, Atomic: false, Members: ci.members,
	}, nil
}

// compileRoundsExpr implements groups.md §3.5's `rounds […] as x { }`
// repetition macro: one "rounds"-kind, 1-lap round_robin sub-group per
// value, with every occurrence of the bound name substituted by that
// value, all wrapped in a sequential "for_time"-kind outer group (laps
// run strictly one after another).
func (c *compiler) compileRoundsExpr(re *parser.RoundsExpr, sc *scope) (ir.Group, error) {
	var laps []ir.Member
	for _, v := range re.Values {
		substituted := make([]any, len(re.Body))
		for i, item := range re.Body {
			substituted[i] = substituteItem(item, re.As, float64(v))
		}
		ci, err := c.compileItemList(substituted, sc, nil)
		if err != nil {
			return ir.Group{}, err
		}
		laps = append(laps, ir.Group{
			Kind: "rounds", Interleave: "round_robin", Rest: ir.AdLibRest{},
			Termination: ir.CountTermination{N: 1}, Atomic: false, Members: ci.members,
		})
	}
	return ir.Group{
		Kind: "for_time", Interleave: "sequential", Rest: ir.AdLibRest{},
		Termination: ir.ForTimeTermination{}, Atomic: false, Members: laps,
	}, nil
}

// substituteExpr replaces every occurrence of a single-segment
// dottedPath matching `name` with a numeric literal — the compile-time
// substitution `rounds … as x` uses (spec/semantics/groups.md §3.5).
func substituteExpr(e parser.Expr, name string, value float64) parser.Expr {
	switch v := e.(type) {
	case *parser.DottedPath:
		if len(v.Segments) == 1 && v.Segments[0] == name {
			return &parser.NumberLit{Value: value, Pos: v.Pos}
		}
		return v
	case *parser.MulExpr:
		return &parser.MulExpr{Left: substituteExpr(v.Left, name, value), Right: substituteExpr(v.Right, name, value)}
	default:
		return v
	}
}

func substituteQuantity(q parser.Quantity, name string, value float64) parser.Quantity {
	return parser.Quantity{Value: substituteExpr(q.Value, name, value), Plus: q.Plus}
}

// substituteItem handles the node types attested inside a `rounds …
// as x` body (InlineCall) and passes everything else through
// unchanged — a `rounds` body nesting a further group expression with
// its own substitutable quantities isn't attested anywhere in the
// conformance corpus.
func substituteItem(item any, name string, value float64) any {
	switch v := item.(type) {
	case *parser.InlineCall:
		nc := &parser.InlineCall{Catalog: v.Catalog, Target: substituteQuantity(v.Target, name, value), Pos: v.Pos}
		if v.Load != nil {
			l := substituteQuantity(*v.Load, name, value)
			nc.Load = &l
		}
		return nc
	default:
		return v
	}
}
