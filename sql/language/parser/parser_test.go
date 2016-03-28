package parser

import (
	"testing"

	"github.com/reflexionhealth/vanilla/sql/language/ast"
	"github.com/stretchr/testify/assert"
)

func TestParseError(t *testing.T) {
	var prsr Parser
	var stmt ast.Stmt
	var err error

	prsr = Make([]byte(`mytable`), Ruleset{})
	stmt, err = prsr.ParseStatement()
	assert.Nil(t, stmt)
	if assert.NotNil(t, err, "expected a parsing error") {
		assert.Equal(t, "sql:1:8: expected 'SELECT, INSERT, or UPDATE' but received 'Identifier'", err.Error())
	}

	prsr = Make([]byte(`SELECT * WHERE`), Ruleset{})
	stmt, err = prsr.ParseStatement()
	assert.Nil(t, stmt)
	if assert.NotNil(t, err, "expected a parsing error") {
		assert.Equal(t, "sql:1:15: expected 'FROM' but received 'WHERE'", err.Error())
	}

	prsr = Make([]byte(`SELECT * FROM *`), Ruleset{})
	stmt, err = prsr.ParseStatement()
	assert.Nil(t, stmt)
	if assert.NotNil(t, err, "expected a parsing error") {
		assert.Equal(t, "sql:1:16: expected 'a table name' but received '*'", err.Error())
	}

	prsr = Make([]byte(`~`), Ruleset{})
	stmt, err = prsr.ParseStatement()
	assert.Nil(t, stmt)
	if assert.NotNil(t, err, "expected a parsing error") {
		assert.Equal(t, "sql:1:1: unexpected character U+007E '~'", err.Error())
	}
}

func TestParseSelect(t *testing.T) {
	prsr := Make([]byte(`SELECT * FROM mytable`), Ruleset{})
	stmt, err := prsr.ParseStatement()
	assert.Nil(t, err)
	if slct, ok := stmt.(*ast.SelectStmt); assert.True(t, ok) {
		assert.Equal(t, ast.SELECT_ALL, slct.Type)
		assert.Equal(t, "mytable", slct.From.Name)
		assert.Empty(t, slct.Selection)
		assert.True(t, slct.Star)
	}

	// disallow unimplmented clauses
	prsr = Make([]byte(`SELECT * FROM mytable PROCEDURE compute(foo)`), Ruleset{})
	stmt, err = prsr.ParseStatement()
	assert.Nil(t, stmt)
	if assert.NotNil(t, err) {
		assert.Equal(t, `sql:1:32: cannot parse statement; reached unimplemented clause at "PROCEDURE"`, err.Error())
	}

	// allow unimplmented clauses if someone says its ok
	prsr = Make([]byte(`SELECT * FROM mytable PROCEDURE compute(foo)`), Ruleset{AllowNotImplemented: true})
	stmt, err = prsr.ParseStatement()
	assert.Nil(t, err)
	if slct, ok := stmt.(*ast.SelectStmt); assert.True(t, ok) {
		assert.Equal(t, ast.SELECT_ALL, slct.Type)
		assert.Equal(t, "mytable", slct.From.Name)
		assert.Empty(t, slct.Selection)
		assert.True(t, slct.Star)
	}
}

func TestParseInsert(t *testing.T) {
	prsr := Make([]byte(`INSERT INTO mytable`), Ruleset{})
	stmt, err := prsr.ParseStatement()
	assert.Nil(t, stmt)
	if assert.NotNil(t, err) {
		assert.Equal(t, `sql:1:20: cannot parse statement; reached unimplemented clause at "mytable"`, err.Error())
	}
}

func TestParseUpdate(t *testing.T) {
	prsr := Make([]byte(`UPDATE mytable SET a = 1`), Ruleset{})
	stmt, err := prsr.ParseStatement()
	assert.Nil(t, stmt)
	if assert.NotNil(t, err) {
		assert.Equal(t, `sql:1:15: cannot parse statement; reached unimplemented clause at "mytable"`, err.Error())
	}
}
