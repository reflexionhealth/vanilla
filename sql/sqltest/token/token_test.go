package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPositionIsValid(t *testing.T) {
	var pos Position

	pos = Position{"", 0, 0, 0}
	assert.False(t, pos.IsValid())

	pos = Position{"", 0, 1, 0}
	assert.True(t, pos.IsValid())
}

func TestPositionString(t *testing.T) {
	var pos Position

	pos = Position{"", 0, 0, 0}
	assert.Equal(t, "-", pos.String())

	pos = Position{"Src", 0, 1, 1}
	assert.Equal(t, "Src:1:1", pos.String())

	pos = Position{"Name", 15, 7, 16}
	assert.Equal(t, "Name:7:16", pos.String())
}

func TestLookup(t *testing.T) {
	// an arbitrary string
	assert.Equal(t, IDENT, Lookup("something"))

	// all keyword tokens and no non-keyword tokens
	for i, name := range tokens {
		tok := Token(i)
		if tok.IsKeyword() {
			assert.Equal(t, tok, Lookup(name))
		} else {
			assert.Equal(t, IDENT, Lookup(name))
		}
	}

	// case-insensitive
	assert.Equal(t, SELECT, Lookup("SELECT"))
	assert.Equal(t, SELECT, Lookup("Select"))
	assert.Equal(t, SELECT, Lookup("select"))
	assert.Equal(t, SELECT, Lookup("sElECt"))
	assert.Equal(t, SELECT, Lookup("selecT"))
	assert.Equal(t, WHERE, Lookup("WHERE"))
	assert.Equal(t, WHERE, Lookup("where"))
}

func TestTokenString(t *testing.T) {
	assert.Equal(t, "Invalid token", INVALID.String())
	assert.Equal(t, "EOL", EOL.String())
	assert.Equal(t, "Comment", COMMENT.String())

	assert.Equal(t, "Identifier", IDENT.String())
	assert.Equal(t, "Quoted identifier", QUOTED_IDENT.String())

	assert.Equal(t, "String", STRING.String())
	assert.Equal(t, "Number", NUMBER.String())

	assert.Equal(t, ";", SEMICOLON.String())
	assert.Equal(t, ":", COLON.String())
	assert.Equal(t, "$", DOLLAR.String())
	assert.Equal(t, "!", BANG.String())
	assert.Equal(t, "=", EQUALS.String())
	assert.Equal(t, "@", AT.String())
	assert.Equal(t, ",", COMMA.String())
	assert.Equal(t, "*", ASTERISK.String())
	assert.Equal(t, "/", SLASH.String())
	assert.Equal(t, "%", PERCENT.String())
	assert.Equal(t, "+", PLUS.String())
	assert.Equal(t, "-", MINUS.String())
	assert.Equal(t, ".", PERIOD.String())

	assert.Equal(t, "(", LEFT_PAREN.String())
	assert.Equal(t, "[", LEFT_BRACKET.String())
	assert.Equal(t, ")", RIGHT_PAREN.String())
	assert.Equal(t, "]", RIGHT_BRACKET.String())

	assert.Equal(t, "(", LEFT_PAREN.String())
	assert.Equal(t, "[", LEFT_BRACKET.String())
	assert.Equal(t, ")", RIGHT_PAREN.String())
	assert.Equal(t, "]", RIGHT_BRACKET.String())

	assert.Equal(t, "Token(2000)", Token(2000).String())
}

func TestHasLiteral(t *testing.T) {
	assert.Equal(t, false, INVALID.HasLiteral())
	assert.Equal(t, false, EOL.HasLiteral())
	assert.Equal(t, true, COMMENT.HasLiteral())

	assert.Equal(t, true, IDENT.HasLiteral())
	assert.Equal(t, true, QUOTED_IDENT.HasLiteral())

	assert.Equal(t, true, STRING.HasLiteral())
	assert.Equal(t, true, NUMBER.HasLiteral())

	assert.Equal(t, false, SEMICOLON.HasLiteral())
	assert.Equal(t, false, COLON.HasLiteral())
	assert.Equal(t, false, DOLLAR.HasLiteral())
	assert.Equal(t, false, BANG.HasLiteral())
	assert.Equal(t, false, EQUALS.HasLiteral())
	assert.Equal(t, false, AT.HasLiteral())
	assert.Equal(t, false, COMMA.HasLiteral())
	assert.Equal(t, false, ASTERISK.HasLiteral())
	assert.Equal(t, false, QUESTION.HasLiteral())
	assert.Equal(t, false, SLASH.HasLiteral())
	assert.Equal(t, false, PERCENT.HasLiteral())
	assert.Equal(t, false, PLUS.HasLiteral())
	assert.Equal(t, false, MINUS.HasLiteral())
	assert.Equal(t, false, PERIOD.HasLiteral())

	assert.Equal(t, false, LEFT_PAREN.HasLiteral())
	assert.Equal(t, false, LEFT_BRACKET.HasLiteral())
	assert.Equal(t, false, RIGHT_PAREN.HasLiteral())
	assert.Equal(t, false, RIGHT_BRACKET.HasLiteral())
}

func TestIsKeyword(t *testing.T) {
	assert.Equal(t, false, INVALID.IsKeyword())
	assert.Equal(t, false, EOL.IsKeyword())
	assert.Equal(t, false, COMMENT.IsKeyword())

	assert.Equal(t, false, IDENT.IsKeyword())
	assert.Equal(t, false, QUOTED_IDENT.IsKeyword())

	assert.Equal(t, false, SEMICOLON.IsKeyword())
	assert.Equal(t, false, COLON.IsKeyword())
	assert.Equal(t, false, DOLLAR.IsKeyword())
	assert.Equal(t, false, BANG.IsKeyword())
	assert.Equal(t, false, EQUALS.IsKeyword())
	assert.Equal(t, false, AT.IsKeyword())
	assert.Equal(t, false, COMMA.IsKeyword())
	assert.Equal(t, false, ASTERISK.IsKeyword())
	assert.Equal(t, false, SLASH.IsKeyword())
	assert.Equal(t, false, PERCENT.IsKeyword())
	assert.Equal(t, false, PLUS.IsKeyword())
	assert.Equal(t, false, MINUS.IsKeyword())
	assert.Equal(t, false, PERIOD.IsKeyword())

	assert.Equal(t, false, LEFT_PAREN.IsKeyword())
	assert.Equal(t, false, LEFT_BRACKET.IsKeyword())
	assert.Equal(t, false, RIGHT_PAREN.IsKeyword())
	assert.Equal(t, false, RIGHT_BRACKET.IsKeyword())
}
