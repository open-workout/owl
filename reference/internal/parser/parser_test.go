package parser

import (
	"os"
	"testing"

	"github.com/open-workout/owl/reference/internal/lexer"
)

// newParser lexes src and returns a *parser positioned at its first
// token, for calling individual production methods directly (this test
// file is in package parser, so unexported methods are reachable) —
// most tests below exercise one grammar production in isolation rather
// than going through the top-level Parse.
func newParser(t *testing.T, src string) *parser {
	t.Helper()
	toks, err := lexer.Lex(src)
	if err != nil {
		t.Fatalf("lex(%q): %v", src, err)
	}
	return &parser{toks: toks}
}

// ---- expressions & quantities ----

func TestParseFactor_NumberNoUnit(t *testing.T) {
	p := newParser(t, "5")
	e, err := p.parseExpr()
	if err != nil {
		t.Fatal(err)
	}
	lit, ok := e.(*NumberLit)
	if !ok {
		t.Fatalf("got %T, want *NumberLit", e)
	}
	if lit.Value != 5 || lit.Unit != "" {
		t.Fatalf("got %+v", lit)
	}
}

func TestParseFactor_NumberWithUnit(t *testing.T) {
	// factor's `unit` production is not context-sensitive to
	// duration/quantity the way `duration` is — "min" stays literally
	// "min" here (not normalized to "m"); only parseDuration normalizes.
	tests := []struct {
		src  string
		want float64
		unit string
	}{
		{"100kg", 100, "kg"},
		{"60lb", 60, "lb"},
		{"100m", 100, "m"},
		{"5km", 5, "km"},
		{"10min", 10, "min"},
		{"30s", 30, "s"},
		{"1h", 1, "h"},
		{"2d", 2, "d"},
	}
	for _, tc := range tests {
		t.Run(tc.src, func(t *testing.T) {
			p := newParser(t, tc.src)
			e, err := p.parseExpr()
			if err != nil {
				t.Fatal(err)
			}
			lit, ok := e.(*NumberLit)
			if !ok {
				t.Fatalf("got %T, want *NumberLit", e)
			}
			if lit.Value != tc.want || lit.Unit != tc.unit {
				t.Fatalf("got {%v %q}, want {%v %q}", lit.Value, lit.Unit, tc.want, tc.unit)
			}
		})
	}
}

func TestParseFactor_DottedPath(t *testing.T) {
	p := newParser(t, "tm.squat.weight")
	e, err := p.parseExpr()
	if err != nil {
		t.Fatal(err)
	}
	dp, ok := e.(*DottedPath)
	if !ok {
		t.Fatalf("got %T, want *DottedPath", e)
	}
	want := []string{"tm", "squat", "weight"}
	if len(dp.Segments) != len(want) {
		t.Fatalf("got %v, want %v", dp.Segments, want)
	}
	for i := range want {
		if dp.Segments[i] != want[i] {
			t.Fatalf("got %v, want %v", dp.Segments, want)
		}
	}
}

func TestParseFactor_CatalogFieldRef(t *testing.T) {
	p := newParser(t, "$BarbellBackSquat.e1rm")
	e, err := p.parseExpr()
	if err != nil {
		t.Fatal(err)
	}
	cf, ok := e.(*CatalogFieldRef)
	if !ok {
		t.Fatalf("got %T, want *CatalogFieldRef", e)
	}
	if cf.Catalog != "BarbellBackSquat" || cf.Field != "e1rm" {
		t.Fatalf("got %+v", cf)
	}
}

func TestParseExpr_MulAssociativityAndPrecedence(t *testing.T) {
	// term ::= factor ('*' factor)* is left-associative.
	p := newParser(t, "2 * 3 * 4")
	e, err := p.parseExpr()
	if err != nil {
		t.Fatal(err)
	}
	outer, ok := e.(*MulExpr)
	if !ok {
		t.Fatalf("got %T, want *MulExpr", e)
	}
	inner, ok := outer.Left.(*MulExpr)
	if !ok {
		t.Fatalf("left operand: got %T, want *MulExpr (left-associative)", outer.Left)
	}
	if inner.Left.(*NumberLit).Value != 2 || inner.Right.(*NumberLit).Value != 3 {
		t.Fatalf("got %+v", inner)
	}
	if outer.Right.(*NumberLit).Value != 4 {
		t.Fatalf("got %+v", outer.Right)
	}
}

func TestParseFactor_Parens(t *testing.T) {
	// A parenthesized expr returns its inner node directly — parens
	// only affect how the tree nests, not the node shape.
	p := newParser(t, "(0.85 * $Foo.e1rm)")
	e, err := p.parseExpr()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := e.(*MulExpr); !ok {
		t.Fatalf("got %T, want *MulExpr", e)
	}
}

