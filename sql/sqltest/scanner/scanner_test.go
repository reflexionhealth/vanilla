package scanner

import (
	"testing"

	"github.com/reflexionhealth/vanilla/sql/sqltest/token"
	"github.com/stretchr/testify/assert"
)

func TestSelect(t *testing.T) {
	query := `SELECT * FROM users WHERE id = 3`

	failOnError := func(pos token.Position, msg string) {
		assert.Fail(t, "At Line %d, Col %d: %s", pos.Line, pos.Column, msg)
	}

	s := Scanner{}
	s.Init([]byte(query), failOnError, ScanRuleset{})

	var tokens []token.Token
	MAX_ITER := 200 // Don't loop forever
	for i := 0; i < MAX_ITER; i++ {
		_, tok, _ := s.Scan()
		tokens = append(tokens, tok)
		if tok == token.INVALID || tok == token.EOF {
			break
		}
	}

	assert.Zero(t, s.ErrorCount)
	assert.Equal(t, []token.Token{
		// SELECT * FROM users
		token.SELECT, token.ASTERISK, token.FROM, token.IDENT,
		// WHERE id = 3
		token.WHERE, token.IDENT, token.EQUALS, token.NUMBER,
		token.EOF,
	}, tokens)
}

type scanToken struct {
	pos int
	tok token.Token
	lit string
}

type scanError struct {
	pos token.Position
	msg string
}

func scanOnce(src string) (scanToken, *scanError) {
	var err *scanError
	handleError := func(pos token.Position, msg string) {
		err = &scanError{pos, msg}
	}

	var t scanToken
	s := Scanner{}
	s.Init([]byte(src), handleError, ScanRuleset{})
	t.pos, t.tok, t.lit = s.Scan()

	return t, err
}

func scanOnceWith(src string, rules ScanRuleset) (scanToken, *scanError) {
	var err *scanError
	handleError := func(pos token.Position, msg string) {
		err = &scanError{pos, msg}
	}

	var t scanToken
	s := Scanner{}
	s.Init([]byte(src), handleError, rules)
	t.pos, t.tok, t.lit = s.Scan()

	return t, err
}

func scanAll(src string) *scanError {
	var err *scanError
	handleError := func(pos token.Position, msg string) {
		err = &scanError{pos, msg}
	}

	var t scanToken
	s := Scanner{}
	s.Init([]byte(src), handleError, ScanRuleset{})

	for i := 0; i < 9999; i++ {
		t.pos, t.tok, t.lit = s.Scan()
		if t.tok == token.EOF {
			break
		} else if err != nil {
			return err
		}
	}

	return nil
}

func TestSkipsWhitesace(t *testing.T) {
	scan, err := scanOnce("\n    SELECT\n")
	assert.Nil(t, err)
	assert.Equal(t, token.SELECT, scan.tok)
	assert.Equal(t, 5, scan.pos)
	assert.Equal(t, "SELECT", scan.lit)

	// scan, err = scanOnce("\n    --comment\n    SELECT--comment\n")
	// assert.Nil(t, err)
	// assert.Equal(t, token.SELECT, scan.tok)
	// assert.Equal(t, 18, scan.pos)
	// assert.Equal(t, "SELECT", scan.lit)
	//
	// scan, err = scanOnce("\n    --comment\r\n    SELECT--comment\n")
	// assert.Nil(t, err)
	// assert.Equal(t, token.SELECT, scan.tok)
	// assert.Equal(t, 19, scan.pos)
	// assert.Equal(t, "SELECT", scan.lit)
}

func TestErrorsRespectWhitespace(t *testing.T) {
	scan, err := scanOnce("\n\n    ~\n")
	assert.Equal(t, token.INVALID, scan.tok)
	if assert.NotNil(t, err) {
		assert.Equal(t, 6, err.pos.Offset)
		assert.Equal(t, 3, err.pos.Line)
		assert.Equal(t, 5, err.pos.Column)
		assert.Equal(t, `Unexpected character U+007E '~'`, err.msg)
	}
}

func TestScansIdentifier(t *testing.T) {
	scan, err := scanOnce(`simple`)
	assert.Nil(t, err)
	assert.Equal(t, token.IDENT, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, `simple`, scan.lit)

	scan, err = scanOnce("sim$ple")
	assert.Equal(t, token.IDENT, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, `sim`, scan.lit)

	scan, err = scanOnceWith("sim$ple", ScanRuleset{DollarIsLetter: true})
	assert.Equal(t, token.IDENT, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, `sim$ple`, scan.lit)
}

