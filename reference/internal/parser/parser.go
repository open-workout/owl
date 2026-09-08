package parser

import (
	"fmt"

	"github.com/open-workout/owl/reference/internal/lexer"
)

// SyntaxError is a parse error at a specific source position.
type SyntaxError struct {
	Msg string
	Pos lexer.Pos
}

func (e *SyntaxError) Error() string { return fmt.Sprintf("%s: %s", e.Pos, e.Msg) }

// Parse tokenizes is assumed done already (tokens from lexer.Lex) and
// parses a full Program.
func Parse(toks []lexer.Token) (*Program, error) {
	p := &parser{toks: toks}
	prog, err := p.parseProgram()
	if err != nil {
		return nil, err
	}
	return prog, nil
}

type parser struct {
	toks []lexer.Token
	pos  int
}

func (p *parser) cur() lexer.Token     { return p.toks[p.pos] }
func (p *parser) kind() lexer.Kind     { return p.toks[p.pos].Kind }
func (p *parser) at(k lexer.Kind) bool { return p.kind() == k }

// peekAt looks ahead `offset` tokens without consuming.
func (p *parser) peekAt(offset int) lexer.Token {
	i := p.pos + offset
	if i >= len(p.toks) {
		return p.toks[len(p.toks)-1] // EOF
	}
	return p.toks[i]
}

func (p *parser) advance() lexer.Token {
	t := p.toks[p.pos]
	if p.pos < len(p.toks)-1 {
		p.pos++
	}
	return t
}

func (p *parser) errf(format string, args ...any) error {
	return &SyntaxError{Msg: fmt.Sprintf(format, args...), Pos: p.cur().Pos}
}

func (p *parser) expect(k lexer.Kind) (lexer.Token, error) {
	if !p.at(k) {
		return lexer.Token{}, p.errf("expected %v, got %v %q", k, p.kind(), p.cur().Lit)
	}
	return p.advance(), nil
}

// skipOptSemi consumes an optional trailing ';' — ';' is never required
// anywhere per grammar.ebnf.
func (p *parser) skipOptSemi() {
	if p.at(lexer.SEMI) {
		p.advance()
	}
}

// skipOptColon consumes the optional cosmetic ':' after 'then'.
func (p *parser) skipOptColon() {
	if p.at(lexer.COLON) {
		p.advance()
	}
}

// ---- Program ----

func (p *parser) parseProgram() (*Program, error) {
	prog := &Program{}
	for !p.at(lexer.EOF) {
		switch p.kind() {
		case lexer.STATE:
			bindings, err := p.parseStateSection()
			if err != nil {
				return nil, err
			}
			prog.State = append(prog.State, bindings...)
		case lexer.UNITS:
			units, err := p.parseUnitsDecl()
			if err != nil {
				return nil, err
			}
			prog.Units = units
		case lexer.PLATES:
			plates, err := p.parsePlatesDecl()
			if err != nil {
				return nil, err
			}
			prog.Plates = plates
		case lexer.BLOCK:
			decl, err := p.parseBlockDecl()
			if err != nil {
				return nil, err
			}
			prog.Items = append(prog.Items, decl)
		case lexer.IF:
			item, err := p.parseCondItem(func(p *parser) (any, error) { return p.parseTopLevelItemInner() })
			if err != nil {
				return nil, err
			}
			prog.Items = append(prog.Items, item)
		default:
			s, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			prog.Items = append(prog.Items, s)
		}
		p.skipOptSemi()
	}
	return prog, nil
}

// parseTopLevelItemInner parses one topLevelItem for use inside a
// conditional's then/else branch (grammar recurses condTopLevelItem's
// branches through the full topLevelItem alternation, including nested
// conditionals and block/state/units/plates decls — but in practice
// only stmt/blockDecl/condTopLevelItem are attested as branches).
func (p *parser) parseTopLevelItemInner() (any, error) {
	switch p.kind() {
	case lexer.BLOCK:
		return p.parseBlockDecl()
	case lexer.IF:
		return p.parseCondItem(func(p *parser) (any, error) { return p.parseTopLevelItemInner() })
	default:
		return p.parseStmt()
	}
}

func (p *parser) parseStateSection() ([]StateBinding, error) {
	p.advance() // 'state'/'stats'
	if _, err := p.expect(lexer.ASSIGN); err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.LBRACE); err != nil {
		return nil, err
	}
	var bindings []StateBinding
	for !p.at(lexer.RBRACE) {
		b, err := p.parseStateBinding()
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, b)
		p.skipOptSemi()
	}
	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}
	return bindings, nil
}

