package token

import (
	"testing"

	"github.com/reflexionhealth/vanilla/expect"
)

func TestPositionIsValid(t *testing.T) {
	var pos Position

	pos = Position{"", 0, 0, 0}
	expect.False(t, pos.IsValid())

	pos = Position{"", 0, 1, 0}
	expect.True(t, pos.IsValid())
}

func TestPositionString(t *testing.T) {
	var pos Position

	pos = Position{"", 0, 0, 0}
	expect.Equal(t, pos.String(), "-")

	pos = Position{"Src", 0, 1, 1}
	expect.Equal(t, pos.String(), "Src:1:1")

	pos = Position{"Name", 15, 7, 16}
	expect.Equal(t, pos.String(), "Name:7:16")
}

func TestLookup(t *testing.T) {
	// an arbitrary string
	expect.Equal(t, Lookup("something"), IDENT)

	// all keyword tokens and no non-keyword tokens
	for i, name := range tokens {
		if len(name) > 0 {
			tok := Token(i)
			if tok.IsKeyword() {
				expect.Equal(t, Lookup(name), tok)
			} else {
				expect.Equal(t, Lookup(name), IDENT)
			}
		}
	}

	// case-insensitive
	expect.Equal(t, Lookup("SELECT"), SELECT)
	expect.Equal(t, Lookup("Select"), SELECT)
	expect.Equal(t, Lookup("select"), SELECT)
	expect.Equal(t, Lookup("sElECt"), SELECT)
	expect.Equal(t, Lookup("selecT"), SELECT)
	expect.Equal(t, Lookup("WHERE"), WHERE)
	expect.Equal(t, Lookup("where"), WHERE)
}

func TestTokenString(t *testing.T) {
	expect.Equal(t, INVALID.String(), "Invalid token")
	expect.Equal(t, EOS.String(), "End of statement")
	expect.Equal(t, COMMENT.String(), "Comment")

	expect.Equal(t, IDENT.String(), "Identifier")
	expect.Equal(t, QUOTED_IDENT.String(), "Quoted identifier")

	expect.Equal(t, STRING.String(), "String")
	expect.Equal(t, NUMBER.String(), "Number")

	expect.Equal(t, SEMICOLON.String(), ";")
	expect.Equal(t, COLON.String(), ":")
	expect.Equal(t, DOLLAR.String(), "$")
	expect.Equal(t, BANG.String(), "!")
	expect.Equal(t, EQUALS.String(), "=")
	expect.Equal(t, AT.String(), "@")
	expect.Equal(t, COMMA.String(), ",")
	expect.Equal(t, ASTERISK.String(), "*")
	expect.Equal(t, SLASH.String(), "/")
	expect.Equal(t, PERCENT.String(), "%")
	expect.Equal(t, PLUS.String(), "+")
	expect.Equal(t, MINUS.String(), "-")
	expect.Equal(t, PERIOD.String(), ".")
	expect.Equal(t, CONS.String(), "::")
	expect.Equal(t, LEFT_ANGLE.String(), "<")
	expect.Equal(t, RIGHT_ANGLE.String(), ">")
	expect.Equal(t, LEFT_EQUAL.String(), "<=")
	expect.Equal(t, RIGHT_EQUAL.String(), ">=")
	expect.Equal(t, BANG_EQUAL.String(), "!=")
	expect.Equal(t, LEFT_RIGHT.String(), "<>")

	expect.Equal(t, LEFT_PAREN.String(), "(")
	expect.Equal(t, LEFT_BRACKET.String(), "[")
	expect.Equal(t, RIGHT_PAREN.String(), ")")
	expect.Equal(t, RIGHT_BRACKET.String(), "]")

	expect.Equal(t, LEFT_PAREN.String(), "(")
	expect.Equal(t, LEFT_BRACKET.String(), "[")
	expect.Equal(t, RIGHT_PAREN.String(), ")")
	expect.Equal(t, RIGHT_BRACKET.String(), "]")

	expect.Equal(t, Token(2000).String(), "Token(2000)")
}

