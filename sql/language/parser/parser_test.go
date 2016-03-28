package parser

import (
	"bytes"
	"testing"

	"github.com/reflexionhealth/vanilla/sql/language/ast"
	"github.com/stretchr/testify/assert"
)

func TestTraceParser(t *testing.T) {
	output := bytes.Buffer{}
	parser := Make([]byte(`SELECT * FROM table_with_long_name WHERE 3`), Ruleset{})
	parser.Trace = &output
	stmt, err := parser.ParseStatement()
	assert.NotNil(t, err, "expected a parsing error")
	assert.Nil(t, stmt)
	assert.Equal(t, `
  SELECT : SELECT         @ Parser.parseSelect:188
         : *              @ Parser.parseSelect:216
    FROM : FROM           @ Parser.parseSelect:234
 table_~ : Identifier     @ Parser.parseSelect:240
 (error) sql:1:41: cannot parse statement; reached unimplemented clause at "WHERE"
`[1:], output.String())
}

func TestParseErrors(t *testing.T) {
	examples := []struct {
		Input string
		Error string
	}{
		{Input: `mytable`,
			Error: `sql:1:8: expected 'SELECT, INSERT, or UPDATE' but received 'Identifier'`},
		{Input: `SELECT * WHERE`,
			Error: `sql:1:15: expected 'FROM' but received 'WHERE'`},
		{Input: `SELECT * FROM *`,
			Error: `sql:1:16: expected 'a table name' but received '*'`},
		{Input: "SELECT `mycolumn` FROM mytable", // backticks (w/ ansi ruleset)
			Error: "sql:1:8: unexpected character U+0060 '`'"},
		{Input: `SELECT DISTINCTROW * FROM mytable`, // distinctrow (w/ ansi ruleset)
			Error: `sql:1:19: statement includes SELECT "DISTINCTROW", but CanSelectDistinctRow is false`},
		{Input: `~`,
			Error: `sql:1:1: unexpected character U+007E '~'`},
		{Input: `SELECT * FROM foos; SELECT * FROM bars;`,
			Error: `sql:1:27: statement does not end at semicolon`},
		{Input: `SELECT * FROM mytable PROCEDURE compute(foo)`, // with HasLiteral
			Error: `sql:1:32: cannot parse statement; reached unimplemented clause at "PROCEDURE"`},
		{Input: `SELECT * FROM mytable +`, // without HasLiteral
			Error: `sql:1:24: cannot parse statement; reached unimplemented clause at "+"`},
	}

	for _, example := range examples {
		parser := Make([]byte(example.Input), Ruleset{})
		stmt, err := parser.ParseStatement()
		assert.Nil(t, stmt)
		if assert.NotNil(t, err, "expected a parsing error") {
			assert.Equal(t, example.Error, err.Error())
		}
	}
}

func TestParseSelect(t *testing.T) {
	examples := []struct {
		Input  string
		Rules  Ruleset
		Result ast.Stmt
	}{
		{Input: `SELECT * FROM mytable`,
			Rules: Ruleset{},
			Result: &ast.SelectStmt{
				Type: ast.SELECT_ALL,
				From: ast.Identifier{Name: "mytable"},
				Star: true,
			}},
		{Input: `SELECT * FROM mytable;`, // with semicolon
			Rules: Ruleset{},
			Result: &ast.SelectStmt{
				Type: ast.SELECT_ALL,
				From: ast.Identifier{Name: "mytable"},
				Star: true,
			}},
		{Input: `SELECT * FROM "mytable"`, // doublequotes (ansi)
			Rules: Ruleset{},
			Result: &ast.SelectStmt{
				Type: ast.SELECT_ALL,
				From: ast.Identifier{Name: "mytable", Quoted: true},
				Star: true,
			}},
		{Input: "SELECT * FROM `mytable`", // backticks (mysql)
			Rules: MysqlRuleset,
			Result: &ast.SelectStmt{
				Type: ast.SELECT_ALL,
				From: ast.Identifier{Name: "mytable", Quoted: true},
				Star: true,
			}},
		{Input: `SELECT foo, bar FROM mytable`,
			Rules: Ruleset{},
			Result: &ast.SelectStmt{
				Type: ast.SELECT_ALL,
				From: ast.Identifier{Name: "mytable"},
				Selection: []ast.Expr{
					&ast.Identifier{Name: "foo"},
					&ast.Identifier{Name: "bar"},
				},
			}},
		{Input: `SELECT "foo", "bar" FROM mytable`,
			Rules: Ruleset{},
			Result: &ast.SelectStmt{
				Type: ast.SELECT_ALL,
				From: ast.Identifier{Name: "mytable"},
				Selection: []ast.Expr{
					&ast.Identifier{Name: "foo", Quoted: true},
					&ast.Identifier{Name: "bar", Quoted: true},
				},
			}},
		{Input: `SELECT * FROM mytable`,
			Rules: Ruleset{},
			Result: &ast.SelectStmt{
				Type: ast.SELECT_ALL,
				From: ast.Identifier{Name: "mytable"},
				Star: true,
			}},
		{Input: `SELECT ALL * FROM mytable`,
			Rules: Ruleset{},
			Result: &ast.SelectStmt{
				Type: ast.SELECT_ALL,
				From: ast.Identifier{Name: "mytable"},
				Star: true,
			}},
		{Input: `SELECT DISTINCT * FROM mytable`,
			Rules: Ruleset{},
			Result: &ast.SelectStmt{
				Type: ast.SELECT_DISTINCT,
				From: ast.Identifier{Name: "mytable"},
				Star: true,
			}},
		{Input: `SELECT DISTINCTROW * FROM mytable`,
			Rules: MysqlRuleset,
			Result: &ast.SelectStmt{
				Type: ast.SELECT_DISTINCTROW,
				From: ast.Identifier{Name: "mytable"},
				Star: true,
			}},

		// allow table-less select if someone says its ok
		{Input: `SELECT *`, // TODO: eventually I'd like this to be `SELECT 1+1;`
			Rules:  Ruleset{CanSelectWithoutFrom: true},
			Result: &ast.SelectStmt{Type: ast.SELECT_ALL, Star: true}},

		// allow unimplmented clauses if someone says its ok
		{Input: `SELECT * FROM mytable PROCEDURE compute(foo)`,
			Rules: Ruleset{AllowNotImplemented: true},
			Result: &ast.SelectStmt{
				Type: ast.SELECT_ALL,
				From: ast.Identifier{Name: "mytable"},
				Star: true,
			}},
	}

	for _, example := range examples {
		parser := Make([]byte(example.Input), example.Rules)
		stmt, err := parser.ParseStatement()
		assert.Nil(t, err)
		assert.Equal(t, example.Result, stmt)
	}
}

func TestParseInsert(t *testing.T) {
	parser := Make([]byte(`INSERT INTO mytable`), Ruleset{})
	stmt, err := parser.ParseStatement()
	assert.Nil(t, stmt)
	if assert.NotNil(t, err) {
		assert.Equal(t, `sql:1:20: cannot parse statement; reached unimplemented clause at "mytable"`, err.Error())
	}
}

func TestParseUpdate(t *testing.T) {
	parser := Make([]byte(`UPDATE mytable SET a = 1`), Ruleset{})
	stmt, err := parser.ParseStatement()
	assert.Nil(t, stmt)
	if assert.NotNil(t, err) {
		assert.Equal(t, `sql:1:15: cannot parse statement; reached unimplemented clause at "mytable"`, err.Error())
	}
}