func (p *parser) parseStateBinding() (StateBinding, error) {
	path, err := p.parseDottedPath()
	if err != nil {
		return StateBinding{}, err
	}
	if _, err := p.expect(lexer.ASSIGN); err != nil {
		return StateBinding{}, err
	}
	expr, err := p.parseExpr()
	if err != nil {
		return StateBinding{}, err
	}
	var fallback *Fallback
	if p.at(lexer.PIPE) {
		p.advance()
		fallback, err = p.parseFallback()
		if err != nil {
			return StateBinding{}, err
		}
	}
	return StateBinding{Path: *path, Expr: expr, Fallback: fallback}, nil
}

func (p *parser) parseFallback() (*Fallback, error) {
	switch p.kind() {
	case lexer.NUMBER:
		v, err := p.parseNumberLiteral()
		if err != nil {
			return nil, err
		}
		plus := false
		if p.at(lexer.PLUS) {
			p.advance()
			plus = true
		}
		return &Fallback{Numeric: &NumericFallback{Value: v, Plus: plus}}, nil
	case lexer.IDENT:
		lit := p.advance().Lit
		return &Fallback{Sentinel: lit}, nil
	default:
		return nil, p.errf("expected a numeric fallback or a sentinel name, got %v", p.kind())
	}
}

func (p *parser) parseUnitsDecl() (string, error) {
	p.advance() // 'units'
	if _, err := p.expect(lexer.ASSIGN); err != nil {
		return "", err
	}
	tok, err := p.expect(lexer.STRING)
	if err != nil {
		return "", err
	}
	return tok.Lit, nil
}

func (p *parser) parsePlatesDecl() ([]float64, error) {
	p.advance() // 'plates'
	if _, err := p.expect(lexer.ASSIGN); err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.LBRACKET); err != nil {
		return nil, err
	}
	var plates []float64
	for {
		v, err := p.parseNumberLiteral()
		if err != nil {
			return nil, err
		}
		plates = append(plates, v)
		if p.at(lexer.COMMA) {
			p.advance()
			continue
		}
		break
	}
	if _, err := p.expect(lexer.RBRACKET); err != nil {
		return nil, err
	}
	return plates, nil
}

// ---- block / day / exercise declarations ----

func (p *parser) parseBlockDecl() (*BlockDecl, error) {
	pos := p.cur().Pos
	p.advance() // 'block'
	name, err := p.expect(lexer.IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.ASSIGN); err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.LBRACE); err != nil {
		return nil, err
	}
	var items []any
	for !p.at(lexer.RBRACE) {
		item, err := p.parseBlockItem()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		p.skipOptSemi()
	}
	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}
	return &BlockDecl{Name: name.Lit, Items: items, Pos: pos}, nil
}

func (p *parser) parseBlockItem() (any, error) {
	switch p.kind() {
	case lexer.DAY:
		return p.parseDayDecl()
	case lexer.IF:
		return p.parseCondItem(func(p *parser) (any, error) { return p.parseBlockItem() })
	default:
		return p.parseStmt()
	}
}

func (p *parser) parseDayDecl() (*DayDecl, error) {
	pos := p.cur().Pos
	p.advance() // 'day'
	name, err := p.expect(lexer.IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.ASSIGN); err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.LBRACE); err != nil {
		return nil, err
	}
	var items []any
	for !p.at(lexer.RBRACE) {
		item, err := p.parseDayItem()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		p.skipOptSemi()
	}
	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}
	return &DayDecl{Name: name.Lit, Items: items, Pos: pos}, nil
}

func (p *parser) parseDayItem() (any, error) {
	switch p.kind() {
	case lexer.EXERCISE:
		return p.parseExerciseDecl()
	case lexer.IF:
		return p.parseCondItem(func(p *parser) (any, error) { return p.parseDayItem() })
	default:
		return p.parseStmt()
	}
}

func (p *parser) parseExerciseDecl() (*ExerciseDecl, error) {
	pos := p.cur().Pos
	p.advance() // 'exercise'
	name, err := p.expect(lexer.IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.ASSIGN); err != nil {
		return nil, err
	}
	catalog, err := p.expect(lexer.CATALOG_NAME)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.LBRACE); err != nil {
		return nil, err
	}
	var items []any
	for !p.at(lexer.RBRACE) {
		item, err := p.parseExerciseItem()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		p.skipOptSemi()
	}
	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}
	return &ExerciseDecl{Name: name.Lit, Catalog: catalog.Lit, Items: items, Pos: pos}, nil
}

