package owl

// JSON encoding/decoding for the sum-typed fields in types.go (Expr,
// Target, Load, Member, RestPolicy, Termination) — encoding/json can't
// unmarshal into a Go interface field on its own, so every such field
// needs a decode function here, dispatching on the wire discriminator
// ("type"/"kind"/"mode") described in ../spec/canonical-form.schema.json.
//
// Marshaling a value that already holds one of these interfaces works
// via the ordinary encoding/json path: each concrete type below
// implements MarshalJSON, and encoding/json calls it automatically for
// any interface-typed field/slice element holding that concrete type —
// so Program, Block, Day, etc. need no custom MarshalJSON of their own.
// Only types whose *own* fields are interfaces (StateBinding, SetRef,
// Group) need a custom UnmarshalJSON.

import (
	"encoding/json"
	"fmt"
)

// ---- Expr ----

type exprWire struct {
	Type    string          `json:"type"`
	Value   *float64        `json:"value,omitempty"`
	Path    string          `json:"path,omitempty"`
	Catalog string          `json:"catalog,omitempty"`
	Field   string          `json:"field,omitempty"`
	Left    json.RawMessage `json:"left,omitempty"`
	Right   json.RawMessage `json:"right,omitempty"`
}

func decodeExpr(raw json.RawMessage) (Expr, error) {
	var w exprWire
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, fmt.Errorf("expr: %w", err)
	}
	switch w.Type {
	case "number":
		if w.Value == nil {
			return nil, fmt.Errorf("expr: number missing value")
		}
		return NumberExpr{Value: *w.Value}, nil
	case "path":
		return PathExpr{Path: w.Path}, nil
	case "catalogField":
		return CatalogFieldExpr{Catalog: w.Catalog, Field: w.Field}, nil
	case "mul":
		left, err := decodeExpr(w.Left)
		if err != nil {
			return nil, fmt.Errorf("expr: mul.left: %w", err)
		}
		right, err := decodeExpr(w.Right)
		if err != nil {
			return nil, fmt.Errorf("expr: mul.right: %w", err)
		}
		return MulExpr{Left: left, Right: right}, nil
	default:
		return nil, fmt.Errorf("expr: unknown type %q", w.Type)
	}
}

func (e NumberExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(exprWire{Type: "number", Value: &e.Value})
}
func (e PathExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(exprWire{Type: "path", Path: e.Path})
}
func (e CatalogFieldExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(exprWire{Type: "catalogField", Catalog: e.Catalog, Field: e.Field})
}
func (e MulExpr) MarshalJSON() ([]byte, error) {
	left, err := json.Marshal(e.Left)
	if err != nil {
		return nil, fmt.Errorf("expr: mul.left: %w", err)
	}
	right, err := json.Marshal(e.Right)
	if err != nil {
		return nil, fmt.Errorf("expr: mul.right: %w", err)
	}
	return json.Marshal(exprWire{Type: "mul", Left: left, Right: right})
}

// ---- Fallback (no polymorphism — Kind is data, not a Go sum type) ----

func decodeFallback(raw json.RawMessage) (*Fallback, error) {
	if raw == nil {
		return nil, nil
	}
	var f Fallback
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("fallback: %w", err)
	}
	return &f, nil
}

// ---- Target ----

type targetWire struct {
	Kind     string          `json:"kind"`
	Expr     json.RawMessage `json:"expr,omitempty"`
	Plus     bool            `json:"plus,omitempty"`
	Unit     string          `json:"unit,omitempty"`
	Fallback json.RawMessage `json:"fallback,omitempty"`
}

func decodeTarget(raw json.RawMessage) (Target, error) {
	var w targetWire
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, fmt.Errorf("target: %w", err)
	}
	expr, err := decodeExpr(w.Expr)
	if err != nil {
		return nil, fmt.Errorf("target: %w", err)
	}
	fallback, err := decodeFallback(w.Fallback)
	if err != nil {
		return nil, fmt.Errorf("target: %w", err)
	}
	switch w.Kind {
	case "reps":
		return RepsTarget{Expr: expr, Plus: w.Plus, Fallback: fallback}, nil
	case "distance":
		return DistanceTarget{Expr: expr, Unit: w.Unit, Fallback: fallback}, nil
	case "duration":
		return DurationTarget{Expr: expr, Unit: w.Unit, Fallback: fallback}, nil
	default:
		return nil, fmt.Errorf("target: unknown kind %q", w.Kind)
	}
}

