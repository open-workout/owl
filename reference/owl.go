package owl

import (
	"encoding/json"
	"errors"

	"github.com/open-workout/owl/reference/internal/compiler"
	"github.com/open-workout/owl/reference/internal/lexer"
	"github.com/open-workout/owl/reference/internal/parser"
)

// ErrNotImplemented is returned by Resolve and Progress, which aren't
// implemented yet — see ../README.md § Status. Compile (lexer -> parser
// -> compiler) is implemented.
var ErrNotImplemented = errors.New("owl: not implemented yet")

// Compile parses OWL source and produces its canonical form
// (spec/canonical-form.md). It does not depend on any athlete's state.
func Compile(source string) (*Program, error) {
	toks, err := lexer.Lex(source)
	if err != nil {
		return nil, err
	}
	ast, err := parser.Parse(toks)
	if err != nil {
		return nil, err
	}
	return compiler.Compile(ast)
}

// Cursor selects which occurrence of a Program to resolve — see
// spec/semantics/resolution.md § Inputs.
type Cursor struct {
	Block     string
	Iteration int // 1-based; defaults to 1
	Day       string
}

// AthleteState is this athlete's current recorded values: `state`
// bindings already written by a previous Progress call, plus per-exercise
// catalog history (e.g. e1rm). See spec/semantics/resolution.md,
// spec/semantics/parameters.md.
type AthleteState struct {
	Bindings      map[string]float64 // dotted state path -> value
	CatalogFields map[CatalogFieldKey]float64
}

type CatalogFieldKey struct {
	Catalog string
	Field   string
}

// Resolve turns one occurrence of a compiled Program into a concrete,
// executable Session for a specific athlete — see
// spec/semantics/resolution.md.
func Resolve(program *Program, state AthleteState, cursor Cursor) (*Day, error) {
	return nil, ErrNotImplemented
}

// SetLog is what an athlete actually recorded for one resolved SetRef.
type SetLog struct {
	Exercise        string
	Label           string
	PerformedTarget float64
	PerformedLoad   *float64
}

// Progress runs each exercise's `progress` code against what was logged
// for that session, and returns the athlete's updated state — see
// spec/semantics/progression.md.
func Progress(program *Program, state AthleteState, log []SetLog) (AthleteState, error) {
	return AthleteState{}, ErrNotImplemented
}

// ParseCanonicalJSON decodes a canonical-form JSON document (as produced
// by Compile, or read from a conformance/parse/*/expected.json fixture)
// into a Program. Unlike Compile, this does not parse OWL source — it
// only decodes JSON already in canonical form.
func ParseCanonicalJSON(data []byte) (*Program, error) {
	var p Program
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
