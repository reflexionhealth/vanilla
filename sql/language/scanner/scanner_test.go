package scanner

import (
	"testing"

	"github.com/reflexionhealth/vanilla/expect"
	"github.com/reflexionhealth/vanilla/sql/language/token"
)

func TestSelect(t *testing.T) {
	query := `SELECT * FROM users WHERE id = 3`

	failOnError := func(pos token.Position, msg string) {
		t.Errorf("At Line %d, Col %d: %s", pos.Line, pos.Column, msg)
	}

	s := Scanner{}
	s.Init([]byte(query), failOnError, Ruleset{})

	var tokens []token.Token
	MAX_ITER := 200 // Don't loop forever
	for i := 0; i < MAX_ITER; i++ {
		_, tok, _ := s.Scan()
		tokens = append(tokens, tok)
		if tok == token.INVALID || tok == token.EOS {
			break
		}
	}

	expect.Equal(t, s.ErrorCount, 0)
	expect.Equal(t, tokens, []token.Token{
		// SELECT * FROM users
		token.SELECT, token.ASTERISK, token.FROM, token.IDENT,
		// WHERE id = 3
		token.WHERE, token.IDENT, token.EQUALS, token.NUMBER,
		token.EOS,
	})
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
	s.Init([]byte(src), handleError, Ruleset{})
	t.pos, t.tok, t.lit = s.Scan()

	return t, err
}

func scanOnceWith(src string, rules Ruleset) (scanToken, *scanError) {
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
	s.Init([]byte(src), handleError, Ruleset{})

	for i := 0; i < 9999; i++ {
		t.pos, t.tok, t.lit = s.Scan()
		if t.tok == token.EOS {
			break
		} else if err != nil {
			return err
		}
	}

	return nil
}

func TestSkipsWhitesace(t *testing.T) {
	scan, err := scanOnce("\n    SELECT\n")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.SELECT)
	expect.Equal(t, scan.pos, 5)
	expect.Equal(t, scan.lit, "SELECT")

	// scan, err = scanOnce("\n    --comment\n    SELECT--comment\n")
	// expect.Nil(t, err)
	// expect.Equal(t, scan.tok, token.SELECT)
	// expect.Equal(t, scan.pos, 18)
	// expect.Equal(t, scan.lit, "SELECT")
	//
	// scan, err = scanOnce("\n    --comment\r\n    SELECT--comment\n")
	// expect.Nil(t, err)
	// expect.Equal(t, scan.tok, token.SELECT)
	// expect.Equal(t, scan.pos, 19)
	// expect.Equal(t, scan.lit, "SELECT")
}

func TestErrorsRespectWhitespace(t *testing.T) {
	scan, err := scanOnce("\n\n    ~\n")
	expect.Equal(t, token.INVALID, scan.tok)
	if expect.NotNil(t, err) {
		expect.Equal(t, err.pos.Offset, 6)
		expect.Equal(t, err.pos.Line, 3)
		expect.Equal(t, err.pos.Column, 5)
		expect.Equal(t, err.msg, `unexpected character U+007E '~'`)
	}
}

func TestScansIdentifier(t *testing.T) {
	scan, err := scanOnce(`simple`)
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.IDENT)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, `simple`)

	scan, err = scanOnce("sim$ple")
	expect.Equal(t, scan.tok, token.IDENT)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, `sim`)

	scan, err = scanOnceWith("sim$ple", Ruleset{DollarIsLetter: true})
	expect.Equal(t, scan.tok, token.IDENT)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, `sim$ple`)
}

func TestScansQuotedIdentifier(t *testing.T) {
	scan, err := scanOnce(`"simple"`)
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.QUOTED_IDENT)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, `simple`)

	scan, err = scanOnceWith("`simple`", Ruleset{BacktickIsQuotemark: true})
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.QUOTED_IDENT)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, `simple`)
}

func TestScansStrings(t *testing.T) {
	scan, err := scanOnce(`'simple'`)
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.STRING)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, `'simple'`)

	scan, err = scanOnce(`' white space '`)
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.STRING)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, `' white space '`)

	scan, err = scanOnce(`'quote\''`)
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.STRING)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, `'quote\''`)

	scan, err = scanOnce(`'escaped \n\r\b\t\f'`)
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.STRING)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, `'escaped \n\r\b\t\f'`)

	scan, err = scanOnce(`'slashes \\ \/'`)
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.STRING)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, `'slashes \\ \/'`)

	scan, err = scanOnceWith(`"simple"`, Ruleset{DoubleQuoteIsString: true})
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.STRING)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, `"simple"`)
}