func TestParseQuantity_Plus(t *testing.T) {
	tests := []struct {
		src  string
		plus bool
	}{
		{"5", false},
		{"5+", true},
		{"tm.squat.reps+", true},
	}
	for _, tc := range tests {
		t.Run(tc.src, func(t *testing.T) {
			p := newParser(t, tc.src)
			q, err := p.parseQuantity()
			if err != nil {
				t.Fatal(err)
			}
			if q.Plus != tc.plus {
				t.Fatalf("got Plus=%v, want %v", q.Plus, tc.plus)
			}
		})
	}
}

// ---- duration (distinct vocabulary from factor's `unit`) ----

func TestParseDuration(t *testing.T) {
	tests := []struct {
		src  string
		val  float64
		unit string
	}{
		{"3m", 3, "m"},
		{"2min", 2, "m"}, // "min" normalized to "m" here, unlike factor's unit
		{"1d", 1, "d"},
		{"30s", 30, "s"},
		{"1h", 1, "h"},
	}
	for _, tc := range tests {
		t.Run(tc.src, func(t *testing.T) {
			p := newParser(t, tc.src)
			d, err := p.parseDuration()
			if err != nil {
				t.Fatal(err)
			}
			if d.Value != tc.val || d.Unit != tc.unit {
				t.Fatalf("got {%v %q}, want {%v %q}", d.Value, d.Unit, tc.val, tc.unit)
			}
		})
	}
}

func TestParseDuration_UnknownUnitIsError(t *testing.T) {
	p := newParser(t, "5x")
	if _, err := p.parseDuration(); err == nil {
		t.Fatal("expected an error for an unknown duration unit")
	}
}

// ---- fallback ----

func TestParseFallback(t *testing.T) {
	t.Run("numeric", func(t *testing.T) {
		p := newParser(t, "20")
		fb, err := p.parseFallback()
		if err != nil {
			t.Fatal(err)
		}
		if fb.Numeric == nil || fb.Numeric.Value != 20 || fb.Numeric.Plus {
			t.Fatalf("got %+v", fb)
		}
	})
	t.Run("numeric plus", func(t *testing.T) {
		p := newParser(t, "1+")
		fb, err := p.parseFallback()
		if err != nil {
			t.Fatal(err)
		}
		if fb.Numeric == nil || fb.Numeric.Value != 1 || !fb.Numeric.Plus {
			t.Fatalf("got %+v", fb)
		}
	})
	t.Run("sentinel", func(t *testing.T) {
		p := newParser(t, "AMW")
		fb, err := p.parseFallback()
		if err != nil {
			t.Fatal(err)
		}
		if fb.Numeric != nil || fb.Sentinel != "AMW" {
			t.Fatalf("got %+v", fb)
		}
	})
}

// ---- state binding ----

func TestParseStateBinding(t *testing.T) {
	p := newParser(t, "tm.squat.weight = 0.85 * $BarbellBackSquat.e1rm | AMW")
	b, err := p.parseStateBinding()
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Path.Segments) != 3 || b.Path.Segments[2] != "weight" {
		t.Fatalf("got path %v", b.Path.Segments)
	}
	if _, ok := b.Expr.(*MulExpr); !ok {
		t.Fatalf("got expr %T, want *MulExpr", b.Expr)
	}
	if b.Fallback == nil || b.Fallback.Sentinel != "AMW" {
		t.Fatalf("got fallback %+v", b.Fallback)
	}
}

func TestParseStateBinding_NoFallback(t *testing.T) {
	p := newParser(t, "tm.x = 10")
	b, err := p.parseStateBinding()
	if err != nil {
		t.Fatal(err)
	}
	if b.Fallback != nil {
		t.Fatalf("got fallback %+v, want nil", b.Fallback)
	}
}

// ---- set / progress / dropset declarations ----

func TestParseSetDecl(t *testing.T) {
	p := newParser(t, "set _ = 5 @ 100kg")
	sd, err := p.parseSetDecl()
	if err != nil {
		t.Fatal(err)
	}
	if sd.Label != "" {
		t.Fatalf("got label %q, want \"\" (surface '_')", sd.Label)
	}
	if sd.Target.Value.(*NumberLit).Value != 5 {
		t.Fatalf("got target %+v", sd.Target)
	}
	if sd.Load == nil || sd.Load.Value.(*NumberLit).Unit != "kg" {
		t.Fatalf("got load %+v", sd.Load)
	}
	if sd.Drop != nil {
		t.Fatalf("got drop %+v, want nil", sd.Drop)
	}
}

