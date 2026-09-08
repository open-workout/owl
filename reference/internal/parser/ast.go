// Package parser implements a hand-written recursive-descent parser for
// OWL source per spec/grammar.ebnf, producing a source-shaped AST. It
// does not resolve names or expand macros — that's internal/compiler's
// job. Item lists use `any` rather than a proliferation of narrow marker
// interfaces (grammar.ebnf's topLevelItem/blockItem/dayItem/
// exerciseItem alternatives largely reuse the same handful of concrete
// node types); the compiler type-switches on the concrete types below.
package parser

import "github.com/open-workout/owl/reference/internal/lexer"

// Program is the root of a parsed .owl file.
type Program struct {
	State  []StateBinding
	Units  string
	Plates []float64
	// Items holds every topLevelItem that isn't a stateSection/unitsDecl/
	// platesDecl: *BlockDecl, *CondItem, or a stmt variant, in source
	// order (this order becomes Program.Sequence, spec/canonical-form.md).
	Items []any
}

type StateBinding struct {
	Path     DottedPath
	Expr     Expr
	Fallback *Fallback
}

// Fallback is either a numeric default or a sentinel name — exactly one
// of Numeric/Sentinel is set.
type Fallback struct {
	Numeric  *NumericFallback
	Sentinel string
}

type NumericFallback struct {
	Value float64
	Plus  bool
}

// ---- Declarations ----

type BlockDecl struct {
	Name string
	// Items: *DayDecl | *CondItem | stmt variant.
	Items []any
	Pos   lexer.Pos
}

type DayDecl struct {
	Name string
	// Items: *ExerciseDecl | *CondItem | stmt variant.
	Items []any
	Pos   lexer.Pos
}

type ExerciseDecl struct {
	Name    string
	Catalog string
	// Items: *SetDecl | *ProgressDecl | *DropsetDecl | *CondItem.
	Items []any
	Pos   lexer.Pos
}

type SetDecl struct {
	Label  string // "" for surface '_'
	Target Quantity
	Load   *Quantity // nil if no '@ quantity'
	Drop   *DropModifier
	Pos    lexer.Pos
}

type DropModifier struct {
	Factors []float64
}

// ProgressDecl is `progress = { ... }` — at most one per exercise
// (spec/semantics/progression.md §2), a small code block rather than a
// call into a named scheme.
type ProgressDecl struct {
	Body []any // ProgressStmt variants: *ProgressAssign | *ProgressIf
	Pos  lexer.Pos
}

// ProgressAssign is one progressBody statement: `<state path> = <expr>`.
type ProgressAssign struct {
	Path DottedPath
	Expr Expr
	Pos  lexer.Pos
}

// ProgressIf is progressBody's own `if`/`then`/`else` (grammar
// `progressIf`) — unlike CondItem, `Else` may be nil (the only place in
// the language `else` is optional), and Then/Else are statement lists,
// not single items.
type ProgressIf struct {
	Cond Cond
	Then []any // ProgressStmt variants
	Else []any // ProgressStmt variants; nil if omitted
	Pos  lexer.Pos
}

type DropsetDecl struct {
	Name string
	Sets []*SetDecl
	Pos  lexer.Pos
}

// CondItem is `if cond then[:] A else B`, used identically at the
// exerciseItem/dayItem/blockItem/topLevelItem grammar levels (Then/Else
// hold whatever single item that level's item list holds).
type CondItem struct {
	Cond Cond
	Then any
	Else any
	Pos  lexer.Pos
}

// Cond is grammar's condExpr: a single relational comparison, or two
// conditions combined with 'and'/'or'. Shared verbatim by CondItem and
// ProgressIf.
type Cond interface{ isCond() }

type Comparison struct {
	Op    string // "<" | ">" | "<=" | ">=" | "==" | "!="
	Left  Expr
	Right Expr
}

// AndCond/OrCond: 'and' binds tighter than 'or' — see parseCond.
type AndCond struct{ Left, Right Cond }
type OrCond struct{ Left, Right Cond }

func (*Comparison) isCond() {}
func (*AndCond) isCond()    {}
func (*OrCond) isCond()     {}

