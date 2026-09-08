package lexer

import (
	"fmt"
	"strings"
)

// Error is a lexical error at a specific source position.
type Error struct {
	Msg string
	Pos Pos
}

func (e *Error) Error() string { return fmt.Sprintf("%s: %s", e.Pos, e.Msg) }

// Lex tokenizes source in full, returning every token including a
// trailing EOF, or the first lexical error encountered.
func Lex(source string) ([]Token, error) {
	l := &lexer{src: []rune(source), line: 1, col: 1}
	var toks []Token
	for {
		tok, err := l.next()
		if err != nil {
			return nil, err
		}
		toks = append(toks, tok)
		if tok.Kind == EOF {
			return toks, nil
		}
	}
}

type lexer struct {
	src       []rune
	pos       int // index into src
	line, col int // position of src[pos]
}

func (l *lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *lexer) peekAt(offset int) rune {
	if l.pos+offset >= len(l.src) {
		return 0
	}
	return l.src[l.pos+offset]
}

func (l *lexer) advance() rune {
	r := l.src[l.pos]
	l.pos++
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}

func (l *lexer) atEnd() bool { return l.pos >= len(l.src) }

func isLetter(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}
func isDigit(r rune) bool { return r >= '0' && r <= '9' }

func (l *lexer) skipInsignificant() {
	for !l.atEnd() {
		r := l.peek()
		switch {
		case r == ' ' || r == '\t' || r == '\r' || r == '\n':
			l.advance()
		case r == '#':
			for !l.atEnd() && l.peek() != '\n' {
				l.advance()
			}
		default:
			return
		}
	}
}

func (l *lexer) next() (Token, error) {
	l.skipInsignificant()
	start := Pos{Line: l.line, Col: l.col}
	if l.atEnd() {
		return Token{Kind: EOF, Pos: start}, nil
	}

	r := l.peek()
	switch {
	case isLetter(r):
		return l.lexIdentOrKeyword(start), nil
	case isDigit(r):
		return l.lexNumber(start)
	case r == '"':
		return l.lexString(start)
	case r == '$':
		return l.lexCatalogName(start)
	default:
		return l.lexPunct(start)
	}
}

func (l *lexer) lexIdentOrKeyword(start Pos) Token {
	var sb strings.Builder
	for !l.atEnd() && (isLetter(l.peek()) || isDigit(l.peek())) {
		sb.WriteRune(l.advance())
	}
	lit := sb.String()
	if kw, ok := keywords[lit]; ok {
		return Token{Kind: kw, Lit: lit, Pos: start}
	}
	return Token{Kind: IDENT, Lit: lit, Pos: start}
}

func (l *lexer) lexNumber(start Pos) (Token, error) {
	var sb strings.Builder
	for !l.atEnd() && isDigit(l.peek()) {
		sb.WriteRune(l.advance())
	}
	if l.peek() == '.' && isDigit(l.peekAt(1)) {
		sb.WriteRune(l.advance()) // '.'
		for !l.atEnd() && isDigit(l.peek()) {
			sb.WriteRune(l.advance())
		}
	}
	return Token{Kind: NUMBER, Lit: sb.String(), Pos: start}, nil
}

func (l *lexer) lexString(start Pos) (Token, error) {
	l.advance() // opening '"'
	var sb strings.Builder
	for {
		if l.atEnd() {
			return Token{}, &Error{Msg: "unterminated string literal", Pos: start}
		}
		if l.peek() == '"' {
			l.advance()
			return Token{Kind: STRING, Lit: sb.String(), Pos: start}, nil
		}
		sb.WriteRune(l.advance())
	}
}

func (l *lexer) lexCatalogName(start Pos) (Token, error) {
	l.advance() // '$'
	if l.atEnd() || !isLetter(l.peek()) {
		return Token{}, &Error{Msg: "'$' must be immediately followed by an identifier", Pos: start}
	}
	var sb strings.Builder
	for !l.atEnd() && (isLetter(l.peek()) || isDigit(l.peek())) {
		sb.WriteRune(l.advance())
	}
	return Token{Kind: CATALOG_NAME, Lit: sb.String(), Pos: start}, nil
}

func (l *lexer) lexPunct(start Pos) (Token, error) {
	r := l.advance()
	two := func(second rune, twoKind, oneKind Kind) Token {
		if l.peek() == second {
			l.advance()
			return Token{Kind: twoKind, Lit: string(r) + string(second), Pos: start}
		}
		return Token{Kind: oneKind, Lit: string(r), Pos: start}
	}
	switch r {
	case '=':
		return two('=', EQ, ASSIGN), nil
	case '{':
		return Token{Kind: LBRACE, Lit: "{", Pos: start}, nil
	case '}':
		return Token{Kind: RBRACE, Lit: "}", Pos: start}, nil
	case '(':
		return Token{Kind: LPAREN, Lit: "(", Pos: start}, nil
	case ')':
		return Token{Kind: RPAREN, Lit: ")", Pos: start}, nil
	case '[':
		return Token{Kind: LBRACKET, Lit: "[", Pos: start}, nil
	case ']':
		return Token{Kind: RBRACKET, Lit: "]", Pos: start}, nil
	case ',':
		return Token{Kind: COMMA, Lit: ",", Pos: start}, nil
	case ';':
		return Token{Kind: SEMI, Lit: ";", Pos: start}, nil
	case '.':
		return Token{Kind: DOT, Lit: ".", Pos: start}, nil
	case '*':
		return Token{Kind: STAR, Lit: "*", Pos: start}, nil
	case '@':
		return Token{Kind: AT, Lit: "@", Pos: start}, nil
	case '+':
		return Token{Kind: PLUS, Lit: "+", Pos: start}, nil
	case '|':
		return Token{Kind: PIPE, Lit: "|", Pos: start}, nil
	case ':':
		return Token{Kind: COLON, Lit: ":", Pos: start}, nil
	case '<':
		return two('=', LE, LT), nil
	case '>':
		return two('=', GE, GT), nil
	case '!':
		if l.peek() == '=' {
			l.advance()
			return Token{Kind: NE, Lit: "!=", Pos: start}, nil
		}
		return Token{}, &Error{Msg: "unexpected character '!'", Pos: start}
	default:
		return Token{}, &Error{Msg: fmt.Sprintf("unexpected character %q", r), Pos: start}
	}
}