func TestParseSetDecl_NoLoad(t *testing.T) {
	p := newParser(t, "set _ = 100m")
	sd, err := p.parseSetDecl()
	if err != nil {
		t.Fatal(err)
	}
	if sd.Load != nil {
		t.Fatalf("got load %+v, want nil", sd.Load)
	}
}

func TestParseSetDecl_DropModifier(t *testing.T) {
	p := newParser(t, "set top = 12 @ tm.leg_ext.weight drop(0.8, 0.6)")
	sd, err := p.parseSetDecl()
	if err != nil {
		t.Fatal(err)
	}
	if sd.Label != "top" {
		t.Fatalf("got label %q", sd.Label)
	}
	if sd.Drop == nil {
		t.Fatal("got nil Drop")
	}
	want := []float64{0.8, 0.6}
	if len(sd.Drop.Factors) != len(want) || sd.Drop.Factors[0] != want[0] || sd.Drop.Factors[1] != want[1] {
		t.Fatalf("got factors %v, want %v", sd.Drop.Factors, want)
	}
}

func TestParseProgressDecl(t *testing.T) {
	src := `progress = {
		if top_set.reps >= 12 then
			tm.squat.weight = tm.squat.weight + 5
			tm.squat.reps = 8
		else if top_set.reps >= 8 then
			tm.squat.reps = tm.squat.reps + 1
	}`
	p := newParser(t, src)
	pd, err := p.parseProgressDecl()
	if err != nil {
		t.Fatal(err)
	}
	if len(pd.Body) != 1 {
		t.Fatalf("got %d top-level statements, want 1 (the outer if)", len(pd.Body))
	}
	outer, ok := pd.Body[0].(*ProgressIf)
	if !ok {
		t.Fatalf("got %T, want *ProgressIf", pd.Body[0])
	}
	cmp, ok := outer.Cond.(*Comparison)
	if !ok || cmp.Op != ">=" {
		t.Fatalf("got cond %+v", outer.Cond)
	}
	dp, ok := cmp.Left.(*DottedPath)
	if !ok || len(dp.Segments) != 2 || dp.Segments[0] != "top_set" || dp.Segments[1] != "reps" {
		t.Fatalf("cond.left: got %+v", cmp.Left)
	}
	if len(outer.Then) != 2 {
		t.Fatalf("got %d then-statements, want 2", len(outer.Then))
	}
	assign, ok := outer.Then[0].(*ProgressAssign)
	if !ok {
		t.Fatalf("then[0]: got %T, want *ProgressAssign", outer.Then[0])
	}
	if got := assign.Path.Segments; len(got) != 3 || got[2] != "weight" {
		t.Fatalf("then[0].Path: got %v", got)
	}
	if _, ok := assign.Expr.(*AddExpr); !ok {
		t.Fatalf("then[0].Expr: got %T, want *AddExpr", assign.Expr)
	}
	// else-if chaining: outer.Else is exactly one nested *ProgressIf.
	if len(outer.Else) != 1 {
		t.Fatalf("got %d else-statements, want 1 (nested if)", len(outer.Else))
	}
	inner, ok := outer.Else[0].(*ProgressIf)
	if !ok {
		t.Fatalf("else[0]: got %T, want *ProgressIf", outer.Else[0])
	}
	if inner.Else != nil {
		t.Fatalf("inner.Else: got %+v, want nil (omitted in source)", inner.Else)
	}
}

func TestParseProgressAssign_DropsetQualifiedLog(t *testing.T) {
	p := newParser(t, "tm.leg_ext.weight = tm.leg_ext.weight + ds.top.weight")
	a, err := p.parseProgressAssign()
	if err != nil {
		t.Fatal(err)
	}
	add, ok := a.Expr.(*AddExpr)
	if !ok {
		t.Fatalf("got %T, want *AddExpr", a.Expr)
	}
	dp, ok := add.Right.(*DottedPath)
	if !ok {
		t.Fatalf("got %T, want *DottedPath", add.Right)
	}
	want := []string{"ds", "top", "weight"}
	if len(dp.Segments) != len(want) {
		t.Fatalf("got %v, want %v", dp.Segments, want)
	}
	for i := range want {
		if dp.Segments[i] != want[i] {
			t.Fatalf("got %v, want %v", dp.Segments, want)
		}
	}
}

