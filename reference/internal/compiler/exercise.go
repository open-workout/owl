package compiler

import (
	"fmt"
	"strings"

	"github.com/open-workout/owl/reference/internal/ir"
	"github.com/open-workout/owl/reference/internal/parser"
)

// restpauseRestDuration is the fixed intra-burst rest for the explicit
// `restpause NAME = { ... }` declaration (spec/semantics/groups.md
// §3.9) — unlike the `rest_pause(duration, ...)` sugar, the explicit
// form doesn't carry its own duration, so this is the one place it's
// decided.
var restpauseRestDuration = ir.Duration{Value: 15, Unit: "s"}

// dropsetBuild is an atomic sequential group (a dropset or a restpause
// group — see `kind`) under construction, named (from an explicit
// `dropset`/`restpause` decl) or anonymous (from `drop(...)`/
// `rest_pause(...)` sugar). Its member SetRefs are tracked by pointer
// for the same reason scope's are — a `progress` line naming one of
// them attaches by mutating the pointee. Both constructs share this one
// builder (and `scope`'s `dropsets` map/`declareDropset`/
// `lookupDropset`) since they compile to the identical `Group` shape,
// differing only in `kind` and `rest.single`'s duration
// (spec/semantics/groups.md §2).
type dropsetBuild struct {
	name string // "" for anonymous sugar
	kind string // "dropset" | "restpause"
	// restDuration is the rest.single duration for a "restpause" build;
	// unused (nil means "omitted", per dropset's existing 0-value rest)
	// for "dropset".
	restDuration *ir.Duration
	sets         []*ir.SetRef
	byLabel      map[string]*ir.SetRef
}

func buildDropsetGroup(db *dropsetBuild) ir.Group {
	members := make([]ir.Member, len(db.sets))
	for i, p := range db.sets {
		members[i] = *p
	}
	return ir.Group{
		Kind: db.kind, Interleave: "sequential", Rest: ir.SingleRest{Duration: db.restDuration},
		Termination: ir.CountTermination{N: len(members)}, Atomic: true, Members: members,
	}
}

// bodyItem is one position in an exercise's declaration-order body list:
// either a flat set or a dropset/restpause group (named or anonymous),
// never both.
type bodyItem struct {
	set     *ir.SetRef
	dropset *dropsetBuild
}

// compileExerciseDecl compiles one `exercise name = $Catalog { ... }`.
// Two passes over ed.Items: sets/dropsets are built first (in source
// order, so a later local reference like `top.weight` can see an
// earlier sibling set), then the exercise's single `progress` block (if
// any) is compiled against the now-complete scope — every set it's
// declared, so a `top_set.reps` log read works regardless of whether
// `progress` came before or after `top_set` in source.
func (c *compiler) compileExerciseDecl(ed *parser.ExerciseDecl, parent *scope) (*ir.ExerciseDecl, error) {
	sc := newScope(parent)

	var body []bodyItem
	var flatOrder []*ir.SetRef
	var namedDropsets []*dropsetBuild
	var progressDecl *parser.ProgressDecl

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
		case *parser.RestpauseDecl:
			db, err := c.buildRestpauseDecl(item, sc, ed.Catalog)
			if err != nil {
				return nil, err
			}
			namedDropsets = append(namedDropsets, db)
			body = append(body, bodyItem{dropset: db})
		case *parser.ProgressDecl:
			if progressDecl != nil {
				return nil, fmt.Errorf("%s: exercise %q already has a 'progress' block (at most one per exercise — spec/semantics/progression.md §2)", item.Pos, ed.Name)
			}
			progressDecl = item
		case *parser.CondItem:
			return nil, fmt.Errorf("%s: a conditional %T at exercise level is not yet supported (spec/semantics/conditionals.md §2's note) — put the condition inside 'progress = { ... }' instead", item.Pos, item.Then)
		default:
			return nil, fmt.Errorf("compiler: unexpected exercise item %T", raw)
		}
	}

	var progress *ir.ProgressionBody
	if progressDecl != nil {
		var err error
		progress, err = c.compileProgressionBody(progressDecl, sc)
		if err != nil {
			return nil, err
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

	return &ir.ExerciseDecl{Name: ed.Name, Catalog: ed.Catalog, Sets: sets, Groups: groups, Body: bodyGroup, Progress: progress}, nil
}

// buildSetDecl compiles one `set` line. If it carries a `drop(...)` or
// `rest_pause(...)` modifier, it returns (nil, dropsetBuild) instead of
// (*ir.SetRef, nil) — the caller places the whole anonymous
// dropset/restpause group at this body position (groups.md §3.8/§3.9's
// sugar rules) rather than the bare set. Either way, the annotated
// set's own label is registered directly in sc (not nested under a
// group name), since sugar keeps the flat exercise-level namespace.
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

	switch {
	case sd.Drop != nil:
		wl, ok := load.(ir.WeightLoad)
		if !ok {
			return nil, nil, fmt.Errorf("%s: drop(...) requires the annotated set to have a load (spec/semantics/groups.md §3.8)", sd.Pos)
		}
		db := &dropsetBuild{kind: "dropset", byLabel: map[string]*ir.SetRef{}}
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
	case sd.RestPause != nil:
		if sd.RestPause.Bursts < 1 {
			return nil, nil, fmt.Errorf("%s: rest_pause(...) burst count must be at least 1", sd.Pos)
		}
		d := toIRDuration(sd.RestPause.Duration)
		db := &dropsetBuild{kind: "restpause", restDuration: &d, byLabel: map[string]*ir.SetRef{}}
		db.sets = append(db.sets, ref)
		if sd.Label != "" {
			db.byLabel[sd.Label] = ref
		}
		for i := 0; i < sd.RestPause.Bursts; i++ {
			// Same target/load as the annotated set — rest-pause holds
			// the load fixed across bursts (spec/semantics/groups.md
			// §2), unlike dropset's descending per-burst loads. Reusing
			// the same Target/Load values is safe: Group/Member trees
			// are never mutated after construction (see expr.go's note
			// on loadExprOf).
			auto := &ir.SetRef{Exercise: catalog, Target: target, Load: load}
			db.sets = append(db.sets, auto)
		}
		return nil, db, nil
	default:
		return ref, nil, nil
	}
}

