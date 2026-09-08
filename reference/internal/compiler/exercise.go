package compiler

import (
	"fmt"

	"github.com/open-workout/owl/reference/internal/ir"
	"github.com/open-workout/owl/reference/internal/parser"
)

// dropsetBuild is a dropset (named, from an explicit `dropset` decl, or
// anonymous, from `drop(...)` sugar) under construction. Its member
// SetRefs are tracked by pointer for the same reason scope's are — a
// `progress` line naming one of them attaches by mutating the pointee.
type dropsetBuild struct {
	name    string // "" for anonymous drop(...) sugar
	sets    []*ir.SetRef
	byLabel map[string]*ir.SetRef
}

func buildDropsetGroup(db *dropsetBuild) ir.Group {
	members := make([]ir.Member, len(db.sets))
	for i, p := range db.sets {
		members[i] = *p
	}
	return ir.Group{
		Kind: "dropset", Interleave: "sequential", Rest: ir.SingleRest{},
		Termination: ir.CountTermination{N: len(members)}, Atomic: true, Members: members,
	}
}

// bodyItem is one position in an exercise's declaration-order body list:
// either a flat set or a dropset (named or anonymous), never both.
type bodyItem struct {
	set     *ir.SetRef
	dropset *dropsetBuild
}

// compileExerciseDecl compiles one `exercise name = $Catalog { ... }`.
// Two passes over ed.Items: sets/dropsets are built first (in source
// order, so a later local reference like `top.weight` can see an
// earlier sibling set), then `progress` lines — which in every attested
// example follow their target set — attach ProgressionRules onto the
// already-built SetRef pointers.
func (c *compiler) compileExerciseDecl(ed *parser.ExerciseDecl, parent *scope) (*ir.ExerciseDecl, error) {
	sc := newScope(parent)

	var body []bodyItem
	var flatOrder []*ir.SetRef
	var namedDropsets []*dropsetBuild
	var progressItems []any // *parser.ProgressDecl | *parser.CondItem (wrapping ProgressDecl)

	for _, raw := range ed.Items {
		switch item := raw.(type) {
		case *parser.SetDecl:
			ref, db, err := c.buildSetDecl(item, sc, ed.Catalog)
			if err != nil {
				return nil, err
			}
			if db != nil {
				body = append(body, bodyItem{dropset: db})
			} else {
				flatOrder = append(flatOrder, ref)
				body = append(body, bodyItem{set: ref})
			}
		case *parser.DropsetDecl:
			db, err := c.buildDropsetDecl(item, sc, ed.Catalog)
			if err != nil {
				return nil, err
			}
			namedDropsets = append(namedDropsets, db)
			body = append(body, bodyItem{dropset: db})
		case *parser.ProgressDecl:
			progressItems = append(progressItems, item)
		case *parser.CondItem:
			if _, ok := item.Then.(*parser.ProgressDecl); ok {
				progressItems = append(progressItems, item)
			} else {
				return nil, fmt.Errorf("%s: a conditional %T at exercise level is not yet supported (only conditional 'progress' is — spec/semantics/conditionals.md §4)", item.Pos, item.Then)
			}
		default:
			return nil, fmt.Errorf("compiler: unexpected exercise item %T", raw)
		}
	}

	for _, raw := range progressItems {
		switch item := raw.(type) {
		case *parser.ProgressDecl:
			ref, rule, err := c.buildProgressionRule(item, sc)
			if err != nil {
				return nil, err
			}
			ref.Progression = rule
		case *parser.CondItem:
			thenPD := item.Then.(*parser.ProgressDecl)
			elsePD, ok := item.Else.(*parser.ProgressDecl)
			if !ok {
				return nil, fmt.Errorf("%s: conditional progress: else branch must also be 'progress'", item.Pos)
			}
			thenRef, thenRule, err := c.buildProgressionRule(thenPD, sc)
			if err != nil {
				return nil, err
			}
			elseRef, elseRule, err := c.buildProgressionRule(elsePD, sc)
			if err != nil {
				return nil, err
			}
			if thenRef != elseRef {
				return nil, fmt.Errorf("%s: conditional progress: then/else target different sets (%q vs %q)", item.Pos, thenRule.Target, elseRule.Target)
			}
			cond, err := c.compileCond(item.Cond, sc)
			if err != nil {
				return nil, err
			}
			thenRef.Progression = ir.ConditionalProgression{Cond: cond, Then: thenRule, Else: elseRule}
		}
	}

	sets := make([]ir.SetRef, len(flatOrder))
	for i, p := range flatOrder {
		sets[i] = *p
	}

	var groups []ir.NamedGroup
	for _, db := range namedDropsets {
		groups = append(groups, ir.NamedGroup{Name: db.name, Group: buildDropsetGroup(db)})
	}

	members := make([]ir.Member, len(body))
	for i, b := range body {
		if b.set != nil {
			members[i] = *b.set
		} else {
			members[i] = buildDropsetGroup(b.dropset)
		}
	}
	bodyGroup := ir.Group{
		Kind: "straight", Interleave: "sequential", Rest: ir.SingleRest{},
		Termination: ir.CountTermination{N: len(members)}, Atomic: false, Members: members,
	}

	return &ir.ExerciseDecl{Name: ed.Name, Catalog: ed.Catalog, Sets: sets, Groups: groups, Body: bodyGroup}, nil
}