func TestReportsUsefulStringErrors(t *testing.T) {
	scan, err := scanOnce(`'`)
	expect.Equal(t, scan.tok, token.INVALID)
	if expect.NotNil(t, err) {
		expect.Equal(t, err.pos.Offset, 0)
		expect.Equal(t, err.pos.Line, 1)
		expect.Equal(t, err.pos.Column, 1)
		expect.Equal(t, err.msg, `unterminated string`)
	}

	scan, err = scanOnce(`'No end quote`)
	expect.Equal(t, scan.tok, token.INVALID)
	if expect.NotNil(t, err) {
		expect.Equal(t, err.pos.Offset, 0)
		expect.Equal(t, err.pos.Line, 1)
		expect.Equal(t, err.pos.Column, 1)
		expect.Equal(t, err.msg, `unterminated string`)
	}

	// scan, err = scanOnce("'contains unescaped \u0007 control char'")
	// expect.Equal(t, scan.tok, token.INVALID)
	// if expect.NotNil(t, err) {
	// 	expect.Equal(t, err.pos.Offset, 0)
	// 	expect.Equal(t, err.pos.Line, 1)
	// 	expect.Equal(t, err.pos.Column, 1)
	// 	expect.Equal(t, err.msg, `unexpected character in string: U+0007`)
	// }

	// scan, err = scanOnce("'null-byte \u0000 in string'")
	// expect.Equal(t, scan.tok, token.INVALID)
	// if expect.NotNil(t, err) {
	// 	expect.Equal(t, err.pos.Offset, 0)
	// 	expect.Equal(t, err.pos.Line, 1)
	// 	expect.Equal(t, err.pos.Column, 1)
	// 	expect.Equal(t, err.msg, `unexpected character in string: U+0000`)
	// }

	// scan, err = scanOnce(`'\u`)
	// expect.Equal(t, scan.tok, token.INVALID)
	// if expect.NotNil(t, err) {
	// 	expect.Equal(t, err.pos.Offset, 2)
	// 	expect.Equal(t, err.pos.Line, 1)
	// 	expect.Equal(t, err.pos.Column, 3)
	// 	expect.Equal(t, err.msg, `unterminated escape sequence`)
	// }

	// scan, err = scanOnce(`'\`)
	// expect.Equal(t, scan.tok, token.INVALID)
	// if expect.NotNil(t, err) {
	// 	expect.Equal(t, err.pos.Offset, 2)
	// 	expect.Equal(t, err.pos.Line, 1)
	// 	expect.Equal(t, err.pos.Column, 3)
	// 	expect.Equal(t, err.msg, `unterminated escape sequence`)
	// }

	// scan, err = scanOnce(`'\m'`)
	// expect.Equal(t, scan.tok, token.INVALID)
	// if expect.NotNil(t, err) {
	// 	expect.Equal(t, err.pos.Offset, 2)
	// 	expect.Equal(t, err.pos.Line, 1)
	// 	expect.Equal(t, err.pos.Column, 3)
	// 	expect.Equal(t, err.msg, `unknown escape sequence`)
	// }

	// scan, err = scanOnce(`'\uD800'`)
	// expect.Equal(t, scan.tok, token.INVALID)
	// if expect.NotNil(t, err) {
	// 	expect.Equal(t, err.pos.Offset, 2)
	// 	expect.Equal(t, err.pos.Line, 1)
	// 	expect.Equal(t, err.pos.Column, 3)
	// 	expect.Equal(t, err.msg, `escape sequence is invalid unicode code point`)
	// }

	scan, err = scanOnce("'multi\nline'")
	expect.Equal(t, scan.tok, token.INVALID)
	if expect.NotNil(t, err) {
		expect.Equal(t, err.pos.Offset, 0)
		expect.Equal(t, err.pos.Line, 1)
		expect.Equal(t, err.pos.Column, 1)
		expect.Equal(t, err.msg, `unterminated string`)
	}

	scan, err = scanOnce("'multi\rline'")
	expect.Equal(t, scan.tok, token.INVALID)
	if expect.NotNil(t, err) {
		expect.Equal(t, err.pos.Offset, 0)
		expect.Equal(t, err.pos.Line, 1)
		expect.Equal(t, err.pos.Column, 1)
		expect.Equal(t, err.msg, `unterminated string`)
	}

	// scan, err = scanOnce(`'bad \z esc'`)
	// expect.Equal(t, scan.tok, token.INVALID)
	// if expect.NotNil(t, err) {
	// 	expect.Equal(t, err.pos.Offset, 6)
	// 	expect.Equal(t, err.pos.Line, 1)
	// 	expect.Equal(t, err.pos.Column, 7)
	// 	expect.Equal(t, err.msg, `unexpected character escape sequence: \z`)
	// }

	// scan, err = scanOnce(`'bad \x esc'`)
	// expect.Equal(t, scan.tok, token.INVALID)
	// if expect.NotNil(t, err) {
	// 	expect.Equal(t, err.pos.Offset, 6)
	// 	expect.Equal(t, err.pos.Line, 1)
	// 	expect.Equal(t, err.pos.Column, 7)
	// 	expect.Equal(t, err.msg, `unexpected character escape sequence: \x`)
	// }

	// scan, err = scanOnce(`'bad \u1 esc'`)
	// expect.Equal(t, scan.tok, token.INVALID)
	// if expect.NotNil(t, err) {
	// 	expect.Equal(t, err.pos.Offset, 6)
	// 	expect.Equal(t, err.pos.Line, 1)
	// 	expect.Equal(t, err.pos.Column, 7)
	// 	expect.Equal(t, err.msg, `unexpected character in escape sequence: U+0020 ' '`)
	// }

	// scan, err = scanOnce(`'bad \u0XX1 esc'`)
	// expect.Equal(t, scan.tok, token.INVALID)
	// if expect.NotNil(t, err) {
	// 	expect.Equal(t, err.pos.Offset, 6)
	// 	expect.Equal(t, err.pos.Line, 1)
	// 	expect.Equal(t, err.pos.Column, 7)
	// 	expect.Equal(t, err.msg, `unexpected character in escape sequence: U+0058 'X'`)
	// }

	// scan, err = scanOnce(`'bad \uXXXX esc'`)
	// expect.Equal(t, scan.tok, token.INVALID)
	// if expect.NotNil(t, err) {
	// 	expect.Equal(t, err.pos.Offset, 6)
	// 	expect.Equal(t, err.pos.Line, 1)
	// 	expect.Equal(t, err.pos.Column, 7)
	// 	expect.Equal(t, err.msg, `unexpected character in escape sequence: U+0058 'X'`)
	// }

	// scan, err = scanOnce(`'bad \uFXXX esc'`)
	// expect.Equal(t, scan.tok, token.INVALID)
	// if expect.NotNil(t, err) {
	// 	expect.Equal(t, err.pos.Offset, 6)
	// 	expect.Equal(t, err.pos.Line, 1)
	// 	expect.Equal(t, err.pos.Column, 7)
	// 	expect.Equal(t, err.msg, `unexpected character in escape sequence: U+0058 'X'`)
	// }

	// scan, err = scanOnce(`'bad \uXXXF esc'`)
	// expect.Equal(t, scan.tok, token.INVALID)
	// if expect.NotNil(t, err) {
	// 	expect.Equal(t, err.pos.Offset, 6)
	// 	expect.Equal(t, err.pos.Line, 1)
	// 	expect.Equal(t, err.pos.Column, 7)
	// 	expect.Equal(t, err.msg, `unexpected character in escape sequence: U+0058 'X'`)
	// }
}