func TestParseDropsetDecl(t *testing.T) {
	src := `dropset ds = {
		set top = 12 @ tm.leg_ext.weight
		set _   = 1+ @ 0.8 * top.weight
		set _   = 1+ @ 0.6 * top.weight
	}`
	p := newParser(t, src)
	dd, err := p.parseDropsetDecl()
	if err != nil {
		t.Fatal(err)
	}
	if dd.Name != "ds" {
		t.Fatalf("got name %q", dd.Name)
	}
	if len(dd.Sets) != 3 {
		t.Fatalf("got %d sets, want 3", len(dd.Sets))
	}
	if dd.Sets[0].Label != "top" {
		t.Fatalf("sets[0].Label = %q, want \"top\"", dd.Sets[0].Label)
	}
	if dd.Sets[1].Label != "" {
		t.Fatalf("sets[1].Label = %q, want \"\" (surface '_')", dd.Sets[1].Label)
	}
}

// ---- conditionals, at all four grammar levels ----

func TestParseCond_AllRelOps(t *testing.T) {
	for op, tok := range map[string]string{
		"<": "<", ">": ">", "<=": "<=", ">=": ">=", "==": "==", "!=": "!=",
	} {
		t.Run(op, func(t *testing.T) {
			p := newParser(t, "tm.a "+tok+" 5")
			c, err := p.parseCond()
			if err != nil {
				t.Fatal(err)
			}
			cmp, ok := c.(*Comparison)
			if !ok {
				t.Fatalf("got %T, want *Comparison", c)
			}
			if cmp.Op != op {
				t.Fatalf("got op %q, want %q", cmp.Op, op)
			}
		})
	}
}

func TestParseCondOr_AndBindsTighterThanOr(t *testing.T) {
	// "a and b or c" must parse as "(a and b) or c", not "a and (b or c)".
	p := newParser(t, "tm.a < 1 and tm.b < 2 or tm.c < 3")
	c, err := p.parseCond()
	if err != nil {
		t.Fatal(err)
	}
	or, ok := c.(*OrCond)
	if !ok {
		t.Fatalf("got %T, want *OrCond at the top", c)
	}
	and, ok := or.Left.(*AndCond)
	if !ok {
		t.Fatalf("or.Left: got %T, want *AndCond", or.Left)
	}
	if _, ok := and.Left.(*Comparison); !ok {
		t.Fatalf("and.Left: got %T, want *Comparison", and.Left)
	}
	if _, ok := or.Right.(*Comparison); !ok {
		t.Fatalf("or.Right: got %T, want *Comparison", or.Right)
	}
}

func TestParseCondAtom_Parens(t *testing.T) {
	// Parens override the default and-tighter-than-or precedence.
	p := newParser(t, "tm.a < 1 and (tm.b < 2 or tm.c < 3)")
	c, err := p.parseCond()
	if err != nil {
		t.Fatal(err)
	}
	and, ok := c.(*AndCond)
	if !ok {
		t.Fatalf("got %T, want *AndCond at the top", c)
	}
	if _, ok := and.Right.(*OrCond); !ok {
		t.Fatalf("and.Right: got %T, want *OrCond (parenthesized)", and.Right)
	}
}

func TestParseCondItem_ExerciseLevel(t *testing.T) {
	// condExerciseItem still parses (grammar symmetry with the other
	// three levels) even though the compiler now rejects it — see
	// spec/semantics/conditionals.md §2's note. A conditional setDecl
	// branch is the shape that production predates.
	src := `if tm.squat.weight > 150 then
		set _ = 5 @ 100kg
	else set _ = 5 @ 80kg`
	p := newParser(t, src)
	item, err := p.parseExerciseItem()
	if err != nil {
		t.Fatal(err)
	}
	ci, ok := item.(*CondItem)
	if !ok {
		t.Fatalf("got %T, want *CondItem", item)
	}
	cmp, ok := ci.Cond.(*Comparison)
	if !ok || cmp.Op != ">" {
		t.Fatalf("got cond %+v", ci.Cond)
	}
	if _, ok := ci.Then.(*SetDecl); !ok {
		t.Fatalf("Then: got %T, want *SetDecl", ci.Then)
	}
	if _, ok := ci.Else.(*SetDecl); !ok {
		t.Fatalf("Else: got %T, want *SetDecl", ci.Else)
	}
}

