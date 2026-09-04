# grammar/

A parser-generator grammar file, shared between the reference
implementation and any tooling that wants its own parser — relevant when
the reference implementation uses a parser generator (ANTLR, pigeon,
participle, …).

**Deliberately empty.** The reference implementation ([`../reference/`](../reference/))
is Go with a hand-written recursive-descent parser, not a parser
generator — [`../spec/grammar.ebnf`](../spec/grammar.ebnf) remains the
authoritative, implementation-independent grammar, and there's no
generated artifact to place here. Revisit only if a future implementation
in another language wants to share a generated parser.