// buildDropsetDecl compiles an explicit `dropset name = { ... }`. Its
// member sets are compiled against a *child* scope so a local reference
// inside it (`top.weight`) can see earlier siblings within the same
// dropset, but a label declared here does NOT leak into the exercise's
// own scope — it's only reachable from outside as `name.label` (grammar
// note under dottedPath), unlike drop(...) sugar's flat namespace.
func (c *compiler) buildDropsetDecl(dd *parser.DropsetDecl, parent *scope, catalog string) (*dropsetBuild, error) {
	inner := newScope(parent)
	db := &dropsetBuild{name: dd.Name, kind: "dropset", byLabel: map[string]*ir.SetRef{}}
	for _, sd := range dd.Sets {
		ref, subDb, err := c.buildSetDecl(sd, inner, catalog)
		if err != nil {
			return nil, err
		}
		if subDb != nil {
			return nil, fmt.Errorf("%s: drop(...)/rest_pause(...) sugar cannot be used on a set inside an explicit dropset", sd.Pos)
		}
		db.sets = append(db.sets, ref)
		if sd.Label != "" {
			db.byLabel[sd.Label] = ref
		}
	}
	parent.declareDropset(dd.Name, db)
	return db, nil
}

// buildRestpauseDecl compiles an explicit `restpause name = { ... }`
// (spec/semantics/groups.md §3.9) — same scoping treatment as
// buildDropsetDecl, fixed 15s rest.single duration
// (restpauseRestDuration; the sugar form, unlike this one, lets the
// author pick a duration).
func (c *compiler) buildRestpauseDecl(rd *parser.RestpauseDecl, parent *scope, catalog string) (*dropsetBuild, error) {
	inner := newScope(parent)
	d := restpauseRestDuration
	db := &dropsetBuild{name: rd.Name, kind: "restpause", restDuration: &d, byLabel: map[string]*ir.SetRef{}}
	for _, sd := range rd.Sets {
		ref, subDb, err := c.buildSetDecl(sd, inner, catalog)
		if err != nil {
			return nil, err
		}
		if subDb != nil {
			return nil, fmt.Errorf("%s: drop(...)/rest_pause(...) sugar cannot be used on a set inside an explicit restpause group", sd.Pos)
		}
		db.sets = append(db.sets, ref)
		if sd.Label != "" {
			db.byLabel[sd.Label] = ref
		}
	}
	parent.declareDropset(rd.Name, db)
	return db, nil
}

// compileProgressionBody compiles an exercise's `progress = { ... }`
// block (spec/semantics/progression.md §2) against a scope that already
// has every set/dropset in the exercise registered.
func (c *compiler) compileProgressionBody(pd *parser.ProgressDecl, sc *scope) (*ir.ProgressionBody, error) {
	stmts, err := c.compileProgressStmtList(pd.Body, sc)
	if err != nil {
		return nil, err
	}
	return &ir.ProgressionBody{Stmts: stmts}, nil
}

func (c *compiler) compileProgressStmtList(items []any, sc *scope) ([]ir.ProgressionStmt, error) {
	if items == nil {
		return nil, nil
	}
	stmts := make([]ir.ProgressionStmt, len(items))
	for i, raw := range items {
		s, err := c.compileProgressStmt(raw, sc)
		if err != nil {
			return nil, err
		}
		stmts[i] = s
	}
	return stmts, nil
}

func (c *compiler) compileProgressStmt(raw any, sc *scope) (ir.ProgressionStmt, error) {
	switch item := raw.(type) {
	case *parser.ProgressAssign:
		path := strings.Join(item.Path.Segments, ".")
		if !c.stateRoots[item.Path.Segments[0]] {
			return nil, fmt.Errorf("%s: progress assignment target %q is not a declared state path", item.Pos, path)
		}
		expr, err := c.compileProgressExpr(item.Expr, sc)
		if err != nil {
			return nil, err
		}
		return ir.Assign{Path: path, Expr: expr}, nil
	case *parser.ProgressIf:
		cond, err := c.compileCond(item.Cond, sc, c.compileProgressExpr)
		if err != nil {
			return nil, err
		}
		then, err := c.compileProgressStmtList(item.Then, sc)
		if err != nil {
			return nil, err
		}
		els, err := c.compileProgressStmtList(item.Else, sc)
		if err != nil {
			return nil, err
		}
		return ir.ProgressionIf{Cond: cond, Then: then, Else: els}, nil
	default:
		return nil, fmt.Errorf("compiler: unexpected progress statement %T", raw)
	}
}