func TestScansQuotedIdentifier(t *testing.T) {
	scan, err := scanOnce(`"simple"`)
	assert.Nil(t, err)
	assert.Equal(t, token.QUOTED_IDENT, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, `simple`, scan.lit)

	scan, err = scanOnceWith(`"simple"`, ScanRuleset{DoubleQuoteIsNotQuotemark: true})
	if assert.NotNil(t, err) {
		assert.Equal(t, 0, err.pos.Offset)
		assert.Equal(t, 1, err.pos.Line)
		assert.Equal(t, 1, err.pos.Column)
		assert.Equal(t, `Unexpected character U+0022 '"'`, err.msg)
	}
}

func TestScansStrings(t *testing.T) {
	scan, err := scanOnce(`'simple'`)
	assert.Nil(t, err)
	assert.Equal(t, token.STRING, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, `'simple'`, scan.lit)

	scan, err = scanOnce(`' white space '`)
	assert.Nil(t, err)
	assert.Equal(t, token.STRING, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, `' white space '`, scan.lit)

	scan, err = scanOnce(`'quote\''`)
	assert.Nil(t, err)
	assert.Equal(t, token.STRING, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, `'quote\''`, scan.lit)

	scan, err = scanOnce(`'escaped \n\r\b\t\f'`)
	assert.Nil(t, err)
	assert.Equal(t, token.STRING, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, `'escaped \n\r\b\t\f'`, scan.lit)

	scan, err = scanOnce(`'slashes \\ \/'`)
	assert.Nil(t, err)
	assert.Equal(t, token.STRING, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, `'slashes \\ \/'`, scan.lit)
}