func marshalTarget(kind string, expr Expr, plus bool, unit string, fallback *Fallback) ([]byte, error) {
	exprJSON, err := json.Marshal(expr)
	if err != nil {
		return nil, fmt.Errorf("target: %w", err)
	}
	var fallbackJSON json.RawMessage
	if fallback != nil {
		fallbackJSON, err = json.Marshal(fallback)
		if err != nil {
			return nil, fmt.Errorf("target: %w", err)
		}
	}
	return json.Marshal(targetWire{Kind: kind, Expr: exprJSON, Plus: plus, Unit: unit, Fallback: fallbackJSON})
}

func (t RepsTarget) MarshalJSON() ([]byte, error) {
	return marshalTarget("reps", t.Expr, t.Plus, "", t.Fallback)
}
func (t DistanceTarget) MarshalJSON() ([]byte, error) {
	return marshalTarget("distance", t.Expr, false, t.Unit, t.Fallback)
}
func (t DurationTarget) MarshalJSON() ([]byte, error) {
	return marshalTarget("duration", t.Expr, false, t.Unit, t.Fallback)
}

// ---- Load ----

type loadWire struct {
	Kind     string          `json:"kind"`
	Expr     json.RawMessage `json:"expr,omitempty"`
	Unit     string          `json:"unit,omitempty"`
	Fallback json.RawMessage `json:"fallback,omitempty"`
}

func decodeLoad(raw json.RawMessage) (Load, error) {
	if raw == nil {
		return nil, nil
	}
	var w loadWire
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	expr, err := decodeExpr(w.Expr)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	fallback, err := decodeFallback(w.Fallback)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	switch w.Kind {
	case "weight":
		return WeightLoad{Expr: expr, Unit: w.Unit, Fallback: fallback}, nil
	default:
		return nil, fmt.Errorf("load: unknown kind %q", w.Kind)
	}
}

func (l WeightLoad) MarshalJSON() ([]byte, error) {
	exprJSON, err := json.Marshal(l.Expr)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	var fallbackJSON json.RawMessage
	if l.Fallback != nil {
		fallbackJSON, err = json.Marshal(l.Fallback)
		if err != nil {
			return nil, fmt.Errorf("load: %w", err)
		}
	}
	return json.Marshal(loadWire{Kind: "weight", Expr: exprJSON, Unit: l.Unit, Fallback: fallbackJSON})
}

// ---- SetRef ----

type setRefWire struct {
	Type        string           `json:"type"`
	Label       string           `json:"label,omitempty"`
	Exercise    string           `json:"exercise"`
	Target      json.RawMessage  `json:"target"`
	Load        json.RawMessage  `json:"load,omitempty"`
	Progression *ProgressionRule `json:"progression,omitempty"`
}

func (s SetRef) MarshalJSON() ([]byte, error) {
	targetJSON, err := json.Marshal(s.Target)
	if err != nil {
		return nil, fmt.Errorf("setRef %q: %w", s.Label, err)
	}
	var loadJSON json.RawMessage
	if s.Load != nil {
		loadJSON, err = json.Marshal(s.Load)
		if err != nil {
			return nil, fmt.Errorf("setRef %q: %w", s.Label, err)
		}
	}
	return json.Marshal(setRefWire{
		Type: "setRef", Label: s.Label, Exercise: s.Exercise,
		Target: targetJSON, Load: loadJSON, Progression: s.Progression,
	})
}

func (s *SetRef) UnmarshalJSON(data []byte) error {
	var w setRefWire
	if err := json.Unmarshal(data, &w); err != nil {
		return fmt.Errorf("setRef: %w", err)
	}
	target, err := decodeTarget(w.Target)
	if err != nil {
		return fmt.Errorf("setRef %q: %w", w.Label, err)
	}
	load, err := decodeLoad(w.Load)
	if err != nil {
		return fmt.Errorf("setRef %q: %w", w.Label, err)
	}
	s.Label = w.Label
	s.Exercise = w.Exercise
	s.Target = target
	s.Load = load
	s.Progression = w.Progression
	return nil
}