func TestParseCondItem_DayLevel_OptionalColon(t *testing.T) {
	for _, src := range []string{
		"if tm.a < 15 then: squat; else bench;",
		"if tm.a < 15 then squat; else bench;", // colon after 'then' is cosmetic
	} {
		t.Run(src, func(t *testing.T) {
			p := newParser(t, src)
			item, err := p.parseDayItem()
			if err != nil {
				t.Fatal(err)
			}
			ci, ok := item.(*CondItem)
			if !ok {
				t.Fatalf("got %T, want *CondItem", item)
			}
			thenRef, ok := ci.Then.(*Ref)
			if !ok || thenRef.Name != "squat" {
				t.Fatalf("Then: got %+v", ci.Then)
			}
			elseRef, ok := ci.Else.(*Ref)
			if !ok || elseRef.Name != "bench" {
				t.Fatalf("Else: got %+v", ci.Else)
			}
		})
	}
}

func TestParseCondItem_BlockLevel(t *testing.T) {
	p := newParser(t, "if tm.fatigue >= 7 then: light; else heavy;")
	item, err := p.parseBlockItem()
	if err != nil {
		t.Fatal(err)
	}
	ci, ok := item.(*CondItem)
	if !ok {
		t.Fatalf("got %T, want *CondItem", item)
	}
	cmp, ok := ci.Cond.(*Comparison)
	if !ok || cmp.Op != ">=" {
		t.Fatalf("got cond %+v", ci.Cond)
	}
}

func TestParseCondItem_TopLevel(t *testing.T) {
	p := newParser(t, "if tm.injured == 1 then:\n    rehab*5;\nelse\n    main*5;")
	item, err := p.parseTopLevelItemInner()
	if err != nil {
		t.Fatal(err)
	}
	ci, ok := item.(*CondItem)
	if !ok {
		t.Fatalf("got %T, want *CondItem", item)
	}
	thenRep, ok := ci.Then.(*RepeatStmt)
	if !ok {
		t.Fatalf("Then: got %T, want *RepeatStmt", ci.Then)
	}
	if ref, ok := thenRep.Repeatable.(*Ref); !ok || ref.Name != "rehab" || thenRep.N != 5 {
		t.Fatalf("Then: got %+v", thenRep)
	}
}

func TestParseCondItem_NestedElseIf(t *testing.T) {
	// `else if ... then ... else ...` parses (grammar recurses through
	// dayItem), even though spec/semantics/conditionals.md §4 flags its
	// semantics as unconfirmed — this test only claims it parses.
	src := "if tm.a == 1 then: x; else if tm.a == 2 then: y; else z;"
	p := newParser(t, src)
	item, err := p.parseDayItem()
	if err != nil {
		t.Fatal(err)
	}
	outer := item.(*CondItem)
	inner, ok := outer.Else.(*CondItem)
	if !ok {
		t.Fatalf("Else: got %T, want a nested *CondItem", outer.Else)
	}
	cmp, ok := inner.Cond.(*Comparison)
	if !ok || cmp.Op != "==" {
		t.Fatalf("got cond %+v", inner.Cond)
	}
}

// ---- statements: repeat, rest, assign ----

func TestParseRepeatStmt_Ref(t *testing.T) {
	p := newParser(t, "leader*5")
	s, err := p.parseStmt()
	if err != nil {
		t.Fatal(err)
	}
	rs, ok := s.(*RepeatStmt)
	if !ok {
		t.Fatalf("got %T, want *RepeatStmt", s)
	}
	ref, ok := rs.Repeatable.(*Ref)
	if !ok || ref.Name != "leader" {
		t.Fatalf("got %+v", rs.Repeatable)
	}
	if rs.N != 5 {
		t.Fatalf("got N=%d, want 5", rs.N)
	}
}

func TestParseRepeatStmt_AnonymousBody(t *testing.T) {
	p := newParser(t, "(squat, rest(1m))*3")
	s, err := p.parseStmt()
	if err != nil {
		t.Fatal(err)
	}
	rs, ok := s.(*RepeatStmt)
	if !ok {
		t.Fatalf("got %T, want *RepeatStmt", s)
	}
	if rs.N != 3 {
		t.Fatalf("got N=%d, want 3", rs.N)
	}
	sl, ok := rs.Repeatable.(*StmtList)
	if !ok {
		t.Fatalf("got %T, want *StmtList", rs.Repeatable)
	}
	if len(sl.Stmts) != 2 {
		t.Fatalf("got %d stmts, want 2", len(sl.Stmts))
	}
	if _, ok := sl.Stmts[0].(*Ref); !ok {
		t.Fatalf("stmts[0]: got %T, want *Ref", sl.Stmts[0])
	}
	if _, ok := sl.Stmts[1].(*RestStmt); !ok {
		t.Fatalf("stmts[1]: got %T, want *RestStmt", sl.Stmts[1])
	}
}

