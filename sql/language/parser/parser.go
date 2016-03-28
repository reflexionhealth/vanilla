package parser

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/reflexionhealth/vanilla/sql/language/ast"
	"github.com/reflexionhealth/vanilla/sql/language/scanner"
	"github.com/reflexionhealth/vanilla/sql/language/token"
)

// A Ruleset specifies the dialect specific parsing rules for a SQL dialect
type Ruleset struct {
	ScanRules scanner.Ruleset
	Operators ast.OperatorSet

	// AllowNotImplemented controls whether the parser will barf an error
	// if it reaches a likely valid part of SQL syntax that just hasn't been
	// implemented in this parser yet.
	// Otherwise, it seeks to the end of the statement.
	AllowNotImplemented bool

	CanSelectDistinctRow bool
	CanSelectWithoutFrom bool

	Operator   ast.OperatorSet
	Initialize func(os *ast.OperatorSet)
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
	rules   Ruleset

	pos int         // next token offset
	tok token.Token // next token type
	lit string      // next token literal

	Trace io.Writer // output for trace (no output if nil)
}

// New will allocate initialize a parser when you don't want to allocate one yourself
func New(src []byte, rules Ruleset) *Parser {
	p := &Parser{}
	p.Init(src, rules)
	return p
}

// Init prepares the parser p to convert a text src into an ast.
func (p *Parser) Init(src []byte, rules Ruleset) {
	scanError := func(pos token.Position, msg string) { p.error(pos, msg) }
	p.scanner.Init(src, scanError, rules.ScanRules)
	p.rules = rules
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
			*err = fmt.Errorf("panic: %v", e)
		}
	}
}

func (p *Parser) error(pos token.Position, msg string) {
	if p.Trace != nil {
		fmt.Fprintf(p.Trace, " (error) %v\n", (&ParseError{pos, msg}).Error())
	}
	p.stopParsing(&ParseError{pos, msg})
}

func (p *Parser) expect(tok token.Token) {
	if p.tok != tok {
		p.error(p.scanner.Pos(), fmt.Sprintf(`expected '%v' but received '%v'`, tok, p.tok))
	}
	p.next()
}

func (p *Parser) expected(what string) {
	p.error(p.scanner.Pos(), fmt.Sprintf(`expected '%v' but received '%v'`, what, p.tok))
}

func (p *Parser) next() {
	if p.Trace != nil && (p.pos > 0 || p.tok != token.INVALID) {
		pc, _, line, _ := runtime.Caller(1)
		path := strings.Split(runtime.FuncForPC(pc).Name(), ".")
		name := path[len(path)-1]
		// ignore expect and expected
		if len(name) >= 6 && name[0:6] == "expect" {
			pc, _, line, _ = runtime.Caller(2)
			path = strings.Split(runtime.FuncForPC(pc).Name(), ".")
			name = path[len(path)-1]
		}
		caller := "Parser." + name
		lit := p.lit
		if len(lit) > 7 {
			lit = lit[0:6] + "~"
		}
		fmt.Fprintf(p.Trace, " %7.7s : %-14s @ %v:%v\n", lit, p.tok, caller, line)
	}

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
		stmt.Type = ast.DISTINCT
		p.next()
	case token.DISTINCTROW:
		if p.rules.CanSelectDistinctRow {
			stmt.Type = ast.DISTINCT_ROW
			p.next()
		} else {
			msg := `statement includes SELECT "DISTINCTROW", but CanSelectDistinctRow is false`
			p.error(p.scanner.Pos(), msg)
		}
	}

	if p.tok == token.ASTERISK {
		stmt.Star = true
		p.next()
	} else {
		stmt.Select = []ast.Expr{p.parseExpression()}
		for p.tok == token.COMMA {
			p.next() // eat comma
			stmt.Select = append(stmt.Select, p.parseExpression())
		}
	}

	// NOTE: The FROM clause is sometimes optional, but since this would be an
	// error in most common uses cases, the default will be that it is required
	// even for dialects where it is technically optional.
	if p.rules.CanSelectWithoutFrom && p.tok == token.EOS {
		return stmt
	}

	p.expect(token.FROM)
	switch p.tok {
	case token.IDENT:
		stmt.From = ast.Name(p.lit)
		p.next()
	case token.QUOTED_IDENT:
		stmt.From = ast.Quoted(p.lit)
		p.next()
	default:
		p.expected("a table name")
	}

	if p.tok == token.WHERE {
		p.next() // eat WHERE
		stmt.Where = p.parseExpression()
	}

	// if p.tok == token.GROUP {
	// 	panic("TODO: parse GROUP BY")
	// }
	//
	// if p.tok == token.HAVING {
	// 	panic("TODO: parse HAVING")
	// }
	//
	// if p.tok == token.ORDER {
	// 	panic("TODO: parse ORDER")
	// }
	//
	// if p.tok == token.LIMIT {
	// 	panic("TODO: parse LIMIT")
	// }

	p.eatUnimplemented("clause")
	return stmt
}

