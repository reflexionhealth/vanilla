package parser

import (
	"fmt"
	"github.com/reflexionhealth/vanilla/sql/sqltest/ast"
	"github.com/reflexionhealth/vanilla/sql/sqltest/scanner"
	"github.com/reflexionhealth/vanilla/sql/sqltest/token"
)

// A ParseRuleset specifies the dialect specific parsing rules for a SQL dialect
type ParseRuleset struct {
	ScanRules scanner.ScanRuleset

	CanSelectDistinctRow bool
	CanSelectWithoutFrom bool
}

type ParseError struct {
	Pos token.Position
	Msg string
}

func (e *ParseError) Error() string {
	return e.Pos.String() + ": " + e.Msg
}

// A parser holds the parser's internal state while processing
// a given text.  It can be allocated as part of another data
// structure but must be initialized via Init before use.
type Parser struct {
	scanner scanner.Scanner
	rules   ParseRuleset
	pos     int         // next token offset
	tok     token.Token // next token type
	lit     string      // next token literal
}

// Make initialize
func Make(src []byte, rules ParseRuleset) Parser {
	p := Parser{}
	p.Init(src, rules)
	return p
}

// Init prepares the parser p to convert a text src into an ast.
func (p *Parser) Init(src []byte, rules ParseRuleset) {
	scanError := func(pos token.Position, msg string) { p.error(pos, msg) }
	p.scanner.Init(src, scanError, rules.ScanRules)
}

// ParseStatement attempts to parse a statement or returns the first error found
func (p *Parser) ParseStatement() (stmt ast.Stmt, err error) {
	defer p.recoverStopped(&err)
	p.next() // scan first
	stmt = p.parseStatement()
	return
}

// A stopParsing panic is raised to indicate early termination.
//
// In most cases I consider panics to be a code smell when they are used for
// control flow.  In this case though, it is far easier to use a panic for
// early termination than it would be to return and check for errors everywhere.
type stopParsing struct{ err *ParseError }

func (p *Parser) stopParsing(err *ParseError) {
	panic(stopParsing{err})
}

func (p *Parser) recoverStopped(err *error) {
	if e := recover(); e != nil {
		if stop, ok := e.(stopParsing); ok {
			*err = stop.err
		} else {
			panic(e)
		}
	}
}

func (p *Parser) error(pos token.Position, msg string) {
	p.stopParsing(&ParseError{pos, msg})
}

func (p *Parser) expect(tok token.Token) {
	if p.tok != tok {
		p.error(p.scanner.Pos(), fmt.Sprintf(`Expected '%v' but received '%v'.`, tok, p.tok))
	}
	p.next()
}

func (p *Parser) expected(what string) {
	p.error(p.scanner.Pos(), fmt.Sprintf(`Expected '%v' but received '%v'.`, what, p.tok))
}

func (p *Parser) next() {
	p.pos, p.tok, p.lit = p.scanner.Scan()
}

func (p *Parser) parseStatement() ast.Stmt {
	switch p.tok {
	case token.SELECT:
		return p.parseSelect()
	case token.INSERT:
		return p.parseInsert()
	case token.UPDATE:
		return p.parseUpdate()
	default:
		p.expected("SELECT, INSERT, or UPDATE")
		return nil
	}
}

func (p *Parser) parseSelect() *ast.SelectStmt {
	p.expect(token.SELECT)
	stmt := &ast.SelectStmt{}
	stmt.Type = ast.SELECT_ALL
	switch p.tok {
	case token.ALL:
		p.next()
	case token.DISTINCT:
		stmt.Type = ast.SELECT_DISTINCT
		p.next()
	case token.DISTINCTROW:
		if p.rules.CanSelectDistinctRow {
			stmt.Type = ast.SELECT_DISTINCTROW
			p.next()
		} else {
			p.error(p.scanner.Pos(), `Query includes SELECT "DISTINCTROW", but CanSelectDistinctRow is false`)
			p.next()
		}
	}

	if p.tok == token.ASTERISK {
		stmt.Star = true
		p.next()
	} else {
		stmt.Selection = []ast.Expr{p.parseExpression()}
		for p.tok == token.COMMA {
			stmt.Selection = append(stmt.Selection, p.parseExpression())
		}
	}

	// NOTE: The FROM clause is sometimes optional, but since this would be an
	// error in most common uses cases, the default will be that it is required
	// even for dialects where it is technically optional.
	if p.rules.CanSelectWithoutFrom && p.tok == token.EOL {
		return stmt
	}

	p.expect(token.FROM)
	switch p.tok {
	case token.IDENT:
		stmt.From.Name = p.lit
		stmt.From.Quoted = false
		p.next()
	case token.QUOTED_IDENT:
		stmt.From.Name = p.lit
		stmt.From.Quoted = true
		p.next()
	default:
		p.expected("table_name")
	}

	if p.tok == token.WHERE {
		panic("TODO: Parse WHERE")
	}

	if p.tok == token.GROUP {
		panic("TODO: Parse GROUP BY")
	}

	if p.tok == token.HAVING {
		panic("TODO: Parse HAVING")
	}

	if p.tok == token.ORDER {
		panic("TODO: Parse ORDER")
	}

	if p.tok == token.LIMIT {
		panic("TODO: Parse LIMIT")
	}

	return stmt
}

func (p *Parser) parseInsert() *ast.InsertStmt {
	p.expect(token.INSERT)
	p.expect(token.INTO)
	panic("TODO: Parse INSERT")
}

func (p *Parser) parseUpdate() *ast.UpdateStmt {
	p.expect(token.UPDATE)
	panic("TODO: Parse UPDATE")
}

func (p *Parser) parseExpression() ast.Expr {
	switch p.tok {
	case token.IDENT:
		ident := &ast.Identifier{p.lit, false}
		p.next()
		return ident
	case token.QUOTED_IDENT:
		ident := &ast.Identifier{p.lit, true}
		p.next()
		return ident
	default:
		panic("TODO: Expected ident, expression parsing hasn't been implemented yet")
	}
}