func TestHasLiteral(t *testing.T) {
	expect.Equal(t, INVALID.HasLiteral(), false)
	expect.Equal(t, EOS.HasLiteral(), false)
	expect.Equal(t, COMMENT.HasLiteral(), true)

	expect.Equal(t, IDENT.HasLiteral(), true)
	expect.Equal(t, QUOTED_IDENT.HasLiteral(), true)

	expect.Equal(t, STRING.HasLiteral(), true)
	expect.Equal(t, NUMBER.HasLiteral(), true)

	expect.Equal(t, SEMICOLON.HasLiteral(), false)
	expect.Equal(t, COLON.HasLiteral(), false)
	expect.Equal(t, DOLLAR.HasLiteral(), false)
	expect.Equal(t, BANG.HasLiteral(), false)
	expect.Equal(t, EQUALS.HasLiteral(), false)
	expect.Equal(t, AT.HasLiteral(), false)
	expect.Equal(t, COMMA.HasLiteral(), false)
	expect.Equal(t, ASTERISK.HasLiteral(), false)
	expect.Equal(t, QUESTION.HasLiteral(), false)
	expect.Equal(t, SLASH.HasLiteral(), false)
	expect.Equal(t, PERCENT.HasLiteral(), false)
	expect.Equal(t, PLUS.HasLiteral(), false)
	expect.Equal(t, MINUS.HasLiteral(), false)
	expect.Equal(t, PERIOD.HasLiteral(), false)
	expect.Equal(t, CONS.HasLiteral(), false)
	expect.Equal(t, LEFT_ANGLE.HasLiteral(), false)
	expect.Equal(t, RIGHT_ANGLE.HasLiteral(), false)
	expect.Equal(t, LEFT_EQUAL.HasLiteral(), false)
	expect.Equal(t, RIGHT_EQUAL.HasLiteral(), false)
	expect.Equal(t, BANG_EQUAL.HasLiteral(), false)
	expect.Equal(t, LEFT_RIGHT.HasLiteral(), false)

	expect.Equal(t, LEFT_PAREN.HasLiteral(), false)
	expect.Equal(t, LEFT_BRACKET.HasLiteral(), false)
	expect.Equal(t, RIGHT_PAREN.HasLiteral(), false)
	expect.Equal(t, RIGHT_BRACKET.HasLiteral(), false)
}

func TestIsKeyword(t *testing.T) {
	expect.Equal(t, INVALID.IsKeyword(), false)
	expect.Equal(t, EOS.IsKeyword(), false)
	expect.Equal(t, COMMENT.IsKeyword(), false)

	expect.Equal(t, IDENT.IsKeyword(), false)
	expect.Equal(t, QUOTED_IDENT.IsKeyword(), false)

	expect.Equal(t, SEMICOLON.IsKeyword(), false)
	expect.Equal(t, COLON.IsKeyword(), false)
	expect.Equal(t, DOLLAR.IsKeyword(), false)
	expect.Equal(t, BANG.IsKeyword(), false)
	expect.Equal(t, EQUALS.IsKeyword(), false)
	expect.Equal(t, AT.IsKeyword(), false)
	expect.Equal(t, COMMA.IsKeyword(), false)
	expect.Equal(t, QUESTION.IsKeyword(), false)
	expect.Equal(t, ASTERISK.IsKeyword(), false)
	expect.Equal(t, SLASH.IsKeyword(), false)
	expect.Equal(t, PERCENT.IsKeyword(), false)
	expect.Equal(t, PLUS.IsKeyword(), false)
	expect.Equal(t, MINUS.IsKeyword(), false)
	expect.Equal(t, PERIOD.IsKeyword(), false)
	expect.Equal(t, CONS.IsKeyword(), false)
	expect.Equal(t, LEFT_ANGLE.IsKeyword(), false)
	expect.Equal(t, RIGHT_ANGLE.IsKeyword(), false)
	expect.Equal(t, LEFT_EQUAL.IsKeyword(), false)
	expect.Equal(t, RIGHT_EQUAL.IsKeyword(), false)
	expect.Equal(t, BANG_EQUAL.IsKeyword(), false)
	expect.Equal(t, LEFT_RIGHT.IsKeyword(), false)

	expect.Equal(t, LEFT_PAREN.IsKeyword(), false)
	expect.Equal(t, LEFT_BRACKET.IsKeyword(), false)
	expect.Equal(t, RIGHT_PAREN.IsKeyword(), false)
	expect.Equal(t, RIGHT_BRACKET.IsKeyword(), false)
}