func TestParseRepeatStmt_NestedAnonymousBody(t *testing.T) {
	// (leader*5, pause)*5 — a repeatStmt nested inside an anonymous
	// repeatable's own statement list.
	p := newParser(t, "(leader*5, pause)*5")
	s, err := p.parseStmt()
	if err != nil {
		t.Fatal(err)
	}
	rs := s.(*RepeatStmt)
	if rs.N != 5 {
		t.Fatalf("got outer N=%d, want 5", rs.N)
	}
	sl := rs.Repeatable.(*StmtList)
	if len(sl.Stmts) != 2 {
		t.Fatalf("got %d stmts, want 2", len(sl.Stmts))
	}
	inner, ok := sl.Stmts[0].(*RepeatStmt)
	if !ok {
		t.Fatalf("stmts[0]: got %T, want *RepeatStmt", sl.Stmts[0])
	}
	if inner.N != 5 || inner.Repeatable.(*Ref).Name != "leader" {
		t.Fatalf("got %+v", inner)
	}
	if ref, ok := sl.Stmts[1].(*Ref); !ok || ref.Name != "pause" {
		t.Fatalf("stmts[1]: got %+v", sl.Stmts[1])
	}
}

func TestParseRestStmt(t *testing.T) {
	p := newParser(t, "rest(1d)")
	s, err := p.parseStmt()
	if err != nil {
		t.Fatal(err)
	}
	rs, ok := s.(*RestStmt)
	if !ok {
		t.Fatalf("got %T, want *RestStmt", s)
	}
	if rs.Duration.Value != 1 || rs.Duration.Unit != "d" {
		t.Fatalf("got %+v", rs.Duration)
	}
}

func TestParseAssignStmt(t *testing.T) {
	p := newParser(t, "partC = amrap(12m) { $KettlebellSwing(15 @ 24kg) }")
	s, err := p.parseStmt()
	if err != nil {
		t.Fatal(err)
	}
	as, ok := s.(*AssignStmt)
	if !ok {
		t.Fatalf("got %T, want *AssignStmt", s)
	}
	if as.Name != "partC" {
		t.Fatalf("got name %q", as.Name)
	}
	if _, ok := as.Group.(*AmrapExpr); !ok {
		t.Fatalf("got group %T, want *AmrapExpr", as.Group)
	}
}

func TestParseStmt_BareRefVsAssignVsRepeat(t *testing.T) {
	// Same leading IDENT token; the parser must look ahead correctly to
	// pick the right production.
	tests := []struct {
		src  string
		want string
	}{
		{"squat", "*parser.Ref"},
		{"squat*5", "*parser.RepeatStmt"},
		{"squat = amrap(1m) { $X(5) }", "*parser.AssignStmt"},
	}
	for _, tc := range tests {
		t.Run(tc.src, func(t *testing.T) {
			p := newParser(t, tc.src)
			s, err := p.parseStmt()
			if err != nil {
				t.Fatal(err)
			}
			got := goType(s)
			if got != tc.want {
				t.Fatalf("got %s, want %s", got, tc.want)
			}
		})
	}
}

func goType(v any) string {
	switch v.(type) {
	case *Ref:
		return "*parser.Ref"
	case *RepeatStmt:
		return "*parser.RepeatStmt"
	case *AssignStmt:
		return "*parser.AssignStmt"
	default:
		return "unknown"
	}
}

// ---- group expressions ----

func TestParseSupersetExpr(t *testing.T) {
	p := newParser(t, "superset(bench, row)")
	g, err := p.parseGroupExpr()
	if err != nil {
		t.Fatal(err)
	}
	se, ok := g.(*SupersetExpr)
	if !ok {
		t.Fatalf("got %T, want *SupersetExpr", g)
	}
	if se.Kind != "superset" {
		t.Fatalf("got kind %q", se.Kind)
	}
	if len(se.Args) != 2 || se.Args[0].Ref.Name != "bench" || se.Args[1].Ref.Name != "row" {
		t.Fatalf("got args %+v", se.Args)
	}
}

func TestParseSupersetExpr_CircuitWithRestArg(t *testing.T) {
	p := newParser(t, "circuit(a, rest(20s), b)")
	g, err := p.parseGroupExpr()
	if err != nil {
		t.Fatal(err)
	}
	se := g.(*SupersetExpr)
	if se.Kind != "circuit" {
		t.Fatalf("got kind %q", se.Kind)
	}
	if len(se.Args) != 3 {
		t.Fatalf("got %d args, want 3", len(se.Args))
	}
	if se.Args[1].Rest == nil || se.Args[1].Rest.Duration.Value != 20 {
		t.Fatalf("args[1]: got %+v", se.Args[1])
	}
}

