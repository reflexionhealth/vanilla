package ast

import "strconv"

type OpType int

const (
	NOOP OpType = iota

	// Binary operators
	AND
	OR
	XOR
	IN
	IS
	LIKE
	ILIKE
	REGEXP
	BETWEEN
	OVERLAPS
	LESS
	LESS_OR_EQUAL
	GREATER
	GREATER_OR_EQUAL
	NOT_EQUAL
	EQUAL
	ADD
	SUBTRACT
	MULTIPLY
	DIVIDE
	MODULO
	SHIFT_LEFT
	SHIFT_RIGHT
	BIT_AND
	BIT_OR
	BIT_XOR

	// Unary operators
	NOT
	IS_NULL
	NOT_NULL
	NEGATE
	BIT_NOT
)

type OpPrecedence int8

const (
	MinPrecedence OpPrecedence = 0
	MaxPrecedence OpPrecedence = 127
)

type OpKind uint8

const (
	Nullary OpKind = iota
	Prefix
	Infix
)

var operatorKinds = [...]string{
	Nullary: "Nullary",
	Prefix:  "Prefix",
	Infix:   "Infix",
}

func (kind OpKind) String() string {
	s := ""
	if 0 <= kind && kind < OpKind(len(operatorKinds)) {
		s = operatorKinds[kind]
	}
	if s == "" {
		s = "OpKind(" + strconv.Itoa(int(kind)) + ")"
	}
	return s
}

type OpAssoc uint8

const (
	NonAssoc OpAssoc = iota
	LeftAssoc
	RightAssoc
)

var associatives = [...]string{
	NonAssoc:   "NonAssoc",
	LeftAssoc:  "LeftAssoc",
	RightAssoc: "RightAssoc",
}

func (asc OpAssoc) String() string {
	s := ""
	if 0 <= asc && asc < OpAssoc(len(associatives)) {
		s = associatives[asc]
	}
	if s == "" {
		s = "Assoc(" + strconv.Itoa(int(asc)) + ")"
	}
	return s
}

type Operator struct {
	Literal    string
	Type       OpType
	Kind       OpKind
	Assoc      OpAssoc
	Precedence OpPrecedence
}

type OperatorSet struct {
	Literals [3]map[string]Operator
}

func (os *OperatorSet) Init() {
	os.Literals[Infix] = make(map[string]Operator)
	os.Literals[Prefix] = make(map[string]Operator)
}

func (os *OperatorSet) Define(op Operator) {
	os.Literals[op.Kind][op.Literal] = op
}

func (os *OperatorSet) Lookup(lit string, kind OpKind) (Operator, bool) {
	operators, exists := os.Literals[kind][lit]
	return operators, exists
}
