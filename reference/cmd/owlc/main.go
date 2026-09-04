// Command owlc is the OWL reference CLI. See ../../README.md.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	owl "github.com/open-workout/owl/reference"
)

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(2)
	}
	cmd, path := os.Args[1], os.Args[2]

	if cmd == "validate-fixtures" {
		if err := validateFixtures(path); err != nil {
			fatal(err)
		}
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fatal(err)
	}

	switch cmd {
	case "parse":
		program, err := owl.Compile(string(data))
		if err != nil {
			fatal(err)
		}
		printJSON(program)

	case "validate":
		// Decodes a canonical-form JSON document (e.g. a
		// conformance/parse/*/expected.json fixture) into the Go
		// types and re-encodes it — exercises the wire types without
		// needing a working parser yet.
		program, err := owl.ParseCanonicalJSON(data)
		if err != nil {
			fatal(err)
		}
		printJSON(program)

	default:
		usage()
		os.Exit(2)
	}
}

func printJSON(v any) {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fatal(err)
	}
	fmt.Println(string(out))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "owlc:", err)
	os.Exit(1)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  owlc parse <source.owl>              parse OWL source to canonical form (not implemented yet)")
	fmt.Fprintln(os.Stderr, "  owlc validate <canonical.json>       round-trip a canonical-form JSON document through the Go types")
	fmt.Fprintln(os.Stderr, "  owlc validate-fixtures <repo-root>   validate every conformance/{parse,resolve,progress}/ fixture against its JSON Schema")
}