func (p *parser) parseExerciseItem() (any, error) {
	switch p.kind() {
	case lexer.SET:
		return p.parseSetDecl()
	case lexer.PROGRESS:
		return p.parseProgressDecl()
	case lexer.DROPSET:
		return p.parseDropsetDecl()
	case lexer.IF:
		return p.parseCondItem(func(p *parser) (any, error) { return p.parseExerciseItem() })
	default:
		return nil, p.errf("expected 'set', 'progress', 'dropset', or 'if', got %v", p.kind())
	}
}

func (p *parser) parseSetDecl() (*SetDecl, error) {
	pos := p.cur().Pos
	p.advance() // 'set'
	labelTok, err := p.expect(lexer.IDENT)
	if err != nil {
		return nil, err
	}
	label := labelTok.Lit
	if label == "_" {
		label = ""
	}
	if _, err := p.expect(lexer.ASSIGN); err != nil {
		return nil, err
	}
	target, err := p.parseQuantity()
	if err != nil {
		return nil, err
	}
	var load *Quantity
	if p.at(lexer.AT) {
		p.advance()
		q, err := p.parseQuantity()
		if err != nil {
			return nil, err
		}
		load = &q
	}
	var drop *DropModifier
	if p.at(lexer.DROP) {
		drop, err = p.parseDropModifier()
		if err != nil {
			return nil, err
		}
	}
	return &SetDecl{Label: label, Target: target, Load: load, Drop: drop, Pos: pos}, nil
}

func (p *parser) parseDropModifier() (*DropModifier, error) {
	p.advance() // 'drop'
	if _, err := p.expect(lexer.LPAREN); err != nil {
		return nil, err
	}
	var factors []float64
	for {
		v, err := p.parseNumberLiteral()
		if err != nil {
			return nil, err
		}
		factors = append(factors, v)
		if p.at(lexer.COMMA) {
			p.advance()
			continue
		}
		break
	}
	if _, err := p.expect(lexer.RPAREN); err != nil {
		return nil, err
	}
	return &DropModifier{Factors: factors}, nil
}

func (p *parser) parseProgressDecl() (*ProgressDecl, error) {
	pos := p.cur().Pos
	p.advance() // 'progress'
	if _, err := p.expect(lexer.ASSIGN); err != nil {
		return nil, err
	}
	scheme, err := p.expect(lexer.IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.LPAREN); err != nil {
		return nil, err
	}
	var args []Expr
	if !p.at(lexer.RPAREN) {
		for {
			e, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, e)
			if p.at(lexer.COMMA) {
				p.advance()
				continue
			}
			break
		}
	}
	if _, err := p.expect(lexer.RPAREN); err != nil {
		return nil, err
	}
	return &ProgressDecl{Scheme: scheme.Lit, Args: args, Pos: pos}, nil
}

func (p *parser) parseDropsetDecl() (*DropsetDecl, error) {
	pos := p.cur().Pos
	p.advance() // 'dropset'
	name, err := p.expect(lexer.IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.ASSIGN); err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.LBRACE); err != nil {
		return nil, err
	}
	var sets []*SetDecl
	for !p.at(lexer.RBRACE) {
		s, err := p.parseSetDecl()
		if err != nil {
			return nil, err
		}
		sets = append(sets, s)
		p.skipOptSemi()
	}
	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}
	return &DropsetDecl{Name: name.Lit, Sets: sets, Pos: pos}, nil
}

// ---- conditionals ----

func (p *parser) parseCondItem(parseItem func(*parser) (any, error)) (*CondItem, error) {
	pos := p.cur().Pos
	p.advance() // 'if'
	cond, err := p.parseCond()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.THEN); err != nil {
		return nil, err
	}
	p.skipOptColon()
	then, err := parseItem(p)
	if err != nil {
		return nil, err
	}
	p.skipOptSemi()
	if _, err := p.expect(lexer.ELSE); err != nil {
		return nil, err
	}
	els, err := parseItem(p)
	if err != nil {
		return nil, err
	}
	return &CondItem{Cond: cond, Then: then, Else: els, Pos: pos}, nil
}

var relOps = map[lexer.Kind]string{
	lexer.LT: "<", lexer.GT: ">", lexer.LE: "<=", lexer.GE: ">=", lexer.EQ: "==", lexer.NE: "!=",
}