func TestParseCadenceExpr(t *testing.T) {
	t.Run("bare emom defaults N to 1", func(t *testing.T) {
		p := newParser(t, "emom(3m){ back_squat }")
		g, err := p.parseGroupExpr()
		if err != nil {
			t.Fatal(err)
		}
		ce := g.(*CadenceExpr)
		if ce.N != 1 {
			t.Fatalf("got N=%d, want 1", ce.N)
		}
		if ce.Duration.Value != 3 || ce.Duration.Unit != "m" {
			t.Fatalf("got duration %+v", ce.Duration)
		}
		if len(ce.Body) != 1 {
			t.Fatalf("got %d body items, want 1", len(ce.Body))
		}
	})
	t.Run("N*emom", func(t *testing.T) {
		p := newParser(t, "5*emom(3m){ back_squat }")
		s, err := p.parseStmt()
		if err != nil {
			t.Fatal(err)
		}
		ce, ok := s.(*CadenceExpr)
		if !ok {
			t.Fatalf("got %T, want *CadenceExpr", s)
		}
		if ce.N != 5 {
			t.Fatalf("got N=%d, want 5", ce.N)
		}
	})
}

func TestParseAmrapExpr(t *testing.T) {
	p := newParser(t, "amrap(12m) { $KettlebellSwing(15 @ 24kg) $BoxJumpOver(12) $ToesToBar(9) }")
	g, err := p.parseGroupExpr()
	if err != nil {
		t.Fatal(err)
	}
	ae := g.(*AmrapExpr)
	if ae.Duration.Value != 12 || ae.Duration.Unit != "m" {
		t.Fatalf("got duration %+v", ae.Duration)
	}
	if len(ae.Body) != 3 {
		t.Fatalf("got %d body items, want 3", len(ae.Body))
	}
	for i, item := range ae.Body {
		if _, ok := item.(*InlineCall); !ok {
			t.Fatalf("body[%d]: got %T, want *InlineCall", i, item)
		}
	}
}

func TestParseForTimeExpr_FlatBody(t *testing.T) {
	p := newParser(t, "for_time { $A(5); $B(10) }")
	g, err := p.parseGroupExpr()
	if err != nil {
		t.Fatal(err)
	}
	fe := g.(*ForTimeExpr)
	if fe.Rounds != nil {
		t.Fatalf("got Rounds %+v, want nil", fe.Rounds)
	}
	if len(fe.Body) != 2 {
		t.Fatalf("got %d body items, want 2", len(fe.Body))
	}
}

func TestParseForTimeExpr_Rounds(t *testing.T) {
	src := `for_time {
		rounds [21, 15, 9] as n {
			$BarbellThruster( n @ 42.5)
			$PullUp( n )
		}
	}`
	p := newParser(t, src)
	g, err := p.parseGroupExpr()
	if err != nil {
		t.Fatal(err)
	}
	fe := g.(*ForTimeExpr)
	if fe.Rounds == nil {
		t.Fatal("got nil Rounds")
	}
	if fe.Body != nil {
		t.Fatalf("got Body %+v, want nil when Rounds is set", fe.Body)
	}
	want := []int{21, 15, 9}
	if len(fe.Rounds.Values) != len(want) {
		t.Fatalf("got values %v, want %v", fe.Rounds.Values, want)
	}
	for i := range want {
		if fe.Rounds.Values[i] != want[i] {
			t.Fatalf("got values %v, want %v", fe.Rounds.Values, want)
		}
	}
	if fe.Rounds.As != "n" {
		t.Fatalf("got As %q, want \"n\"", fe.Rounds.As)
	}
	if len(fe.Rounds.Body) != 2 {
		t.Fatalf("got %d body items, want 2", len(fe.Rounds.Body))
	}
	// The bound variable is NOT substituted at parse time — that's the
	// compiler's job (spec/semantics/groups.md §3.5).
	thruster := fe.Rounds.Body[0].(*InlineCall)
	dp, ok := thruster.Target.Value.(*DottedPath)
	if !ok || len(dp.Segments) != 1 || dp.Segments[0] != "n" {
		t.Fatalf("got target %+v, want an unsubstituted DottedPath{\"n\"}", thruster.Target.Value)
	}
}

