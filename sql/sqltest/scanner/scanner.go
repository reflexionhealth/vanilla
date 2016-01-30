package scanner

import (
	"fmt"
	"unicode/utf8"

	"github.com/reflexionhealth/vanilla/sql/sqltest/token"
)

// isLetter returns true if the rune matches [A-Za-z_]
func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

// An ErrorHandler may be provided to Scanner.Init. If a syntax error is
// encountered and a handler was installed, the handler is called with a
// position and an error message. The position points to the beginning of
// the offending token.
type ErrorHandler func(pos token.Position, msg string)

// A ScanRuleset specifies the dialect specific tokenizing rules for a SQL dialect
type ScanRuleset struct {
	BracketsAreQuotes         bool
	BacktickIsQuotemark       bool
	DoubleQuoteIsNotQuotemark bool

	DollarIsLetter bool

	// CStyleComment bool
	// CStyleEscapeSeq bool
}

// A Scanner holds the scanner's internal state.
type Scanner struct {
	// immutable state
	src   []byte
	err   ErrorHandler
	rules ScanRuleset

	// scanning state
	char       rune // current character
	offset     int  // byte offset to current char
	readOffset int  // reading offset (position after current character)
	lineOffset int  // current line offset
	line       int  // current line

	// public state
	ErrorCount int // number of errors encountered
}

// Init prepares the scanner s to tokenize the text src by setting the
// scanner at the beginning of src.
//
// Calls to Scan will invoke the error handler err if they encounter a
// syntax error and err is not nil. Also, for each error encountered,
// the Scanner field ErrorCount is incremented by one.
//
// Note that Init may call err if there is an error in the first character
// of the file.
func (s *Scanner) Init(src []byte, err ErrorHandler, rules ScanRuleset) {
	s.src = src
	s.err = err
	s.rules = rules

	s.char = ' '
	s.offset = 0
	s.readOffset = 0
	s.lineOffset = 0
	s.line = 0

	s.next()
}

// Scan scans the next token and returns the token position, the token, and its
// literal string if applicable. The source end is indicated by the EOL token.
//
// If the returned token is a literal the literal string has the corresponding value.
//
// If the returned token is a keyword, the literal string is the keyword.
//
// If the returned token is an identifier, the literal string is the identifier.
//
// If the returned token is a quoted identifier, the literal string is
// the identifier without the quotes.
//
// If the returned token is invalid, the literal string is the offending character.
//
// In all other cases, Scan returns an empty literal string.
func (s *Scanner) Scan() (pos int, tok token.Token, lit string) {
	// scanAgain:
	s.skipWhitespace()

	pos = s.offset
	ch := s.char
	switch {
	case isLetter(ch):
		lit = s.scanIdentifier()
		tok = token.IDENT
		if len(lit) > 1 {
			// keywords are longer than one letter - avoid lookup otherwise
			tok = token.Lookup(lit)
		}
	case isDigit(ch):
		tok, lit = s.scanNumber(false)
	default:
		s.next() // always make progress
		switch ch {
		case -1:
			tok = token.EOL
		// case ???:
		// 	s.scanComment()
		// 	goto scanAgain
		case '"':
			if s.rules.DoubleQuoteIsNotQuotemark {
				s.error(pos, fmt.Sprintf("Unexpected character %#U", ch))
				tok = token.INVALID
				lit = string(ch)
			} else {
				tok, lit = s.scanQuotedIdentifier('"')
			}
		case '`':
			if s.rules.BacktickIsQuotemark {
			} else {
				s.error(pos, fmt.Sprintf("Unexpected character %#U", ch))
				tok = token.INVALID
				lit = string(ch)
			}
		case '\'':
			tok, lit = s.scanString()
		case ';':
			tok = token.SEMICOLON
		case ':':
			tok = token.COLON
		case '$':
			tok = token.DOLLAR
		case '*':
			tok = token.ASTERISK
		case '+':
			tok = token.PLUS
		case '-':
			tok = token.MINUS
		case '/':
			tok = token.SLASH
		case ',':
			tok = token.COMMA
		case '=':
			tok = token.EQUALS
		case '@':
			tok = token.AT
		case '(':
			tok = token.LEFT_PAREN
		case '[':
			if s.rules.BracketsAreQuotes {
				tok, lit = s.scanQuotedIdentifier(']')
			} else {
				tok = token.LEFT_BRACKET
			}
		case ')':
			tok = token.RIGHT_PAREN
		case ']':
			tok = token.RIGHT_BRACKET
		case '.':
			if isDigit(s.char) {
				tok, lit = s.scanNumber(true)
			} else {
				tok = token.PERIOD
			}
		default:
			s.error(pos, fmt.Sprintf("Unexpected character %#U", ch))
			tok = token.INVALID
			lit = string(ch)
		}
	}

	return
}

