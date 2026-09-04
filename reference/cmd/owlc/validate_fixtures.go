package main

// validateFixtures walks conformance/{parse,resolve,progress}/ and checks
// every fixture file against the JSON Schema it's supposed to match. Port
// of the former conformance/schema/validate.mjs (Node/ajv) — see
// ../../README.md for why this moved to Go.
//
// This only checks structural shape — it cannot check that a fixture's
// expected.json is the *correct* output for its input, since that
// requires Compile/Resolve/Progress to actually be implemented (see
// owl.go). It exists so a malformed fixture fails fast in CI.

import (
	"fmt"
	"os"
	"path/filepath"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

type schemaSet struct {
	canonicalForm *jsonschema.Schema
	state         *jsonschema.Schema
	cursor        *jsonschema.Schema
	sessionLog    *jsonschema.Schema
	parseError    *jsonschema.Schema
	session       *jsonschema.Schema
}

func compileSchemas(root string) (*schemaSet, error) {
	c := jsonschema.NewCompiler()
	compile := func(rel string) (*jsonschema.Schema, error) {
		sch, err := c.Compile(filepath.Join(root, rel))
		if err != nil {
			return nil, fmt.Errorf("compiling %s: %w", rel, err)
		}
		return sch, nil
	}

	var s schemaSet
	var err error
	for _, step := range []struct {
		rel string
		dst **jsonschema.Schema
	}{
		{"spec/canonical-form.schema.json", &s.canonicalForm},
		{"conformance/schema/state.schema.json", &s.state},
		{"conformance/schema/cursor.schema.json", &s.cursor},
		{"conformance/schema/session-log.schema.json", &s.sessionLog},
		{"conformance/schema/parse-error.schema.json", &s.parseError},
		{"conformance/schema/session.schema.json", &s.session},
	} {
		*step.dst, err = compile(step.rel)
		if err != nil {
			return nil, err
		}
	}
	return &s, nil
}

func loadDoc(path string) (any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return jsonschema.UnmarshalJSON(f)
}

func validateFixtures(root string) error {
	schemas, err := compileSchemas(root)
	if err != nil {
		return err
	}

	failures := 0
	checked := 0

	check := func(label string, sch *jsonschema.Schema, path string) {
		checked++
		doc, err := loadDoc(path)
		if err != nil {
			failures++
			fmt.Printf("FAIL [%s] %s: %v\n", label, rel(root, path), err)
			return
		}
		if err := sch.Validate(doc); err != nil {
			failures++
			fmt.Printf("FAIL [%s] %s:\n%v\n", label, rel(root, path), err)
		}
	}

	// parse/<case>/: source.owl + exactly one of expected.json | expected-error.json
	for _, dir := range subdirs(filepath.Join(root, "conformance/parse")) {
		expected := filepath.Join(dir, "expected.json")
		expectedError := filepath.Join(dir, "expected-error.json")
		hasExpected := fileExists(expected)
		hasError := fileExists(expectedError)
		if hasExpected == hasError {
			failures++
			fmt.Printf("FAIL [parse] %s: expected exactly one of expected.json / expected-error.json\n", rel(root, dir))
			continue
		}
		if hasExpected {
			check("parse", schemas.canonicalForm, expected)
		} else {
			check("parse", schemas.parseError, expectedError)
		}
	}

	// resolve/<case>/: canonical.json, state.json, cursor.json, expected.json (session)
	for _, dir := range subdirs(filepath.Join(root, "conformance/resolve")) {
		check("resolve", schemas.canonicalForm, filepath.Join(dir, "canonical.json"))
		check("resolve", schemas.state, filepath.Join(dir, "state.json"))
		check("resolve", schemas.cursor, filepath.Join(dir, "cursor.json"))
		check("resolve", schemas.session, filepath.Join(dir, "expected.json"))
	}

	// progress/<case>/: canonical.json, state.json (before), log.json, expected.json (state after)
	for _, dir := range subdirs(filepath.Join(root, "conformance/progress")) {
		check("progress", schemas.canonicalForm, filepath.Join(dir, "canonical.json"))
		check("progress", schemas.state, filepath.Join(dir, "state.json"))
		check("progress", schemas.sessionLog, filepath.Join(dir, "log.json"))
		check("progress", schemas.state, filepath.Join(dir, "expected.json"))
	}

	fmt.Printf("\n%d file(s) checked, %d failure(s).\n", checked, failures)
	if failures > 0 {
		return fmt.Errorf("%d fixture(s) failed validation", failures)
	}
	return nil
}

func subdirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(dir, e.Name()))
		}
	}
	return dirs
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func rel(root, path string) string {
	r, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return r
}