func TestParseInlineCall(t *testing.T) {
	t.Run("with load", func(t *testing.T) {
		p := newParser(t, "$KettlebellSwing(15 @ 24kg)")
		s, err := p.parseStmt()
		if err != nil {
			t.Fatal(err)
		}
		ic := s.(*InlineCall)
		if ic.Catalog != "KettlebellSwing" {
			t.Fatalf("got catalog %q", ic.Catalog)
		}
		if ic.Target.Value.(*NumberLit).Value != 15 {
			t.Fatalf("got target %+v", ic.Target)
		}
		if ic.Load == nil || ic.Load.Value.(*NumberLit).Value != 24 {
			t.Fatalf("got load %+v", ic.Load)
		}
	})
	t.Run("target only", func(t *testing.T) {
		p := newParser(t, "$PullUp(n)")
		s, err := p.parseStmt()
		if err != nil {
			t.Fatal(err)
		}
		ic := s.(*InlineCall)
		if ic.Load != nil {
			t.Fatalf("got load %+v, want nil", ic.Load)
		}
		if _, ok := ic.Target.Value.(*DottedPath); !ok {
			t.Fatalf("got target %T, want *DottedPath", ic.Target.Value)
		}
	})
}

// ---- semicolons are always optional ----

func TestParse_SemicolonsOptional(t *testing.T) {
	withSemi := "state = {};units = \"kg\";plates = [1];block a = {day d = {exercise e = $X {set _ = 5;};e;};};"
	withoutSemi := "state = {}\nunits = \"kg\"\nplates = [1]\nblock a = {\nday d = {\nexercise e = $X {\nset _ = 5\n}\ne\n}\n}"
	for name, src := range map[string]string{"with-semicolons": withSemi, "newlines-only": withoutSemi} {
		t.Run(name, func(t *testing.T) {
			toks, err := lexer.Lex(src)
			if err != nil {
				t.Fatal(err)
			}
			prog, err := Parse(toks)
			if err != nil {
				t.Fatal(err)
			}
			if prog.Units != "kg" || len(prog.Items) != 1 {
				t.Fatalf("got %+v", prog)
			}
		})
	}
}

// ---- full program, against a real fixture ----

func TestParse_MinimalStraightSetFixture(t *testing.T) {
	src, err := os.ReadFile("../../../conformance/parse/minimal-straight-set/source.owl")
	if err != nil {
		t.Fatal(err)
	}
	toks, err := lexer.Lex(string(src))
	if err != nil {
		t.Fatal(err)
	}
	prog, err := Parse(toks)
	if err != nil {
		t.Fatal(err)
	}
	if prog.Units != "kg" {
		t.Fatalf("got units %q", prog.Units)
	}
	if len(prog.Plates) != 7 {
		t.Fatalf("got %d plates, want 7", len(prog.Plates))
	}
	if len(prog.State) != 0 {
		t.Fatalf("got %d state bindings, want 0", len(prog.State))
	}
	if len(prog.Items) != 1 {
		t.Fatalf("got %d top-level items, want 1", len(prog.Items))
	}
	blk, ok := prog.Items[0].(*BlockDecl)
	if !ok {
		t.Fatalf("got %T, want *BlockDecl", prog.Items[0])
	}
	if blk.Name != "main" || len(blk.Items) != 1 {
		t.Fatalf("got %+v", blk)
	}
	day, ok := blk.Items[0].(*DayDecl)
	if !ok {
		t.Fatalf("got %T, want *DayDecl", blk.Items[0])
	}
	if day.Name != "day1" || len(day.Items) != 2 { // exercise decl + bare `squat;` ref
		t.Fatalf("got %+v", day)
	}
	ex, ok := day.Items[0].(*ExerciseDecl)
	if !ok || ex.Name != "squat" || ex.Catalog != "BarbellBackSquat" {
		t.Fatalf("got %+v", day.Items[0])
	}
	ref, ok := day.Items[1].(*Ref)
	if !ok || ref.Name != "squat" {
		t.Fatalf("got %+v", day.Items[1])
	}
}

// ---- errors ----

func TestParse_Errors(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{"block missing '='", "block main { }"},
		{"unterminated block", "block main = { day d = { }"},
		{"cond missing relop", "if tm.a then x else y"},
		{"set missing '='", "set _ 5 @ 100kg"},
		{"unknown top-level token", "42"},
		{"progress missing '('", "progress = double top_set, 8, 12, 5)"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			toks, err := lexer.Lex(tc.src)
			if err != nil {
				return // a lex error is also an acceptable failure mode here
			}
			if _, err := Parse(toks); err == nil {
				t.Fatalf("expected a parse error for %q", tc.src)
			}
		})
	}
}

func TestParse_ErrorIsSyntaxError(t *testing.T) {
	toks, err := lexer.Lex("block main { }")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Parse(toks)
	if err == nil {
		t.Fatal("expected an error")
	}
	if _, ok := err.(*SyntaxError); !ok {
		t.Fatalf("got error type %T, want *SyntaxError", err)
	}
}
