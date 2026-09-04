package owl

import (
	"encoding/json"
	"errors"
)

// ErrNotImplemented is returned by every function in this file. This
// package currently only defines the canonical-form wire types
// (types.go, codec.go) — the lexer/parser/compiler/resolve/progress
// logic itself hasn't been written yet. See ../README.md § Status.
var ErrNotImplemented = errors.New("owl: not implemented yet")

// Compile parses OWL source and produces its canonical form
// (spec/canonical-form.md). It does not depend on any athlete's state.
func Compile(source string) (*Program, error) {
	return nil, ErrNotImplemented
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

// Progress evaluates every progressable SetRef a session touched against
// what was logged, and returns the athlete's updated state — see
// spec/semantics/progression.md, spec/stdlib/schemes.md.
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