func (p *parser) parseCond() (Cond, error) {
	left, err := p.parseExpr()
	if err != nil {
		return Cond{}, err
	}
	op, ok := relOps[p.kind()]
	if !ok {
		return Cond{}, p.errf("expected a relational operator, got %v", p.kind())
	}
	p.advance()
	right, err := p.parseExpr()
	if err != nil {
		return Cond{}, err
	}
	return Cond{Op: op, Left: left, Right: right}, nil
}

// ---- statements ----

func (p *parser) parseStmt() (any, error) {
	switch p.kind() {
	case lexer.REST:
		return p.parseRestStmt()
	case lexer.CATALOG_NAME:
		return p.parseInlineCall()
	case lexer.SUPERSET, lexer.CIRCUIT:
		return p.parseSupersetExpr()
	case lexer.EMOM:
		return p.parseCadenceExpr(1)
	case lexer.AMRAP:
		return p.parseAmrapExpr()
	case lexer.FOR_TIME:
		return p.parseForTimeExpr()
	case lexer.NUMBER:
		// Only valid stmt-start shape: `NUMBER '*' 'emom' ...`.
		n, err := p.parseNumberLiteral()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.STAR); err != nil {
			return nil, err
		}
		if !p.at(lexer.EMOM) {
			return nil, p.errf("expected 'emom' after 'N*', got %v", p.kind())
		}
		return p.parseCadenceExpr(int(n))
	case lexer.LPAREN:
		return p.parseRepeatStmtParen()
	case lexer.IDENT:
		if p.peekAt(1).Kind == lexer.ASSIGN {
			return p.parseAssignStmt()
		}
		if p.peekAt(1).Kind == lexer.STAR {
			return p.parseRepeatStmtRef()
		}
		name := p.advance()
		return &Ref{Name: name.Lit, Pos: name.Pos}, nil
	default:
		return nil, p.errf("unexpected token %v %q at start of statement", p.kind(), p.cur().Lit)
	}
}

func (p *parser) parseAssignStmt() (*AssignStmt, error) {
	pos := p.cur().Pos
	name := p.advance() // IDENT
	p.advance()         // '='
	g, err := p.parseGroupExpr()
	if err != nil {
		return nil, err
	}
	return &AssignStmt{Name: name.Lit, Group: g, Pos: pos}, nil
}

func (p *parser) parseRepeatStmtRef() (*RepeatStmt, error) {
	pos := p.cur().Pos
	name := p.advance() // IDENT
	p.advance()         // '*'
	n, err := p.parseNumberLiteral()
	if err != nil {
		return nil, err
	}
	return &RepeatStmt{Repeatable: &Ref{Name: name.Lit, Pos: name.Pos}, N: int(n), Pos: pos}, nil
}

func (p *parser) parseRepeatStmtParen() (*RepeatStmt, error) {
	pos := p.cur().Pos
	p.advance() // '('
	var stmts []any
	for {
		s, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, s)
		if p.at(lexer.COMMA) {
			p.advance()
			continue
		}
		break
	}
	if _, err := p.expect(lexer.RPAREN); err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.STAR); err != nil {
		return nil, err
	}
	n, err := p.parseNumberLiteral()
	if err != nil {
		return nil, err
	}
	return &RepeatStmt{Repeatable: &StmtList{Stmts: stmts}, N: int(n), Pos: pos}, nil
}

func (p *parser) parseRestStmt() (*RestStmt, error) {
	pos := p.cur().Pos
	p.advance() // 'rest'
	if _, err := p.expect(lexer.LPAREN); err != nil {
		return nil, err
	}
	d, err := p.parseDuration()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.RPAREN); err != nil {
		return nil, err
	}
	return &RestStmt{Duration: d, Pos: pos}, nil
}

func (p *parser) parseInlineCall() (*InlineCall, error) {
	pos := p.cur().Pos
	catalog := p.advance() // CATALOG_NAME
	if _, err := p.expect(lexer.LPAREN); err != nil {
		return nil, err
	}
	target, err := p.parseQuantity()
	if err != nil {
		return nil, err
	}
	var load *Quantity
	if p.at(lexer.AT) {
		p.advance()
		q, err := p.parseQuantity()
		if err != nil {
			return nil, err
		}
		load = &q
	}
	if _, err := p.expect(lexer.RPAREN); err != nil {
		return nil, err
	}
	return &InlineCall{Catalog: catalog.Lit, Target: target, Load: load, Pos: pos}, nil
}