func TestReportsUsefulStringErrors(t *testing.T) {
	scan, err := scanOnce(`'`)
	assert.Equal(t, token.INVALID, scan.tok)
	if assert.NotNil(t, err) {
		assert.Equal(t, 0, err.pos.Offset)
		assert.Equal(t, 1, err.pos.Line)
		assert.Equal(t, 1, err.pos.Column)
		assert.Equal(t, `Unterminated string`, err.msg)
	}

	scan, err = scanOnce(`'No end quote`)
	assert.Equal(t, token.INVALID, scan.tok)
	if assert.NotNil(t, err) {
		assert.Equal(t, 0, err.pos.Offset)
		assert.Equal(t, 1, err.pos.Line)
		assert.Equal(t, 1, err.pos.Column)
		assert.Equal(t, `Unterminated string`, err.msg)
	}

	// scan, err = scanOnce("'contains unescaped \u0007 control char'")
	// assert.Equal(t, token.INVALID, scan.tok)
	// if assert.NotNil(t, err) {
	// 	assert.Equal(t, 0, err.pos.Offset)
	// 	assert.Equal(t, 1, err.pos.Line)
	// 	assert.Equal(t, 1, err.pos.Column)
	// 	assert.Equal(t, `Unexpected character in string: U+0007`, err.msg)
	// }

	// scan, err = scanOnce("'null-byte \u0000 in string'")
	// assert.Equal(t, token.INVALID, scan.tok)
	// if assert.NotNil(t, err) {
	// 	assert.Equal(t, 0, err.pos.Offset)
	// 	assert.Equal(t, 1, err.pos.Line)
	// 	assert.Equal(t, 1, err.pos.Column)
	// 	assert.Equal(t, `Unexpected character in string: U+0000`, err.msg)
	// }

	// scan, err = scanOnce(`'\u`)
	// assert.Equal(t, token.INVALID, scan.tok)
	// if assert.NotNil(t, err) {
	// 	assert.Equal(t, 2, err.pos.Offset)
	// 	assert.Equal(t, 1, err.pos.Line)
	// 	assert.Equal(t, 3, err.pos.Column)
	// 	assert.Equal(t, `Unterminated escape sequence`, err.msg)
	// }

	// scan, err = scanOnce(`'\`)
	// assert.Equal(t, token.INVALID, scan.tok)
	// if assert.NotNil(t, err) {
	// 	assert.Equal(t, 2, err.pos.Offset)
	// 	assert.Equal(t, 1, err.pos.Line)
	// 	assert.Equal(t, 3, err.pos.Column)
	// 	assert.Equal(t, `Unterminated escape sequence`, err.msg)
	// }

	// scan, err = scanOnce(`'\m'`)
	// assert.Equal(t, token.INVALID, scan.tok)
	// if assert.NotNil(t, err) {
	// 	assert.Equal(t, 2, err.pos.Offset)
	// 	assert.Equal(t, 1, err.pos.Line)
	// 	assert.Equal(t, 3, err.pos.Column)
	// 	assert.Equal(t, `Unknown escape sequence`, err.msg)
	// }

	// scan, err = scanOnce(`'\uD800'`)
	// assert.Equal(t, token.INVALID, scan.tok)
	// if assert.NotNil(t, err) {
	// 	assert.Equal(t, 2, err.pos.Offset)
	// 	assert.Equal(t, 1, err.pos.Line)
	// 	assert.Equal(t, 3, err.pos.Column)
	// 	assert.Equal(t, `Escape sequence is invalid Unicode code point`, err.msg)
	// }

	scan, err = scanOnce("'multi\nline'")
	assert.Equal(t, token.INVALID, scan.tok)
	if assert.NotNil(t, err) {
		assert.Equal(t, 0, err.pos.Offset)
		assert.Equal(t, 1, err.pos.Line)
		assert.Equal(t, 1, err.pos.Column)
		assert.Equal(t, `Unterminated string`, err.msg)
	}

	scan, err = scanOnce("'multi\rline'")
	assert.Equal(t, token.INVALID, scan.tok)
	if assert.NotNil(t, err) {
		assert.Equal(t, 0, err.pos.Offset)
		assert.Equal(t, 1, err.pos.Line)
		assert.Equal(t, 1, err.pos.Column)
		assert.Equal(t, `Unterminated string`, err.msg)
	}

	// scan, err = scanOnce(`'bad \z esc'`)
	// assert.Equal(t, token.INVALID, scan.tok)
	// if assert.NotNil(t, err) {
	// 	assert.Equal(t, 6, err.pos.Offset)
	// 	assert.Equal(t, 1, err.pos.Line)
	// 	assert.Equal(t, 7, err.pos.Column)
	// 	assert.Equal(t, `Unexpected character escape sequence: \z`, err.msg)
	// }

	// scan, err = scanOnce(`'bad \x esc'`)
	// assert.Equal(t, token.INVALID, scan.tok)
	// if assert.NotNil(t, err) {
	// 	assert.Equal(t, 6, err.pos.Offset)
	// 	assert.Equal(t, 1, err.pos.Line)
	// 	assert.Equal(t, 7, err.pos.Column)
	// 	assert.Equal(t, `Unexpected character escape sequence: \x`, err.msg)
	// }

	// scan, err = scanOnce(`'bad \u1 esc'`)
	// assert.Equal(t, token.INVALID, scan.tok)
	// if assert.NotNil(t, err) {
	// 	assert.Equal(t, 6, err.pos.Offset)
	// 	assert.Equal(t, 1, err.pos.Line)
	// 	assert.Equal(t, 7, err.pos.Column)
	// 	assert.Equal(t, `Unexpected character in escape sequence: U+0020 ' '`, err.msg)
	// }

	// scan, err = scanOnce(`'bad \u0XX1 esc'`)
	// assert.Equal(t, token.INVALID, scan.tok)
	// if assert.NotNil(t, err) {
	// 	assert.Equal(t, 6, err.pos.Offset)
	// 	assert.Equal(t, 1, err.pos.Line)
	// 	assert.Equal(t, 7, err.pos.Column)
	// 	assert.Equal(t, `Unexpected character in escape sequence: U+0058 'X'`, err.msg)
	// }

	// scan, err = scanOnce(`'bad \uXXXX esc'`)
	// assert.Equal(t, token.INVALID, scan.tok)
	// if assert.NotNil(t, err) {
	// 	assert.Equal(t, 6, err.pos.Offset)
	// 	assert.Equal(t, 1, err.pos.Line)
	// 	assert.Equal(t, 7, err.pos.Column)
	// 	assert.Equal(t, `Unexpected character in escape sequence: U+0058 'X'`, err.msg)
	// }

	// scan, err = scanOnce(`'bad \uFXXX esc'`)
	// assert.Equal(t, token.INVALID, scan.tok)
	// if assert.NotNil(t, err) {
	// 	assert.Equal(t, 6, err.pos.Offset)
	// 	assert.Equal(t, 1, err.pos.Line)
	// 	assert.Equal(t, 7, err.pos.Column)
	// 	assert.Equal(t, `Unexpected character in escape sequence: U+0058 'X'`, err.msg)
	// }

	// scan, err = scanOnce(`'bad \uXXXF esc'`)
	// assert.Equal(t, token.INVALID, scan.tok)
	// if assert.NotNil(t, err) {
	// 	assert.Equal(t, 6, err.pos.Offset)
	// 	assert.Equal(t, 1, err.pos.Line)
	// 	assert.Equal(t, 7, err.pos.Column)
	// 	assert.Equal(t, `Unexpected character in escape sequence: U+0058 'X'`, err.msg)
	// }
}

