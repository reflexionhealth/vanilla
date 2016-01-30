package token

import (
	"fmt"
	"strconv"
	"strings"
)

// Position describes an arbitrary source position including the name, line,
// and column location. A Position is valid if the line number is > 0.
type Position struct {
	Name   string // source name, if any
	Offset int    // offset, starting at 0
	Line   int    // line number, starting at 1
	Column int    // column number, starting at 1
}

// IsValid reports whether the position is valid.
func (pos *Position) IsValid() bool { return pos.Line > 0 }

// String returns a string in one of several forms:
//
//	name:line:column    valid position with name
//	line:column         valid position without name
//	name                invalid position with name
//	-                   invalid position without name
//
func (pos Position) String() string {
	s := pos.Name
	if pos.IsValid() {
		if s != "" {
			s += ":"
		}
		s += fmt.Sprintf("%d:%d", pos.Line, pos.Column)
	}
	if s == "" {
		s = "-"
	}
	return s
}

// Token is the set of lexical tokens in SQL
type Token int

const (
	// Special tokens
	INVALID Token = iota
	EOL
	COMMENT

	// Identifiers
	IDENT
	QUOTED_IDENT

	// Literals
	STRING
	NUMBER

	// Punctuation
	SEMICOLON
	COLON
	DOLLAR
	BANG
	EQUALS
	AT
	COMMA
	ASTERISK
	QUESTION
	SLASH
	PERCENT
	PLUS
	MINUS
	PERIOD

	// Delimiters
	LEFT_PAREN
	LEFT_BRACKET
	RIGHT_PAREN
	RIGHT_BRACKET

	// Keywords
	keywords_begin
	CREATE
	TABLE

	DROP

	SELECT
	FROM
	WHERE
	HAVING
	GROUP
	ORDER
	BY
	ASC
	DESC
	LIMIT
	OFFSET

	INSERT
	INTO
	VALUES

	UPDATE
	SET

	WITH
	AS
	ALL
	DISTINCT
	DISTINCTROW
	FILTER

	NULL
	TRUE
	FALSE

	AND
	OR
	IS
	NOT
	IN
	BETWEEN
	LIKE
	SIMILAR

	keywords_end
)

var tokens = [...]string{
	INVALID: "Invalid token",
	EOL:     "EOL",
	COMMENT: "Comment",

	IDENT:        "Identifier",
	QUOTED_IDENT: "Quoted identifier",

	STRING: "String",
	NUMBER: "Number",

	SEMICOLON: ";",
	COLON:     ":",
	DOLLAR:    "$",
	BANG:      "!",
	EQUALS:    "=",
	AT:        "@",
	COMMA:     ",",
	ASTERISK:  "*",
	QUESTION:  "?",
	SLASH:     "/",
	PERCENT:   "%",
	PLUS:      "+",
	MINUS:     "-",
	PERIOD:    ".",

	LEFT_PAREN:    "(",
	LEFT_BRACKET:  "[",
	RIGHT_PAREN:   ")",
	RIGHT_BRACKET: "]",

	CREATE: "CREATE",
	TABLE:  "TABLE",

	DROP: "DROP",

	SELECT: "SELECT",
	FROM:   "FROM",
	WHERE:  "WHERE",
	HAVING: "HAVING",
	GROUP:  "GROUP",
	ORDER:  "ORDER",
	BY:     "BY",
	ASC:    "ASC",
	DESC:   "DESC",
	LIMIT:  "LIMIT",
	OFFSET: "OFFSET",

	INSERT: "INSERT",
	INTO:   "INTO",
	VALUES: "VALUES",

	UPDATE: "UPDATE",
	SET:    "SET",

	WITH:        "WITH",
	AS:          "AS",
	ALL:         "ALL",
	DISTINCT:    "DISTINCT",
	DISTINCTROW: "DISTINCTROW",
	FILTER:      "FILTER",

	NULL:  "NULL",
	TRUE:  "TRUE",
	FALSE: "FALSE",

	AND:     "AND",
	OR:      "OR",
	IS:      "IS",
	NOT:     "NOT",
	IN:      "IN",
	BETWEEN: "BETWEEN",
	LIKE:    "LIKE",
	SIMILAR: "SIMILAR",
}

func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "Token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keywords_begin + 1; i < keywords_end; i++ {
		keywords[tokens[i]] = i
	}
}

// Lookup maps an identifier to its keyword token or IDENT (if not a keyword).
//
func Lookup(ident string) Token {
	if tok, is_keyword := keywords[strings.ToUpper(ident)]; is_keyword {
		return tok
	}
	return IDENT
}

func (tok Token) HasLiteral() bool {
	return COMMENT <= tok && tok <= NUMBER
}

func (tok Token) IsKeyword() bool {
	return keywords_begin < tok && tok < keywords_end
}