// buildSetDecl compiles one `set` line. If it carries a `drop(...)`
// modifier, it returns (nil, dropsetBuild) instead of (*ir.SetRef, nil)
// — the caller places the whole anonymous dropset at this body
// position (groups.md §3.8's sugar rule) rather than the bare set.
// Either way, the annotated set's own label is registered directly in
// sc (not nested under a dropset name), since sugar keeps the flat
// exercise-level namespace.
func (c *compiler) buildSetDecl(sd *parser.SetDecl, sc *scope, catalog string) (*ir.SetRef, *dropsetBuild, error) {
	target, err := c.compileTarget(sd.Target, sc)
	if err != nil {
		return nil, nil, err
	}
	var load ir.Load
	if sd.Load != nil {
		load, err = c.compileLoad(*sd.Load, sc)
		if err != nil {
			return nil, nil, err
		}
	}
	ref := &ir.SetRef{Label: sd.Label, Exercise: catalog, Target: target, Load: load}
	sc.declareSet(sd.Label, ref)

	if sd.Drop == nil {
		return ref, nil, nil
	}
	wl, ok := load.(ir.WeightLoad)
	if !ok {
		return nil, nil, fmt.Errorf("%s: drop(...) requires the annotated set to have a load (spec/semantics/groups.md §3.8)", sd.Pos)
	}
	db := &dropsetBuild{byLabel: map[string]*ir.SetRef{}}
	db.sets = append(db.sets, ref)
	if sd.Label != "" {
		db.byLabel[sd.Label] = ref
	}
	for _, f := range sd.Drop.Factors {
		auto := &ir.SetRef{
			Exercise: catalog,
			Target:   ir.RepsTarget{Expr: ir.NumberExpr{Value: 1}, Plus: true},
			Load:     ir.WeightLoad{Expr: ir.MulExpr{Left: ir.NumberExpr{Value: f}, Right: wl.Expr}, Unit: wl.Unit},
		}
		db.sets = append(db.sets, auto)
	}
	return nil, db, nil
}

// buildDropsetDecl compiles an explicit `dropset name = { ... }`. Its
// member sets are compiled against a *child* scope so a local reference
// inside it (`top.weight`) can see earlier siblings within the same
// dropset, but a label declared here does NOT leak into the exercise's
// own scope — it's only reachable from outside as `name.label` (grammar
// note under dottedPath), unlike drop(...) sugar's flat namespace.
func (c *compiler) buildDropsetDecl(dd *parser.DropsetDecl, parent *scope, catalog string) (*dropsetBuild, error) {
	inner := newScope(parent)
	db := &dropsetBuild{name: dd.Name, byLabel: map[string]*ir.SetRef{}}
	for _, sd := range dd.Sets {
		ref, subDb, err := c.buildSetDecl(sd, inner, catalog)
		if err != nil {
			return nil, err
		}
		if subDb != nil {
			return nil, fmt.Errorf("%s: drop(...) sugar cannot be used on a set inside an explicit dropset", sd.Pos)
		}
		db.sets = append(db.sets, ref)
		if sd.Label != "" {
			db.byLabel[sd.Label] = ref
		}
	}
	parent.declareDropset(dd.Name, db)
	return db, nil
}

// buildProgressionRule resolves a `progress = scheme(target, ...args)`
// line: Args[0] must be a set-label reference (bare or
// dropset-qualified), the rest must be number literals.
func (c *compiler) buildProgressionRule(pd *parser.ProgressDecl, sc *scope) (*ir.SetRef, ir.ProgressionRule, error) {
	if len(pd.Args) == 0 {
		return nil, ir.ProgressionRule{}, fmt.Errorf("%s: progress %q: missing target argument", pd.Pos, pd.Scheme)
	}
	dp, ok := pd.Args[0].(*parser.DottedPath)
	if !ok {
		return nil, ir.ProgressionRule{}, fmt.Errorf("%s: progress %q: first argument must be a set label", pd.Pos, pd.Scheme)
	}
	ref, label, err := resolveLabelRef(dp, sc)
	if err != nil {
		return nil, ir.ProgressionRule{}, err
	}
	var args []float64
	for i, a := range pd.Args[1:] {
		lit, ok := a.(*parser.NumberLit)
		if !ok {
			return nil, ir.ProgressionRule{}, fmt.Errorf("%s: progress %q: argument %d must be a number literal", pd.Pos, pd.Scheme, i+2)
		}
		args = append(args, lit.Value)
	}
	return ref, ir.ProgressionRule{Scheme: pd.Scheme, Target: label, Args: args}, nil
}