func TestScansNumbers(t *testing.T) {
	scan, err := scanOnce("4")
	assert.Nil(t, err)
	assert.Equal(t, token.NUMBER, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "4", scan.lit)

	scan, err = scanOnce("4.123")
	assert.Nil(t, err)
	assert.Equal(t, token.NUMBER, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "4.123", scan.lit)

	scan, err = scanOnce(".4")
	assert.Nil(t, err)
	assert.Equal(t, token.NUMBER, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, ".4", scan.lit)

	scan, err = scanOnce(".123")
	assert.Nil(t, err)
	assert.Equal(t, token.NUMBER, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, ".123", scan.lit)

	scan, err = scanOnce("9")
	assert.Nil(t, err)
	assert.Equal(t, token.NUMBER, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "9", scan.lit)

	scan, err = scanOnce("0")
	assert.Nil(t, err)
	assert.Equal(t, token.NUMBER, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "0", scan.lit)

	scan, err = scanOnce("0.123")
	assert.Nil(t, err)
	assert.Equal(t, token.NUMBER, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "0.123", scan.lit)

	scan, err = scanOnce("123e4")
	assert.Nil(t, err)
	assert.Equal(t, token.NUMBER, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "123e4", scan.lit)

	scan, err = scanOnce("123e-4")
	assert.Nil(t, err)
	assert.Equal(t, token.NUMBER, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "123e-4", scan.lit)

	scan, err = scanOnce("123e+4")
	assert.Nil(t, err)
	assert.Equal(t, token.NUMBER, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "123e+4", scan.lit)

	scan, err = scanOnce(".123e4")
	assert.Nil(t, err)
	assert.Equal(t, token.NUMBER, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, ".123e4", scan.lit)

	scan, err = scanOnce(".123e-4")
	assert.Nil(t, err)
	assert.Equal(t, token.NUMBER, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, ".123e-4", scan.lit)

	scan, err = scanOnce(".123e+4")
	assert.Nil(t, err)
	assert.Equal(t, token.NUMBER, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, ".123e+4", scan.lit)

	scan, err = scanOnce(".123e4567")
	assert.Nil(t, err)
	assert.Equal(t, token.NUMBER, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, ".123e4567", scan.lit)
}

func TestReportsUsefulNumberErrors(t *testing.T) {
	scan, err := scanOnce("1.")
	assert.Equal(t, token.INVALID, scan.tok)
	if assert.NotNil(t, err) {
		assert.Equal(t, 0, err.pos.Offset)
		assert.Equal(t, 1, err.pos.Line)
		assert.Equal(t, 1, err.pos.Column)
		assert.Equal(t, `Missing digits after decimal point in number`, err.msg)
	}

	scan, err = scanOnce("1.A")
	assert.Equal(t, token.INVALID, scan.tok)
	if assert.NotNil(t, err) {
		assert.Equal(t, 0, err.pos.Offset)
		assert.Equal(t, 1, err.pos.Line)
		assert.Equal(t, 1, err.pos.Column)
		assert.Equal(t, `Missing digits after decimal point in number`, err.msg)
	}

	scan, err = scanOnce("1.0e")
	assert.Equal(t, token.INVALID, scan.tok)
	if assert.NotNil(t, err) {
		assert.Equal(t, 0, err.pos.Offset)
		assert.Equal(t, 1, err.pos.Line)
		assert.Equal(t, 1, err.pos.Column)
		assert.Equal(t, `Missing digits after exponent in number`, err.msg)
	}

	scan, err = scanOnce("1.0eA")
	assert.Equal(t, token.INVALID, scan.tok)
	if assert.NotNil(t, err) {
		assert.Equal(t, 0, err.pos.Offset)
		assert.Equal(t, 1, err.pos.Line)
		assert.Equal(t, 1, err.pos.Column)
		assert.Equal(t, `Missing digits after exponent in number`, err.msg)
	}
}