// ---- group expressions ----

func (p *parser) parseGroupExpr() (GroupExpr, error) {
	switch p.kind() {
	case lexer.SUPERSET, lexer.CIRCUIT:
		return p.parseSupersetExpr()
	case lexer.EMOM:
		return p.parseCadenceExpr(1)
	case lexer.NUMBER:
		n, err := p.parseNumberLiteral()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.STAR); err != nil {
			return nil, err
		}
		if !p.at(lexer.EMOM) {
			return nil, p.errf("expected 'emom' after 'N*', got %v", p.kind())
		}
		return p.parseCadenceExpr(int(n))
	case lexer.AMRAP:
		return p.parseAmrapExpr()
	case lexer.FOR_TIME:
		return p.parseForTimeExpr()
	default:
		return nil, p.errf("expected a group expression, got %v", p.kind())
	}
}

func (p *parser) parseSupersetExpr() (*SupersetExpr, error) {
	pos := p.cur().Pos
	kindTok := p.advance() // 'superset'|'circuit'
	if _, err := p.expect(lexer.LPAREN); err != nil {
		return nil, err
	}
	var args []SupersetArg
	for {
		if p.at(lexer.REST) {
			r, err := p.parseRestStmt()
			if err != nil {
				return nil, err
			}
			args = append(args, SupersetArg{Rest: r})
		} else {
			tok, err := p.expect(lexer.IDENT)
			if err != nil {
				return nil, err
			}
			args = append(args, SupersetArg{Ref: &Ref{Name: tok.Lit, Pos: tok.Pos}})
		}
		if p.at(lexer.COMMA) {
			p.advance()
			continue
		}
		break
	}
	if _, err := p.expect(lexer.RPAREN); err != nil {
		return nil, err
	}
	return &SupersetExpr{Kind: kindTok.Lit, Args: args, Pos: pos}, nil
}

func (p *parser) parseCadenceExpr(n int) (*CadenceExpr, error) {
	pos := p.cur().Pos
	p.advance() // 'emom'
	if _, err := p.expect(lexer.LPAREN); err != nil {
		return nil, err
	}
	d, err := p.parseDuration()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.RPAREN); err != nil {
		return nil, err
	}
	body, err := p.parseGroupBody()
	if err != nil {
		return nil, err
	}
	return &CadenceExpr{N: n, Duration: d, Body: body, Pos: pos}, nil
}

func (p *parser) parseAmrapExpr() (*AmrapExpr, error) {
	pos := p.cur().Pos
	p.advance() // 'amrap'
	if _, err := p.expect(lexer.LPAREN); err != nil {
		return nil, err
	}
	d, err := p.parseDuration()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.RPAREN); err != nil {
		return nil, err
	}
	body, err := p.parseGroupBody()
	if err != nil {
		return nil, err
	}
	return &AmrapExpr{Duration: d, Body: body, Pos: pos}, nil
}

func (p *parser) parseForTimeExpr() (*ForTimeExpr, error) {
	pos := p.cur().Pos
	p.advance() // 'for_time'
	if _, err := p.expect(lexer.LBRACE); err != nil {
		return nil, err
	}
	if p.at(lexer.ROUNDS) {
		r, err := p.parseRoundsExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.RBRACE); err != nil {
			return nil, err
		}
		return &ForTimeExpr{Rounds: r, Pos: pos}, nil
	}
	var body []any
	for !p.at(lexer.RBRACE) {
		s, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		body = append(body, s)
		p.skipOptSemi()
	}
	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}
	return &ForTimeExpr{Body: body, Pos: pos}, nil
}

func (p *parser) parseRoundsExpr() (*RoundsExpr, error) {
	pos := p.cur().Pos
	p.advance() // 'rounds'
	if _, err := p.expect(lexer.LBRACKET); err != nil {
		return nil, err
	}
	var values []int
	for {
		v, err := p.parseNumberLiteral()
		if err != nil {
			return nil, err
		}
		values = append(values, int(v))
		if p.at(lexer.COMMA) {
			p.advance()
			continue
		}
		break
	}
	if _, err := p.expect(lexer.RBRACKET); err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.AS); err != nil {
		return nil, err
	}
	as, err := p.expect(lexer.IDENT)
	if err != nil {
		return nil, err
	}
	body, err := p.parseGroupBody()
	if err != nil {
		return nil, err
	}
	return &RoundsExpr{Values: values, As: as.Lit, Body: body, Pos: pos}, nil
}

