package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func isStmt(i interface{}) bool {
	v, ok := i.(Stmt)
	if ok {
		v.ImplementsStmt()
	}
	return ok
}

func isExpr(i interface{}) bool {
	v, ok := i.(Expr)
	if ok {
		v.ImplementsExpr()
	}
	return ok
}

func TestTrivial(t *testing.T) {
	assert.True(t, isStmt(&SelectStmt{}))
	assert.True(t, isStmt(&InsertStmt{}))
	assert.True(t, isStmt(&UpdateStmt{}))
	assert.True(t, isExpr(&Identifier{}))
	assert.True(t, isExpr(&Literal{}))
}