func (s *Scanner) Pos() token.Position {
	// Get length of current line in UTF-8 characters
	column := 1 + len(string(s.src[s.lineOffset:s.offset]))
	return token.Position{
		Name:   "sql",
		Offset: s.offset,
		Line:   s.line + 1,
		Column: column,
	}
}

func (s *Scanner) error(offset int, msg string) {
	s.ErrorCount++

	if s.err != nil {
		column := 1 + len(string(s.src[s.lineOffset:offset]))
		pos := token.Position{
			Name:   "sql",
			Offset: offset,
			Line:   s.line + 1,
			Column: column,
		}

		s.err(pos, msg)
	}
}

func (s *Scanner) next() {
	if s.readOffset < len(s.src) {
		s.offset = s.readOffset

		wasCarriageReturn := false
		if s.char == '\n' {
			s.line += 1
			s.lineOffset = s.offset
		} else if s.char == '\r' {
			s.line += 1
			s.lineOffset = s.offset
			wasCarriageReturn = true
		}

		r, width := rune(s.src[s.readOffset]), 1
		switch {
		case r == 0:
			s.error(s.offset, "Unexpected character NUL")
		case r >= 0x80:
			// not ASCII
			r, width = utf8.DecodeRune(s.src[s.readOffset:])
			if r == utf8.RuneError && width == 1 {
				s.error(s.offset, "Invalid UTF-8 encoding")
			}
		}
		s.readOffset += width
		s.char = r

		if s.char == '\n' && wasCarriageReturn {
			s.line -= 1
		}
	} else {
		s.offset = len(s.src)
		if s.char == '\n' || s.char == '\r' {
			s.lineOffset = s.offset
		}
		s.char = -1 // eof
	}
}

func (s *Scanner) skipWhitespace() {
	for s.char == ' ' || s.char == '\t' || s.char == '\n' || s.char == '\r' {
		s.next()
	}
}

func (s *Scanner) scanIdentifier() string {
	offset := s.offset
	for isLetter(s.char) || isDigit(s.char) || (s.char == '$' && s.rules.DollarIsLetter) {
		s.next()
	}

	return string(s.src[offset:s.offset])
}

func (s *Scanner) scanQuotedIdentifier(closemark rune) (token.Token, string) {
	// opening quotemark already consumed
	offset := s.offset - 1
	tok := token.QUOTED_IDENT
	lit := s.scanIdentifier()

	if s.char == closemark {
		s.next()
	} else if s.char == ' ' {
		tok = token.INVALID
		lit = string(s.src[offset:s.offset])
		s.error(offset, "Unterminated identifier")
	} else {
		tok = token.INVALID
		lit = string(s.src[offset:s.offset])
		s.error(offset, fmt.Sprintf("Unexpected character in identifier: %#U", s.char))
	}

	return tok, lit
}

func (s *Scanner) scanMantissa() {
	for isDigit(s.char) {
		s.next()
	}
}

func (s *Scanner) scanNumber(afterDecimal bool) (token.Token, string) {
	tok := token.NUMBER
	offset := s.offset
	if afterDecimal {
		offset -= 1
	}

	s.scanMantissa()
	if s.char == '.' && !afterDecimal { // TODO: maybe an error?
		s.next()
		decOffset := s.offset
		s.scanMantissa()
		if s.offset == decOffset {
			s.error(offset, "Missing digits after decimal point in number")
			tok = token.INVALID
		}
	}
	if s.char == 'e' {
		s.next()
		if s.char == '+' || s.char == '-' {
			s.next()
		}
		expOffset := s.offset
		s.scanMantissa()
		if s.offset == expOffset {
			s.error(offset, "Missing digits after exponent in number")
			tok = token.INVALID
		}
	}

	return tok, string(s.src[offset:s.offset])
}

func (s *Scanner) scanString() (token.Token, string) {
	// opening single-quote already consumed
	offset := s.offset - 1
	tok := token.STRING

	for {
		ch := s.char
		if ch == '\n' || ch == '\r' || ch < 0 {
			tok = token.INVALID
			s.error(offset, "Unterminated string")
			break
		} else if ch == '\\' {
			s.next()
		}

		s.next()
		if ch == '\'' {
			break
		}
	}

	return tok, string(s.src[offset:s.offset])
}