// ---- Ref ----

type refWire struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func (r Ref) MarshalJSON() ([]byte, error) {
	return json.Marshal(refWire{Type: "ref", Name: r.Name})
}

// ---- RestPolicy ----

type restPolicyWire struct {
	Mode     string          `json:"mode"`
	Duration json.RawMessage `json:"duration,omitempty"`
	Intra    *Duration       `json:"intra,omitempty"`
	Inter    *Duration       `json:"inter,omitempty"`
}

func decodeRestPolicy(raw json.RawMessage) (RestPolicy, error) {
	var w restPolicyWire
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, fmt.Errorf("rest: %w", err)
	}
	switch w.Mode {
	case "single":
		var d *Duration
		if w.Duration != nil {
			d = &Duration{}
			if err := json.Unmarshal(w.Duration, d); err != nil {
				return nil, fmt.Errorf("rest: single.duration: %w", err)
			}
		}
		return SingleRest{Duration: d}, nil
	case "two_level":
		if w.Intra == nil || w.Inter == nil {
			return nil, fmt.Errorf("rest: two_level requires intra and inter")
		}
		return TwoLevelRest{Intra: *w.Intra, Inter: *w.Inter}, nil
	case "ad_lib":
		return AdLibRest{}, nil
	case "remainder":
		return RemainderRest{}, nil
	default:
		return nil, fmt.Errorf("rest: unknown mode %q", w.Mode)
	}
}

func (r SingleRest) MarshalJSON() ([]byte, error) {
	var d json.RawMessage
	if r.Duration != nil {
		var err error
		d, err = json.Marshal(r.Duration)
		if err != nil {
			return nil, fmt.Errorf("rest: %w", err)
		}
	}
	return json.Marshal(restPolicyWire{Mode: "single", Duration: d})
}
func (r TwoLevelRest) MarshalJSON() ([]byte, error) {
	return json.Marshal(restPolicyWire{Mode: "two_level", Intra: &r.Intra, Inter: &r.Inter})
}
func (r AdLibRest) MarshalJSON() ([]byte, error) {
	return json.Marshal(restPolicyWire{Mode: "ad_lib"})
}
func (r RemainderRest) MarshalJSON() ([]byte, error) {
	return json.Marshal(restPolicyWire{Mode: "remainder"})
}

// ---- Termination ----

type terminationWire struct {
	Mode     string    `json:"mode"`
	N        *int      `json:"n,omitempty"`
	Interval *Duration `json:"interval,omitempty"`
	Cap      *Duration `json:"cap,omitempty"`
}

func decodeTermination(raw json.RawMessage) (Termination, error) {
	var w terminationWire
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, fmt.Errorf("termination: %w", err)
	}
	switch w.Mode {
	case "count":
		if w.N == nil {
			return nil, fmt.Errorf("termination: count requires n")
		}
		return CountTermination{N: *w.N}, nil
	case "emom":
		if w.Interval == nil || w.N == nil {
			return nil, fmt.Errorf("termination: emom requires interval and n")
		}
		return EmomTermination{Interval: *w.Interval, N: *w.N}, nil
	case "time_cap":
		if w.Cap == nil {
			return nil, fmt.Errorf("termination: time_cap requires cap")
		}
		return TimeCapTermination{Cap: *w.Cap}, nil
	case "for_time":
		return ForTimeTermination{}, nil
	default:
		return nil, fmt.Errorf("termination: unknown mode %q", w.Mode)
	}
}

func (t CountTermination) MarshalJSON() ([]byte, error) {
	return json.Marshal(terminationWire{Mode: "count", N: &t.N})
}
func (t EmomTermination) MarshalJSON() ([]byte, error) {
	return json.Marshal(terminationWire{Mode: "emom", Interval: &t.Interval, N: &t.N})
}
func (t TimeCapTermination) MarshalJSON() ([]byte, error) {
	return json.Marshal(terminationWire{Mode: "time_cap", Cap: &t.Cap})
}
func (t ForTimeTermination) MarshalJSON() ([]byte, error) {
	return json.Marshal(terminationWire{Mode: "for_time"})
}

