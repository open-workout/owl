package owl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// TestCompileParseFixtures runs Compile against every conformance/parse
// fixture and compares the result against expected.json/expected-error.json
// as Go values (not raw JSON text — see the plan's note on why: fixtures
// aren't byte-consistent with what encoding/json+omitempty produces).
func TestCompileParseFixtures(t *testing.T) {
	root := "../conformance/parse"
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		t.Run(e.Name(), func(t *testing.T) {
			src, err := os.ReadFile(filepath.Join(dir, "source.owl"))
			if err != nil {
				t.Fatal(err)
			}
			got, err := Compile(string(src))

			errPath := filepath.Join(dir, "expected-error.json")
			if _, statErr := os.Stat(errPath); statErr == nil {
				if err == nil {
					t.Fatalf("expected a compile error, got none (result: %+v)", got)
				}
				return // message text isn't asserted verbatim; parse-error.schema.json doesn't require exact wording
			}

			if err != nil {
				t.Fatalf("Compile: %v", err)
			}
			wantJSON, err := os.ReadFile(filepath.Join(dir, "expected.json"))
			if err != nil {
				t.Fatal(err)
			}
			want, err := ParseCanonicalJSON(wantJSON)
			if err != nil {
				t.Fatalf("bad fixture: %v", err)
			}
			if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("canonical form mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestCompileProgramsSmoke is a structural smoke test over
// conformance/programs/*.owl — these have no golden expected.json (too
// error-prone to hand-compute at this size), so this only asserts they
// compile without error, catching regressions/panics on realistic
// programs. multiple-blocks.owl is included now that grammar.ebnf
// supports its anonymous-repeat syntax (spec/semantics/groups.md §3.7a).
func TestCompileProgramsSmoke(t *testing.T) {
	root := "../conformance/programs"
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".owl" {
			continue
		}
		e := e
		t.Run(e.Name(), func(t *testing.T) {
			src, err := os.ReadFile(filepath.Join(root, e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := Compile(string(src)); err != nil {
				t.Fatalf("Compile: %v", err)
			}
		})
	}
}
