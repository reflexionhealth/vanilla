package parser

import (
	"testing"

	"github.com/reflexionhealth/vanilla/sql/sqltest/ast"
	"github.com/stretchr/testify/assert"
)

func TestParseError(t *testing.T) {
	var prsr Parser
	var stmt ast.Stmt
	var err error

	prsr = Make([]byte(`SELECT * WHERE`), ParseRuleset{})
	stmt, err = prsr.ParseStatement()
	assert.Nil(t, stmt)
	if assert.NotNil(t, err, "Expected a parsing error") {
		assert.Equal(t, "sql:1:15: Expected 'FROM' but received 'WHERE'.", err.Error())
	}

	prsr = Make([]byte(`~`), ParseRuleset{})
	stmt, err = prsr.ParseStatement()
	assert.Nil(t, stmt)
	if assert.NotNil(t, err, "Expected a parsing error") {
		assert.Equal(t, "sql:1:1: Unexpected character U+007E '~'", err.Error())
	}
}

func TestParseSelect(t *testing.T) {
	prsr := Make([]byte(`SELECT * FROM mytable`), ParseRuleset{})
	stmt, err := prsr.ParseStatement()
	assert.Nil(t, err)
	if slct, ok := stmt.(*ast.SelectStmt); assert.True(t, ok) {
		assert.Equal(t, ast.SELECT_ALL, slct.Type)
		assert.Equal(t, "mytable", slct.From.Name)
		assert.Empty(t, slct.Selection)
		assert.True(t, slct.Star)
	}
}
