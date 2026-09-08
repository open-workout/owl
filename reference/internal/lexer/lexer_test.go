package lexer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLexBasics(t *testing.T) {
	src := `state = { tm.a = 0.85 * $Foo.e1rm | AMW }
units = "kg"
if tm.a <= 5 then: x; else y; # trailing comment
!= >= == < >`
	toks, err := Lex(src)
	if err != nil {
		t.Fatal(err)
	}
	if toks[len(toks)-1].Kind != EOF {
		t.Fatalf("last token should be EOF, got %v", toks[len(toks)-1].Kind)
	}
	// Spot check a few tokens rather than the whole stream.
	assertKindAt(t, toks, 0, STATE)
	assertKindAt(t, toks, 3, IDENT)        // tm
	assertKindAt(t, toks, 9, CATALOG_NAME) // $Foo
	assertKindAt(t, toks, 11, IDENT)       // e1rm, part of catalogFieldRef (CATALOG_NAME '.' IDENT)
	assertKindAt(t, toks, 12, PIPE)
	assertKindAt(t, toks, 13, IDENT) // AMW (sentinel)
}

func assertKindAt(t *testing.T, toks []Token, i int, want Kind) {
	t.Helper()
	if i >= len(toks) {
		t.Fatalf("token %d out of range (len %d)", i, len(toks))
	}
	if toks[i].Kind != want {
		t.Fatalf("token %d: got %v, want %v", i, toks[i].Kind, want)
	}
}

func TestCatalogNameRequiresNoGap(t *testing.T) {
	if _, err := Lex("$ Foo"); err == nil {
		t.Fatal("expected an error for '$' followed by whitespace")
	}
	toks, err := Lex("$Foo")
	if err != nil {
		t.Fatal(err)
	}
	if toks[0].Kind != CATALOG_NAME || toks[0].Lit != "Foo" {
		t.Fatalf("got %+v", toks[0])
	}
}

func TestNumbers(t *testing.T) {
	toks, err := Lex("5 0.85 100")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"5", "0.85", "100"}
	for i, w := range want {
		if toks[i].Kind != NUMBER || toks[i].Lit != w {
			t.Fatalf("token %d: got %+v, want NUMBER %q", i, toks[i], w)
		}
	}
}

func TestSemicolonsAlwaysOptional(t *testing.T) {
	a, err := Lex("x;y")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Lex("x\ny")
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 4 { // IDENT SEMI IDENT EOF
		t.Fatalf("got %d tokens", len(a))
	}
	if len(b) != 3 { // IDENT IDENT EOF
		t.Fatalf("got %d tokens", len(b))
	}
}

// TestLexAllFixtures is a smoke test: every real .owl file in the
// conformance corpus must lex without error, even before the parser
// exists.
func TestLexAllFixtures(t *testing.T) {
	roots := []string{"../../../conformance/parse", "../../../conformance/programs"}
	for _, root := range roots {
		var files []string
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && filepath.Ext(path) == ".owl" {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range files {
			f := f
			t.Run(f, func(t *testing.T) {
				data, err := os.ReadFile(f)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := Lex(string(data)); err != nil {
					t.Fatalf("lex error: %v", err)
				}
			})
		}
	}
}