// ---- Statements ----
// A stmt variant (AssignStmt, RepeatStmt, RestStmt, a GroupExpr,
// InlineCall, Ref) is valid wherever grammar.ebnf's `stmt` is: as a
// topLevelItem, blockItem, dayItem, or a groupBody entry.

type AssignStmt struct {
	Name  string
	Group GroupExpr
	Pos   lexer.Pos
}

type RepeatStmt struct {
	Repeatable Repeatable // *Ref | *StmtList
	N          int
	Pos        lexer.Pos
}

// Repeatable is *Ref (a declared name) or *StmtList (an inline,
// anonymous body) — grammar.ebnf's `repeatable` production.
type Repeatable interface{ isRepeatable() }

// StmtList is a parenthesized, comma-separated statement list used only
// as a RepeatStmt's anonymous body: '(' stmt (',' stmt)* ')'.
type StmtList struct{ Stmts []any }

func (*Ref) isRepeatable()      {}
func (*StmtList) isRepeatable() {}

type RestStmt struct {
	Duration Duration
	Pos      lexer.Pos
}

type Ref struct {
	Name string
	Pos  lexer.Pos
}

type InlineCall struct {
	Catalog string
	Target  Quantity
	Load    *Quantity
	Pos     lexer.Pos
}

// ---- Group expressions ----

type GroupExpr interface{ isGroupExpr() }

type SupersetExpr struct {
	Kind string // "superset" | "circuit"
	Args []SupersetArg
	Pos  lexer.Pos
}

// SupersetArg is exactly one of Ref/Rest, per grammar's `supersetArg ::=
// ref | restStmt`.
type SupersetArg struct {
	Ref  *Ref
	Rest *RestStmt
}

type CadenceExpr struct {
	N        int // '[' NUMBER '*' ']' prefix; defaults to 1
	Duration Duration
	Body     []any
	Pos      lexer.Pos
}

type AmrapExpr struct {
	Duration Duration
	Body     []any
	Pos      lexer.Pos
}

type ForTimeExpr struct {
	Rounds *RoundsExpr // non-nil for the `rounds […] as x {}` shape
	Body   []any       // used when Rounds == nil (flat statement list)
	Pos    lexer.Pos
}

type RoundsExpr struct {
	Values []int
	As     string
	Body   []any
	Pos    lexer.Pos
}

func (*SupersetExpr) isGroupExpr() {}
func (*CadenceExpr) isGroupExpr()  {}
func (*AmrapExpr) isGroupExpr()    {}
func (*ForTimeExpr) isGroupExpr()  {}

// ---- Expressions & quantities ----

// Quantity is grammar's `quantity ::= term '+'?`.
type Quantity struct {
	Value Expr
	Plus  bool
}

type Expr interface{ isExpr() }

type NumberLit struct {
	Value float64
	// Unit is "" (dimensionless) or a unit/duration-unit string already
	// disambiguated by the parser (grammar position decides 'm' vs
	// 'min' vs bare) — see parser.go's parseQuantityUnit/parseDurationUnit.
	Unit string
	Pos  lexer.Pos
}

// DottedPath is grammar's `dottedPath ::= IDENT ('.' IDENT)*` — parsed
// as pure syntax. The compiler classifies it (state path / local
// set-field ref / dropset-qualified label) via scope lookup; see
// grammar.ebnf's note under dottedPath.
type DottedPath struct {
	Segments []string
	Pos      lexer.Pos
}

type CatalogFieldRef struct {
	Catalog string
	Field   string
	Pos     lexer.Pos
}

type MulExpr struct{ Left, Right Expr }
type DivExpr struct{ Left, Right Expr }
type AddExpr struct{ Left, Right Expr }
type SubExpr struct{ Left, Right Expr }

func (*NumberLit) isExpr()       {}
func (*DottedPath) isExpr()      {}
func (*CatalogFieldRef) isExpr() {}
func (*MulExpr) isExpr()         {}
func (*DivExpr) isExpr()         {}
func (*AddExpr) isExpr()         {}
func (*SubExpr) isExpr()         {}

type Duration struct {
	Value float64
	Unit  string // "s" | "m" | "h" | "d" (already normalized; "min" -> "m")
}
