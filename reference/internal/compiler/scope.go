package compiler

import "github.com/open-workout/owl/reference/internal/ir"

// scope tracks names resolvable via a bare Ref or dottedPath at one
// nesting level, chained to its parent for enclosing-scope lookups — a
// child scope only ever adds names, never shadows/removes an ancestor's.
//
// SetRefs are tracked by pointer, not value, so that a progression rule
// discovered later in the same exercise (progress lines always follow
// their target set(s) in every attested example) can be attached by
// mutating the pointee in place — every place that SetRef is later
// embedded (ExerciseDecl.Sets, a dropset's Group.Members, the
// exercise's own Body.Members) is built by dereferencing these same
// pointers *after* attachment, so the embedded copies stay consistent.
type scope struct {
	parent *scope

	names    map[string]bool          // declared exercise/day/block/named-group names, for Ref validation
	sets     map[string]*ir.SetRef    // in-scope `set` labels
	dropsets map[string]*dropsetBuild // in-scope `dropset` names

	// bodies maps a declared exercise/day/block/named-group's name to
	// its own compiled Group, for `name*N` repetition (groups.md §3.7)
	// and superset/circuit round-count inference (§3.2) — both need the
	// actual compiled content, not just a validity check.
	bodies map[string]*ir.Group
}

func newScope(parent *scope) *scope {
	return &scope{
		parent:   parent,
		names:    map[string]bool{},
		sets:     map[string]*ir.SetRef{},
		dropsets: map[string]*dropsetBuild{},
		bodies:   map[string]*ir.Group{},
	}
}

func (s *scope) declareName(name string) { s.names[name] = true }

func (s *scope) hasName(name string) bool {
	for cur := s; cur != nil; cur = cur.parent {
		if cur.names[name] {
			return true
		}
	}
	return false
}

func (s *scope) declareSet(label string, ref *ir.SetRef) {
	if label != "" {
		s.sets[label] = ref
	}
}

func (s *scope) lookupSet(label string) *ir.SetRef {
	for cur := s; cur != nil; cur = cur.parent {
		if r, ok := cur.sets[label]; ok {
			return r
		}
	}
	return nil
}

func (s *scope) declareBody(name string, g *ir.Group) { s.bodies[name] = g }

func (s *scope) lookupBody(name string) *ir.Group {
	for cur := s; cur != nil; cur = cur.parent {
		if g, ok := cur.bodies[name]; ok {
			return g
		}
	}
	return nil
}

func (s *scope) declareDropset(name string, d *dropsetBuild) { s.dropsets[name] = d }

func (s *scope) lookupDropset(name string) *dropsetBuild {
	for cur := s; cur != nil; cur = cur.parent {
		if d, ok := cur.dropsets[name]; ok {
			return d
		}
	}
	return nil
}