// ---- Member / Group ----

// memberTypeProbe reads just enough of a Member's JSON to know which
// concrete type to decode into.
type memberTypeProbe struct {
	Type string `json:"type"`
}

func decodeMember(raw json.RawMessage) (Member, error) {
	var probe memberTypeProbe
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, fmt.Errorf("member: %w", err)
	}
	switch probe.Type {
	case "setRef":
		var s SetRef
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, fmt.Errorf("member: %w", err)
		}
		return s, nil
	case "ref":
		var w refWire
		if err := json.Unmarshal(raw, &w); err != nil {
			return nil, fmt.Errorf("member: %w", err)
		}
		return Ref{Name: w.Name}, nil
	case "group":
		var g Group
		if err := json.Unmarshal(raw, &g); err != nil {
			return nil, fmt.Errorf("member: %w", err)
		}
		return g, nil
	default:
		return nil, fmt.Errorf("member: unknown type %q", probe.Type)
	}
}

type groupWire struct {
	Type        string            `json:"type"`
	Kind        string            `json:"kind"`
	Members     []json.RawMessage `json:"members"`
	Interleave  string            `json:"interleave"`
	Rest        json.RawMessage   `json:"rest"`
	Termination json.RawMessage   `json:"termination"`
	Atomic      bool              `json:"atomic"`
}

func (g Group) MarshalJSON() ([]byte, error) {
	members := make([]json.RawMessage, len(g.Members))
	for i, m := range g.Members {
		raw, err := json.Marshal(m)
		if err != nil {
			return nil, fmt.Errorf("group %q: member %d: %w", g.Kind, i, err)
		}
		members[i] = raw
	}
	rest, err := json.Marshal(g.Rest)
	if err != nil {
		return nil, fmt.Errorf("group %q: rest: %w", g.Kind, err)
	}
	termination, err := json.Marshal(g.Termination)
	if err != nil {
		return nil, fmt.Errorf("group %q: termination: %w", g.Kind, err)
	}
	return json.Marshal(groupWire{
		Type: "group", Kind: g.Kind, Members: members, Interleave: g.Interleave,
		Rest: rest, Termination: termination, Atomic: g.Atomic,
	})
}

func (g *Group) UnmarshalJSON(data []byte) error {
	var w groupWire
	if err := json.Unmarshal(data, &w); err != nil {
		return fmt.Errorf("group: %w", err)
	}
	members := make([]Member, len(w.Members))
	for i, raw := range w.Members {
		m, err := decodeMember(raw)
		if err != nil {
			return fmt.Errorf("group %q: member %d: %w", w.Kind, i, err)
		}
		members[i] = m
	}
	rest, err := decodeRestPolicy(w.Rest)
	if err != nil {
		return fmt.Errorf("group %q: %w", w.Kind, err)
	}
	termination, err := decodeTermination(w.Termination)
	if err != nil {
		return fmt.Errorf("group %q: %w", w.Kind, err)
	}
	g.Kind = w.Kind
	g.Members = members
	g.Interleave = w.Interleave
	g.Rest = rest
	g.Termination = termination
	g.Atomic = w.Atomic
	return nil
}

// ---- StateBinding ----

type stateBindingWire struct {
	Path     string          `json:"path"`
	Expr     json.RawMessage `json:"expr"`
	Fallback json.RawMessage `json:"fallback,omitempty"`
}

func (b *StateBinding) UnmarshalJSON(data []byte) error {
	var w stateBindingWire
	if err := json.Unmarshal(data, &w); err != nil {
		return fmt.Errorf("stateBinding: %w", err)
	}
	expr, err := decodeExpr(w.Expr)
	if err != nil {
		return fmt.Errorf("stateBinding %q: %w", w.Path, err)
	}
	fallback, err := decodeFallback(w.Fallback)
	if err != nil {
		return fmt.Errorf("stateBinding %q: %w", w.Path, err)
	}
	b.Path = w.Path
	b.Expr = expr
	b.Fallback = fallback
	return nil
}