func (p *Parser) parseInsert() *ast.InsertStmt {
	p.expect(token.INSERT)
	p.expect(token.INTO)
	p.eatUnimplemented("clause")
	return nil
}

func (p *Parser) parseUpdate() *ast.UpdateStmt {
	p.expect(token.UPDATE)
	p.eatUnimplemented("clause")
	return nil
}

// parseExpression uses table-based operator parsing (see parseExprWithOperators)
func (p *Parser) parseExpression() ast.Expr {
	return p.parseExprWithOperators(ast.MinPrecedence)
}

func (p *Parser) parseExprWithOperators(precedence ast.OpPrecedence) ast.Expr {
	lhs := p.parseBaseExpression()
	if p.tok == token.LEFT_PAREN {
		// TODO: functions like MAX(), MIN(), AVERAGE()
		p.eatUnimplemented("expression")
	} else if !p.tok.IsOperator() {
		return lhs
	}

	op, exists := p.rules.Operators.Lookup(p.tok.String(), ast.Infix)
	if !exists {
		msg := `statement includes '` + p.tok.String() + `', but it is not defined as an operator`
		p.error(p.scanner.Pos(), msg)
	}

	consumable := ast.MaxPrecedence
	for (op.Kind == ast.Infix) &&
		(precedence <= op.Precedence && op.Precedence <= consumable) {

		p.next() // eat operator
		rhs := p.parseExprWithOperators(rightPrec(op))
		lhs = ast.Binary(lhs, op.Type, rhs)

		if p.tok.IsOperator() {
			var exists bool
			op, exists = p.rules.Operators.Lookup(p.tok.String(), ast.Infix)
			if !exists {
				msg := `statement includes '` + p.tok.String() + `', but it is not defined as an operator`
				p.error(p.scanner.Pos(), msg)
			}
			consumable = nextPrec(op)
		} else {
			break
		}
	}

	return lhs
}

func rightPrec(op ast.Operator) ast.OpPrecedence {
	if op.Assoc == ast.RightAssoc {
		return op.Precedence
	} else {
		return op.Precedence + 1
	}
}

func nextPrec(op ast.Operator) ast.OpPrecedence {
	if op.Assoc == ast.LeftAssoc {
		return op.Precedence
	} else {
		return op.Precedence - 1
	}
}

func (p *Parser) parseBaseExpression() ast.Expr {
	/* handle prefix operator */
	if p.tok.IsOperator() {
		op, exists := p.rules.Operators.Lookup(p.tok.String(), ast.Prefix)
		if !exists {
			msg := `statement includes '` + p.tok.String() + `', but it is not defined as a unary operator`
			p.error(p.scanner.Pos(), msg)
		}

		p.next() // eat operator
		expr := p.parseExprWithOperators(op.Precedence)
		return ast.Unary(op.Type, expr)
	}

	switch p.tok {
	case token.IDENT:
		ident := ast.Name(p.lit)
		p.next()
		return ident
	case token.QUOTED_IDENT:
		ident := ast.Quoted(p.lit)
		p.next()
		return ident
	case token.STRING, token.NUMBER:
		lit := ast.Lit(p.lit)
		p.next()
		return lit
	default:
		p.eatUnimplemented("expression")
		return nil
	}
}

// eatUnimplemented eats till the end of statement if AllowsNotImplemented is true
func (p *Parser) eatUnimplemented(what string) {
	if !p.rules.AllowNotImplemented && !(p.tok == token.EOS || p.tok == token.SEMICOLON) {
		var errorValue string
		if p.tok.HasLiteral() {
			errorValue = p.lit
		} else {
			errorValue = p.tok.String()
		}
		msg := `cannot parse statement; reached unimplemented ` + what + ` at '` + errorValue + `'`
		p.error(p.scanner.Pos(), msg)
	}

	// eat till the end of statement
	for p.tok != token.EOS {
		if p.tok == token.SEMICOLON {
			p.next()
			if p.tok != token.EOS {
				p.error(p.scanner.Pos(), `statement does not end at semicolon`)
			}
		}
		p.next()
	}
}