func TestScansNumbers(t *testing.T) {
	scan, err := scanOnce("4")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.NUMBER)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "4")

	scan, err = scanOnce("4.123")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.NUMBER)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "4.123")

	scan, err = scanOnce(".4")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.NUMBER)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, ".4")

	scan, err = scanOnce(".123")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.NUMBER)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, ".123")

	scan, err = scanOnce("9")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.NUMBER)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "9")

	scan, err = scanOnce("0")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.NUMBER)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "0")

	scan, err = scanOnce("0.123")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.NUMBER)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "0.123")

	scan, err = scanOnce("123e4")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.NUMBER)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "123e4")

	scan, err = scanOnce("123e-4")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.NUMBER)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "123e-4")

	scan, err = scanOnce("123e+4")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.NUMBER)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "123e+4")

	scan, err = scanOnce(".123e4")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.NUMBER)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, ".123e4")

	scan, err = scanOnce(".123e-4")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.NUMBER)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, ".123e-4")

	scan, err = scanOnce(".123e+4")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.NUMBER)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, ".123e+4")

	scan, err = scanOnce(".123e4567")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.NUMBER)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, ".123e4567")
}

func TestReportsUsefulNumberErrors(t *testing.T) {
	scan, err := scanOnce("1.")
	expect.Equal(t, scan.tok, token.INVALID)
	if expect.NotNil(t, err) {
		expect.Equal(t, err.pos.Offset, 0)
		expect.Equal(t, err.pos.Line, 1)
		expect.Equal(t, err.pos.Column, 1)
		expect.Equal(t, err.msg, `missing digits after decimal point in number`)
	}

	scan, err = scanOnce("1.A")
	expect.Equal(t, scan.tok, token.INVALID)
	if expect.NotNil(t, err) {
		expect.Equal(t, err.pos.Offset, 0)
		expect.Equal(t, err.pos.Line, 1)
		expect.Equal(t, err.pos.Column, 1)
		expect.Equal(t, err.msg, `missing digits after decimal point in number`)
	}

	scan, err = scanOnce("1.0e")
	expect.Equal(t, scan.tok, token.INVALID)
	if expect.NotNil(t, err) {
		expect.Equal(t, err.pos.Offset, 0)
		expect.Equal(t, err.pos.Line, 1)
		expect.Equal(t, err.pos.Column, 1)
		expect.Equal(t, err.msg, `missing digits after exponent in number`)
	}

	scan, err = scanOnce("1.0eA")
	expect.Equal(t, scan.tok, token.INVALID)
	if expect.NotNil(t, err) {
		expect.Equal(t, err.pos.Offset, 0)
		expect.Equal(t, err.pos.Line, 1)
		expect.Equal(t, err.pos.Column, 1)
		expect.Equal(t, err.msg, `missing digits after exponent in number`)
	}
}

