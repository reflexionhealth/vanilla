package parser

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/reflexionhealth/vanilla/expect"
	"github.com/reflexionhealth/vanilla/sql/language/ast"
	"github.com/reflexionhealth/vanilla/utils"
)

func TestTraceParser(t *testing.T) {
	output := bytes.Buffer{}
	parser := New([]byte(`SELECT * FROM table_with_long_name WHERE ♫`), Ruleset{})
	parser.Trace = &output
	stmt, err := parser.ParseStatement()
	expect.NotNil(t, err, "expected a parsing error")
	expect.Nil(t, stmt)

	expected := []string{
		regexp.QuoteMeta(`  SELECT : SELECT         @ Parser.parseSelect:`) + "[0-9]+",
		regexp.QuoteMeta(`         : *              @ Parser.parseSelect:`) + "[0-9]+",
		regexp.QuoteMeta(`    FROM : FROM           @ Parser.parseSelect:`) + "[0-9]+",
		regexp.QuoteMeta(` table_~ : Identifier     @ Parser.parseSelect:`) + "[0-9]+",
		regexp.QuoteMeta(`   WHERE : WHERE          @ Parser.parseSelect:`) + "[0-9]+",
		regexp.QuoteMeta(` (error) sql:1:42: unexpected character U+266B '♫'`),
		"$", // string ends with newline
	}

	// compare trace output, ignoring the source line numbers
	lines := strings.Split(output.String(), "\n")
	if expect.Equal(t, len(lines), len(expected)) {
		maxSafe := utils.MinInt(len(expected), len(lines))
		for i := 0; i < maxSafe; i++ {
			expect.Regexp(t, lines[i], "^"+expected[i])
		}
	} else {
		t.Log("Error:", err)
		t.Log("Full trace output:\n", output.String())
	}

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
			Error: `sql:1:32: cannot parse statement; reached unimplemented clause at 'PROCEDURE'`},
		{Input: `SELECT * FROM mytable +`, // without HasLiteral
			Error: `sql:1:24: cannot parse statement; reached unimplemented clause at '+'`},
	}

	for _, example := range examples {
		parser := New([]byte(example.Input), Ruleset{})
		stmt, err := parser.ParseStatement()
		expect.Nil(t, stmt)
		if expect.NotNil(t, err, "expected a parsing error") {
			expect.Equal(t, err.Error(), example.Error)
		}
	}
}

func TestParseSelect(t *testing.T) {
	examples := []struct {
		Input  string
		Rules  Ruleset
		Result ast.Stmt
		Trace  bool // for debugging
	}{
		{Input: `SELECT * FROM mytable`, // basics
			Result: &ast.SelectStmt{
				Type: ast.SELECT_ALL,
				From: ast.Name("mytable"),
				Star: true,
			}},
		{Input: `SELECT * FROM mytable;`, // with semicolon
			Result: &ast.SelectStmt{
				Type: ast.SELECT_ALL,
				From: ast.Name("mytable"),
				Star: true,
			}},
		{Input: `SELECT * FROM "mytable"`, // doublequotes (ansi)
			Result: &ast.SelectStmt{
				Type: ast.SELECT_ALL,
				From: ast.Quoted("mytable"),
				Star: true,
			}},
		{Input: "SELECT * FROM `mytable`", // backticks (mysql)
			Rules: MysqlRuleset,
			Result: &ast.SelectStmt{
				Type: ast.SELECT_ALL,
				From: ast.Quoted("mytable"),
				Star: true,
			}},
		{Input: `SELECT foo, bar FROM mytable`, // with columns
			Result: &ast.SelectStmt{
				Type:   ast.SELECT_ALL,
				From:   ast.Name("mytable"),
				Select: []ast.Expr{ast.Name("foo"), ast.Name("bar")},
			}},
		{Input: `SELECT "foo", "bar" FROM mytable`, // with quoted columns
			Result: &ast.SelectStmt{
				Type:   ast.SELECT_ALL,
				From:   ast.Name("mytable"),
				Select: []ast.Expr{ast.Quoted("foo"), ast.Quoted("bar")},
			}},
		{Input: `SELECT ALL * FROM mytable`, // ALL
			Result: &ast.SelectStmt{
				Type: ast.SELECT_ALL,
				From: ast.Name("mytable"),
				Star: true,
			}},
		{Input: `SELECT DISTINCT * FROM mytable`, // DISTINCT
			Result: &ast.SelectStmt{
				Type: ast.DISTINCT,
				From: ast.Name("mytable"),
				Star: true,
			}},
		{Input: `SELECT DISTINCTROW * FROM mytable`, // DISTINCT ROW (mysql)
			Rules: MysqlRuleset,
			Result: &ast.SelectStmt{
				Type: ast.DISTINCT_ROW,
				From: ast.Name("mytable"),
				Star: true,
			}},
		{Input: `SELECT * FROM mytable WHERE id = 3`, // simple WHERE clause
			Rules: AnsiRuleset,
			Result: &ast.SelectStmt{
				Type:  ast.SELECT_ALL,
				Star:  true,
				From:  ast.Name("mytable"),
				Where: ast.Binary(ast.Name("id"), ast.EQUAL, ast.Lit("3")),
			}},

		// a WHERE clause with nested expressions (binary and unary)
		{Input: `SELECT * FROM mytable WHERE kind != "muppet" AND 5 < -size`,
			Rules: MysqlRuleset,
			Result: &ast.SelectStmt{
				Type: ast.SELECT_ALL,
				Star: true,
				From: ast.Name("mytable"),
				Where: ast.Binary(
					ast.Binary(ast.Name("kind"), ast.NOT_EQUAL, ast.Lit(`"muppet"`)),
					ast.AND,
					ast.Binary(ast.Lit("5"), ast.LESS, ast.Unary(ast.NEGATE, ast.Name("size"))),
				),
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
				From: ast.Name("mytable"),
				Star: true,
			}},
	}

	for _, example := range examples {
		parser := New([]byte(example.Input), example.Rules)
		if example.Trace {
			parser.Trace = os.Stdout
		}
		stmt, err := parser.ParseStatement()
		expect.Nil(t, err, "Error for `"+example.Input+"`")
		expect.Equal(t, stmt, example.Result, example.Input)
	}
}

func TestParseInsert(t *testing.T) {
	parser := New([]byte(`INSERT INTO mytable`), Ruleset{})
	stmt, err := parser.ParseStatement()
	expect.Nil(t, stmt)
	if expect.NotNil(t, err) {
		expect.Equal(t, err.Error(), `sql:1:20: cannot parse statement; reached unimplemented clause at 'mytable'`)
	}
}

func TestParseUpdate(t *testing.T) {
	parser := New([]byte(`UPDATE mytable SET a = 1`), Ruleset{})
	stmt, err := parser.ParseStatement()
	expect.Nil(t, stmt)
	if expect.NotNil(t, err) {
		expect.Equal(t, err.Error(), `sql:1:15: cannot parse statement; reached unimplemented clause at 'mytable'`)
	}
}
