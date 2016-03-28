package parser

import (
	. "github.com/reflexionhealth/vanilla/sql/language/ast"
	"github.com/reflexionhealth/vanilla/sql/language/scanner"
)

var AnsiRuleset = Ruleset{Operators: AnsiOperators}
var MysqlRuleset = Ruleset{
	CanSelectDistinctRow: true,

	Operators: MysqlOperators,
	ScanRules: scanner.Ruleset{
		BacktickIsQuotemark: true,
		DoubleQuoteIsString: true,
	},
}

// NOTE: The precedence values in the builtin operator sets may not be the same
// from version to version. If you define your own operators, copy instead of
// extending a builtin set.
const (
	UNARY   OpPrecedence = 80
	NUMERIC OpPrecedence = 60
	COMPARE OpPrecedence = 40
	LOGICAL OpPrecedence = 20
)

// AnsiOperators gives a set of the operators defined in the SQL standard
// all with left-associtivity and equal precedence except Assignment which
// has right-associativity and the lowest precedence.
//
// The precedence of operators between SQL implementations is very diverse,
// so when parsing complicated expressions, use the OperatorSet appropriate
// to the database being used.
var AnsiOperators = OperatorSet{
	Literals: [3]map[string]Operator{
		Prefix: {},
		Infix: {
			"=":  Operator{"=", EQUAL, Infix, LeftAssoc, COMPARE},
			"<>": Operator{"<>", NOT_EQUAL, Infix, LeftAssoc, COMPARE},
			">":  Operator{">", GREATER, Infix, LeftAssoc, COMPARE},
			"<":  Operator{"<", LESS, Infix, LeftAssoc, COMPARE},
			">=": Operator{">=", GREATER_OR_EQUAL, Infix, LeftAssoc, COMPARE},
			"<=": Operator{"<=", LESS_OR_EQUAL, Infix, LeftAssoc, COMPARE},

			"BETWEEN": Operator{"BETWEEN", BETWEEN, Infix, LeftAssoc, COMPARE},
			"LIKE":    Operator{"LIKE", LIKE, Infix, LeftAssoc, COMPARE},
			"IS":      Operator{"IS", IS, Infix, LeftAssoc, COMPARE},
			"IN":      Operator{"IN", IN, Infix, LeftAssoc, COMPARE},
		},
	},
}

// MysqlOperators gives the set of the operators defined by MySQL
var MysqlOperators = OperatorSet{
	Literals: [3]map[string]Operator{
		Prefix: {
			// unary operators
			"NOT": Operator{"NOT", NOT, Prefix, RightAssoc, LOGICAL + 6},
			"!":   Operator{"!", NOT, Prefix, RightAssoc, UNARY + 1},
			"-":   Operator{"-", NEGATE, Prefix, RightAssoc, UNARY},
			"~":   Operator{"~", BIT_NOT, Prefix, RightAssoc, UNARY},
		},
		Infix: {
			"^":   Operator{"^", BIT_XOR, Infix, LeftAssoc, NUMERIC + 10},
			"*":   Operator{"*", MULTIPLY, Infix, LeftAssoc, NUMERIC + 8},
			"/":   Operator{"/", DIVIDE, Infix, LeftAssoc, NUMERIC + 8},
			"%":   Operator{"%", MODULO, Infix, LeftAssoc, NUMERIC + 8},
			"DIV": Operator{"DIV", DIVIDE, Infix, LeftAssoc, NUMERIC + 8},
			"MOD": Operator{"MOD", MODULO, Infix, LeftAssoc, NUMERIC + 8},
			"+":   Operator{"+", ADD, Infix, LeftAssoc, NUMERIC + 6},
			"-":   Operator{"-", SUBTRACT, Infix, LeftAssoc, NUMERIC + 6},
			">>":  Operator{">>", SHIFT_RIGHT, Infix, LeftAssoc, NUMERIC + 4},
			"<<":  Operator{"<<", SHIFT_LEFT, Infix, LeftAssoc, NUMERIC + 4},
			"&":   Operator{"&", BIT_AND, Infix, LeftAssoc, NUMERIC + 2},
			"|":   Operator{"|", BIT_OR, Infix, LeftAssoc, NUMERIC},

			// symbolic comparisons
			"<=": Operator{"<=", LESS_OR_EQUAL, Infix, LeftAssoc, COMPARE},
			"!=": Operator{"!=", NOT_EQUAL, Infix, LeftAssoc, COMPARE},
			"<>": Operator{"<>", NOT_EQUAL, Infix, LeftAssoc, COMPARE},
			">":  Operator{">", GREATER, Infix, LeftAssoc, COMPARE},
			"<":  Operator{"<", LESS, Infix, LeftAssoc, COMPARE},
			">=": Operator{">=", GREATER_OR_EQUAL, Infix, LeftAssoc, COMPARE},
			"=":  Operator{"=", EQUAL, Infix, LeftAssoc, COMPARE},

			// keyword comparisons
			"IN":      Operator{"IN", IN, Infix, LeftAssoc, COMPARE},
			"IS":      Operator{"IS", IS, Infix, LeftAssoc, COMPARE},
			"LIKE":    Operator{"LIKE", LIKE, Infix, LeftAssoc, COMPARE},
			"REGEXP":  Operator{"REGEXP", REGEXP, Infix, LeftAssoc, COMPARE},
			"BETWEEN": Operator{"BETWEEN", BETWEEN, Infix, LeftAssoc, COMPARE - 2},

			// logical operators
			"AND": Operator{"AND", AND, Infix, LeftAssoc, LOGICAL + 4},
			"&&":  Operator{"&&", AND, Infix, LeftAssoc, LOGICAL + 4},
			"XOR": Operator{"XOR", XOR, Infix, LeftAssoc, LOGICAL + 2},
			"OR":  Operator{"OR", OR, Infix, LeftAssoc, LOGICAL},
			"||":  Operator{"||", OR, Infix, LeftAssoc, LOGICAL},
		},
	},
}