func TestScansPunctuation(t *testing.T) {
	scan, err := scanOnce("$")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.DOLLAR)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")

	scan, err = scanOnce("(")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.LEFT_PAREN)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")

	scan, err = scanOnce(")")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.RIGHT_PAREN)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")

	scan, err = scanOnce(";")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.SEMICOLON)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")

	scan, err = scanOnce(":")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.COLON)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")

	scan, err = scanOnce("=")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.EQUALS)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")

	scan, err = scanOnce("@")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.AT)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")

	scan, err = scanOnce("+")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.PLUS)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")

	scan, err = scanOnce("-")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.MINUS)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")

	scan, err = scanOnce("/")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.SLASH)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")

	scan, err = scanOnce(",")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.COMMA)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")

	scan, err = scanOnce(".")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.PERIOD)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")

	scan, err = scanOnce("*")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.ASTERISK)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")

	scan, err = scanOnce("?")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.QUESTION)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")

	scan, err = scanOnce("[")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.LEFT_BRACKET)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")

	scan, err = scanOnce("]")
	expect.Nil(t, err)
	expect.Equal(t, scan.tok, token.RIGHT_BRACKET)
	expect.Equal(t, scan.pos, 0)
	expect.Equal(t, scan.lit, "")
}

func TestReportsUsefulunknownCharacter(t *testing.T) {
	scan, err := scanOnce("\u203B")
	expect.Equal(t, scan.tok, token.INVALID)
	if expect.NotNil(t, err) {
		expect.Equal(t, err.pos.Offset, 0)
		expect.Equal(t, err.pos.Line, 1)
		expect.Equal(t, err.pos.Column, 1)
		expect.Equal(t, err.msg, "unexpected character U+203B '\u203B'")
	}

	scan, err = scanOnce("\u200b")
	expect.Equal(t, scan.tok, token.INVALID)
	if expect.NotNil(t, err) {
		expect.Equal(t, err.pos.Offset, 0)
		expect.Equal(t, err.pos.Line, 1)
		expect.Equal(t, err.pos.Column, 1)
		expect.Equal(t, err.msg, `unexpected character U+200B`)
	}
}

func TestScannerNextCharacter(t *testing.T) {
	var err *scanError

	err = scanAll("SELECT * FROM candies\r\n  WHERE sweetness = 11\n\r\r")
	expect.Nil(t, err)

	err = scanAll(string([]byte{0x00, 0xFF}))
	if expect.NotNil(t, err) {
		expect.Equal(t, err.pos.Offset, 0)
		expect.Equal(t, err.pos.Line, 1)
		expect.Equal(t, err.pos.Column, 1)
		expect.Equal(t, err.msg, `unexpected character U+0000`)
	}
}

func TestScanPos(t *testing.T) {
	var err *scanError
	handleError := func(pos token.Position, msg string) {
		err = &scanError{pos, msg}
	}

	var scan scanToken
	s := Scanner{}
	s.Init([]byte("CREATE TABLE\n  candies\n()"), handleError, Ruleset{})
	expect.Equal(t, s.Pos(), token.Position{"sql", 0, 1, 1})
	expect.Nil(t, err)

	scan.pos, scan.tok, scan.lit = s.Scan()
	expect.Equal(t, s.Pos(), token.Position{"sql", 6, 1, 7})
	expect.Nil(t, err)

	scan.pos, scan.tok, scan.lit = s.Scan()
	expect.Equal(t, s.Pos(), token.Position{"sql", 12, 1, 13})
	expect.Nil(t, err)

	scan.pos, scan.tok, scan.lit = s.Scan()
	expect.Equal(t, s.Pos(), token.Position{"sql", 22, 2, 10})
	expect.Nil(t, err)

	scan.pos, scan.tok, scan.lit = s.Scan()
	expect.Equal(t, s.Pos(), token.Position{"sql", 24, 3, 2})
	expect.Nil(t, err)

	scan.pos, scan.tok, scan.lit = s.Scan()
	expect.Equal(t, s.Pos(), token.Position{"sql", 25, 3, 3})
	expect.Nil(t, err)
}