func TestIsOperator(t *testing.T) {
	expect.Equal(t, INVALID.IsOperator(), false)
	expect.Equal(t, EOS.IsOperator(), false)
	expect.Equal(t, COMMENT.IsOperator(), false)

	expect.Equal(t, IDENT.IsOperator(), false)
	expect.Equal(t, QUOTED_IDENT.IsOperator(), false)

	expect.Equal(t, SEMICOLON.IsOperator(), false)
	expect.Equal(t, COLON.IsOperator(), false)
	expect.Equal(t, DOLLAR.IsOperator(), false)
	expect.Equal(t, AT.IsOperator(), false)
	expect.Equal(t, COMMA.IsOperator(), false)
	expect.Equal(t, QUESTION.IsOperator(), false)

	expect.Equal(t, BANG.IsOperator(), true)
	expect.Equal(t, EQUALS.IsOperator(), true)
	expect.Equal(t, ASTERISK.IsOperator(), true)
	expect.Equal(t, SLASH.IsOperator(), true)
	expect.Equal(t, PERCENT.IsOperator(), true)
	expect.Equal(t, PLUS.IsOperator(), true)
	expect.Equal(t, MINUS.IsOperator(), true)
	expect.Equal(t, PERIOD.IsOperator(), true)
	expect.Equal(t, CONS.IsOperator(), true)
	expect.Equal(t, LEFT_ANGLE.IsOperator(), true)
	expect.Equal(t, RIGHT_ANGLE.IsOperator(), true)
	expect.Equal(t, LEFT_EQUAL.IsOperator(), true)
	expect.Equal(t, RIGHT_EQUAL.IsOperator(), true)
	expect.Equal(t, BANG_EQUAL.IsOperator(), true)
	expect.Equal(t, LEFT_RIGHT.IsOperator(), true)

	expect.Equal(t, LEFT_PAREN.IsOperator(), false)
	expect.Equal(t, LEFT_BRACKET.IsOperator(), false)
	expect.Equal(t, RIGHT_PAREN.IsOperator(), false)
	expect.Equal(t, RIGHT_BRACKET.IsOperator(), false)

	expect.Equal(t, SELECT.IsOperator(), false)
	expect.Equal(t, INSERT.IsOperator(), false)
	expect.Equal(t, UPDATE.IsOperator(), false)
	expect.Equal(t, WHERE.IsOperator(), false)
	expect.Equal(t, GROUP.IsOperator(), false)
	expect.Equal(t, ORDER.IsOperator(), false)
	expect.Equal(t, HAVING.IsOperator(), false)

	expect.Equal(t, AND.IsOperator(), true)
	expect.Equal(t, OR.IsOperator(), true)
	expect.Equal(t, IS.IsOperator(), true)
	expect.Equal(t, NOT.IsOperator(), true)
	expect.Equal(t, IN.IsOperator(), true)
	expect.Equal(t, BETWEEN.IsOperator(), true)
	expect.Equal(t, OVERLAPS.IsOperator(), true)
	expect.Equal(t, LIKE.IsOperator(), true)
	expect.Equal(t, ILIKE.IsOperator(), true)
	expect.Equal(t, REGEXP.IsOperator(), true)
	expect.Equal(t, SIMILAR.IsOperator(), true)
}