func (p *parser) parseGroupBody() ([]any, error) {
	if _, err := p.expect(lexer.LBRACE); err != nil {
		return nil, err
	}
	var body []any
	for !p.at(lexer.RBRACE) {
		s, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		body = append(body, s)
		p.skipOptSemi()
	}
	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}
	return body, nil
}

// ---- expressions & quantities ----

func (p *parser) parseQuantity() (Quantity, error) {
	v, err := p.parseTerm()
	if err != nil {
		return Quantity{}, err
	}
	plus := false
	if p.at(lexer.PLUS) {
		p.advance()
		plus = true
	}
	return Quantity{Value: v, Plus: plus}, nil
}

func (p *parser) parseExpr() (Expr, error) { return p.parseTerm() }

func (p *parser) parseTerm() (Expr, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for p.at(lexer.STAR) {
		p.advance()
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		left = &MulExpr{Left: left, Right: right}
	}
	return left, nil
}

func (p *parser) parseFactor() (Expr, error) {
	switch p.kind() {
	case lexer.NUMBER:
		pos := p.cur().Pos
		v, err := p.parseNumberLiteral()
		if err != nil {
			return nil, err
		}
		unit := ""
		if p.at(lexer.IDENT) {
			if u, ok := quantityUnits[p.cur().Lit]; ok {
				unit = u
				p.advance()
			}
		}
		return &NumberLit{Value: v, Unit: unit, Pos: pos}, nil
	case lexer.CATALOG_NAME:
		pos := p.cur().Pos
		catalog := p.advance()
		if _, err := p.expect(lexer.DOT); err != nil {
			return nil, err
		}
		field, err := p.expect(lexer.IDENT)
		if err != nil {
			return nil, err
		}
		return &CatalogFieldRef{Catalog: catalog.Lit, Field: field.Lit, Pos: pos}, nil
	case lexer.IDENT:
		return p.parseDottedPath()
	case lexer.LPAREN:
		p.advance()
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.RPAREN); err != nil {
			return nil, err
		}
		return e, nil
	default:
		return nil, p.errf("expected a number, identifier, catalog field, or '(', got %v", p.kind())
	}
}

// quantityUnits is grammar's `unit` production. Bare "m" means meters
// here — see grammar.ebnf's note under `dottedPath`/unit.
var quantityUnits = map[string]string{
	"kg": "kg", "lb": "lb", "m": "m", "km": "km", "s": "s", "min": "min", "h": "h", "d": "d",
}

// durationUnits is grammar's `duration` production's unit set. Bare "m"
// AND "min" both mean minutes here, normalized to "m" (the only spelling
// canonical-form.schema.json's duration.unit enum accepts) — never
// meters, unlike quantityUnits above; the two vocabularies are never
// used in the same slot.
var durationUnits = map[string]string{
	"s": "s", "m": "m", "min": "m", "h": "h", "d": "d",
}

func (p *parser) parseDuration() (Duration, error) {
	v, err := p.parseNumberLiteral()
	if err != nil {
		return Duration{}, err
	}
	if !p.at(lexer.IDENT) {
		return Duration{}, p.errf("expected a duration unit (s/m/min/h/d), got %v", p.kind())
	}
	u, ok := durationUnits[p.cur().Lit]
	if !ok {
		return Duration{}, p.errf("unknown duration unit %q", p.cur().Lit)
	}
	p.advance()
	return Duration{Value: v, Unit: u}, nil
}

func (p *parser) parseDottedPath() (*DottedPath, error) {
	pos := p.cur().Pos
	first, err := p.expect(lexer.IDENT)
	if err != nil {
		return nil, err
	}
	segs := []string{first.Lit}
	for p.at(lexer.DOT) {
		p.advance()
		seg, err := p.expect(lexer.IDENT)
		if err != nil {
			return nil, err
		}
		segs = append(segs, seg.Lit)
	}
	return &DottedPath{Segments: segs, Pos: pos}, nil
}

func (p *parser) parseNumberLiteral() (float64, error) {
	tok, err := p.expect(lexer.NUMBER)
	if err != nil {
		return 0, err
	}
	var v float64
	if _, scanErr := fmt.Sscanf(tok.Lit, "%g", &v); scanErr != nil {
		return 0, &SyntaxError{Msg: fmt.Sprintf("invalid number %q", tok.Lit), Pos: tok.Pos}
	}
	return v, nil
}
