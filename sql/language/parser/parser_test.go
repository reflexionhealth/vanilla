package parser

import (
	"bytes"
	"testing"

	"github.com/reflexionhealth/vanilla/sql/language/ast"
	"github.com/stretchr/testify/assert"
)

func TestParseErrors(t *testing.T) {
	examples := []struct {
		Statement string
		Error     string
	}{
		{Statement: `mytable`,
			Error: `sql:1:8: expected 'SELECT, INSERT, or UPDATE' but received 'Identifier'`},
		{Statement: `SELECT * WHERE`,
			Error: `sql:1:15: expected 'FROM' but received 'WHERE'`},
		{Statement: `SELECT * FROM *`,
			Error: `sql:1:16: expected 'a table name' but received '*'`},
		{Statement: `~`,
			Error: `sql:1:1: unexpected character U+007E '~'`},
		{Statement: `SELECT * FROM foos; SELECT * FROM bars;`,
			Error: `sql:1:27: statement does not end at semicolon`},
		{Statement: `SELECT * FROM mytable PROCEDURE compute(foo)`, // with HasLiteral
			Error: `sql:1:32: cannot parse statement; reached unimplemented clause at "PROCEDURE"`},
		{Statement: `SELECT * FROM mytable +`, // without HasLiteral
			Error: `sql:1:24: cannot parse statement; reached unimplemented clause at "+"`},
	}

	for _, example := range examples {
		parser := Make([]byte(example.Statement), Ruleset{})
		stmt, err := parser.ParseStatement()
		assert.Nil(t, stmt)
		if assert.NotNil(t, err, "expected a parsing error") {
			assert.Equal(t, example.Error, err.Error())
		}
	}
}

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

func TestParseSelect(t *testing.T) {
	parser := Make([]byte(`SELECT * FROM mytable`), Ruleset{})
	stmt, err := parser.ParseStatement()
	assert.Nil(t, err)
	if slct, ok := stmt.(*ast.SelectStmt); assert.True(t, ok) {
		assert.Equal(t, ast.SELECT_ALL, slct.Type)
		assert.Equal(t, "mytable", slct.From.Name)
		assert.Empty(t, slct.Selection)
		assert.True(t, slct.Star)
	}

	// allow unimplmented clauses if someone says its ok
	parser = Make([]byte(`SELECT * FROM mytable PROCEDURE compute(foo)`), Ruleset{AllowNotImplemented: true})
	stmt, err = parser.ParseStatement()
	assert.Nil(t, err)
	if slct, ok := stmt.(*ast.SelectStmt); assert.True(t, ok) {
		assert.Equal(t, ast.SELECT_ALL, slct.Type)
		assert.Equal(t, "mytable", slct.From.Name)
		assert.Empty(t, slct.Selection)
		assert.True(t, slct.Star)
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