func TestScansPunctuation(t *testing.T) {
	scan, err := scanOnce("$")
	assert.Nil(t, err)
	assert.Equal(t, token.DOLLAR, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "", scan.lit)

	scan, err = scanOnce("(")
	assert.Nil(t, err)
	assert.Equal(t, token.LEFT_PAREN, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "", scan.lit)

	scan, err = scanOnce(")")
	assert.Nil(t, err)
	assert.Equal(t, token.RIGHT_PAREN, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "", scan.lit)

	scan, err = scanOnce(";")
	assert.Nil(t, err)
	assert.Equal(t, token.SEMICOLON, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "", scan.lit)

	scan, err = scanOnce(":")
	assert.Nil(t, err)
	assert.Equal(t, token.COLON, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "", scan.lit)

	scan, err = scanOnce("=")
	assert.Nil(t, err)
	assert.Equal(t, token.EQUALS, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "", scan.lit)

	scan, err = scanOnce("@")
	assert.Nil(t, err)
	assert.Equal(t, token.AT, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "", scan.lit)

	scan, err = scanOnce("+")
	assert.Nil(t, err)
	assert.Equal(t, token.PLUS, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "", scan.lit)

	scan, err = scanOnce("-")
	assert.Nil(t, err)
	assert.Equal(t, token.MINUS, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "", scan.lit)

	scan, err = scanOnce("/")
	assert.Nil(t, err)
	assert.Equal(t, token.SLASH, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "", scan.lit)

	scan, err = scanOnce(",")
	assert.Nil(t, err)
	assert.Equal(t, token.COMMA, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "", scan.lit)

	scan, err = scanOnce(".")
	assert.Nil(t, err)
	assert.Equal(t, token.PERIOD, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "", scan.lit)

	scan, err = scanOnce("*")
	assert.Nil(t, err)
	assert.Equal(t, token.ASTERISK, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "", scan.lit)

	scan, err = scanOnce("[")
	assert.Nil(t, err)
	assert.Equal(t, token.LEFT_BRACKET, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "", scan.lit)

	scan, err = scanOnce("]")
	assert.Nil(t, err)
	assert.Equal(t, token.RIGHT_BRACKET, scan.tok)
	assert.Equal(t, 0, scan.pos)
	assert.Equal(t, "", scan.lit)
}

func TestReportsUsefulUnknownCharacter(t *testing.T) {
	scan, err := scanOnce("\u203B")
	assert.Equal(t, token.INVALID, scan.tok)
	if assert.NotNil(t, err) {
		assert.Equal(t, 0, err.pos.Offset)
		assert.Equal(t, 1, err.pos.Line)
		assert.Equal(t, 1, err.pos.Column)
		assert.Equal(t, "Unexpected character U+203B '\u203B'", err.msg)
	}

	scan, err = scanOnce("\u200b")
	assert.Equal(t, token.INVALID, scan.tok)
	if assert.NotNil(t, err) {
		assert.Equal(t, 0, err.pos.Offset)
		assert.Equal(t, 1, err.pos.Line)
		assert.Equal(t, 1, err.pos.Column)
		assert.Equal(t, `Unexpected character U+200B`, err.msg)
	}
}

func TestScannerNextCharacter(t *testing.T) {
	var err *scanError

	err = scanAll("SELECT * FROM candies\r\n  WHERE sweetness = 11\n\r\r")
	assert.Nil(t, err)

	err = scanAll(string([]byte{0x00, 0xFF}))
	if assert.NotNil(t, err) {
		assert.Equal(t, 0, err.pos.Offset)
		assert.Equal(t, 1, err.pos.Line)
		assert.Equal(t, 1, err.pos.Column)
		assert.Equal(t, `Unexpected character U+0000`, err.msg)
	}
}

func TestScanPos(t *testing.T) {
	var err *scanError
	handleError := func(pos token.Position, msg string) {
		err = &scanError{pos, msg}
	}

	var scan scanToken
	s := Scanner{}
	s.Init([]byte("CREATE TABLE\n  candies\n()"), handleError, ScanRuleset{})
	assert.Equal(t, token.Position{"", 0, 1, 1}, s.Pos())
	assert.Nil(t, err)

	scan.pos, scan.tok, scan.lit = s.Scan()
	assert.Equal(t, token.Position{"", 6, 1, 7}, s.Pos())
	assert.Nil(t, err)

	scan.pos, scan.tok, scan.lit = s.Scan()
	assert.Equal(t, token.Position{"", 12, 1, 13}, s.Pos())
	assert.Nil(t, err)

	scan.pos, scan.tok, scan.lit = s.Scan()
	assert.Equal(t, token.Position{"", 22, 2, 10}, s.Pos())
	assert.Nil(t, err)

	scan.pos, scan.tok, scan.lit = s.Scan()
	assert.Equal(t, token.Position{"", 24, 3, 2}, s.Pos())
	assert.Nil(t, err)

	scan.pos, scan.tok, scan.lit = s.Scan()
	assert.Equal(t, token.Position{"", 25, 3, 3}, s.Pos())
	assert.Nil(t, err)
}
