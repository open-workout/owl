// Package lexer tokenizes OWL source per spec/grammar.ebnf's lexical
// grammar.
package lexer

import "fmt"

type Kind int

const (
	EOF Kind = iota
	IDENT
	NUMBER
	STRING
	CATALOG_NAME // '$' IDENT, no whitespace between — scanned as one token

	// Keywords. Lexed as distinct kinds (not IDENT) since grammar.ebnf
	// enumerates them as a small closed set.
	STATE // "state" or "stats" (accepted synonyms)
	UNITS
	PLATES
	BLOCK
	DAY
	EXERCISE
	SET
	PROGRESS
	DROPSET
	DROP
	REST
	IF
	THEN
	ELSE
	SUPERSET
	CIRCUIT
	EMOM
	AMRAP
	FOR_TIME
	ROUNDS
	AS

	// Punctuation / operators.
	ASSIGN // '='
	LBRACE
	RBRACE
	LPAREN
	RPAREN
	LBRACKET
	RBRACKET
	COMMA
	SEMI
	DOT
	STAR
	AT
	PLUS
	LT
	GT
	LE
	GE
	EQ
	NE
	PIPE  // '|', introduces a state binding's fallback
	COLON // ':', the optional cosmetic marker after 'then'
)

var keywords = map[string]Kind{
	"state":    STATE,
	"stats":    STATE,
	"units":    UNITS,
	"plates":   PLATES,
	"block":    BLOCK,
	"day":      DAY,
	"exercise": EXERCISE,
	"set":      SET,
	"progress": PROGRESS,
	"dropset":  DROPSET,
	"drop":     DROP,
	"rest":     REST,
	"if":       IF,
	"then":     THEN,
	"else":     ELSE,
	"superset": SUPERSET,
	"circuit":  CIRCUIT,
	"emom":     EMOM,
	"amrap":    AMRAP,
	"for_time": FOR_TIME,
	"rounds":   ROUNDS,
	"as":       AS,
}

var kindNames = map[Kind]string{
	EOF: "EOF", IDENT: "IDENT", NUMBER: "NUMBER", STRING: "STRING",
	CATALOG_NAME: "CATALOG_NAME",
	ASSIGN:       "'='", LBRACE: "'{'", RBRACE: "'}'", LPAREN: "'('", RPAREN: "')'",
	LBRACKET: "'['", RBRACKET: "']'", COMMA: "','", SEMI: "';'", DOT: "'.'",
	STAR: "'*'", AT: "'@'", PLUS: "'+'",
	LT: "'<'", GT: "'>'", LE: "'<='", GE: "'>='", EQ: "'=='", NE: "'!='",
	PIPE: "'|'", COLON: "':'",
}

func (k Kind) String() string {
	if s, ok := kindNames[k]; ok {
		return s
	}
	for lit, kw := range keywords {
		if kw == k {
			return fmt.Sprintf("%q", lit)
		}
	}
	return "?"
}

// Pos is a 1-based source position, for error messages.
type Pos struct {
	Line, Col int
}

func (p Pos) String() string { return fmt.Sprintf("%d:%d", p.Line, p.Col) }

type Token struct {
	Kind Kind
	Lit  string // the literal text; for STRING, unquoted
	Pos  Pos
}
